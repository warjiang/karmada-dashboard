package misc

import "C"
import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/karmada-io/dashboard/cmd/api/app/router"
	v1 "github.com/karmada-io/dashboard/cmd/api/app/types/api/v1"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/client"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/remotecommand"
	"time"
)

const (
	payloadKey = "kubernetes.io/serviceaccount/service-account.uid"
	namespace  = "karmada-system"
)

func handleCreateTerminal(c *gin.Context) {
	restConfig, err := client.GetKarmadaConfigFromRequest(c.Request)
	if err != nil {
		common.Fail(c, err)
		return
	}
	payload, err := client.ParseToken(restConfig.BearerToken)
	if err != nil {
		common.Fail(c, err)
		return
	}
	saUid, exist := payload[payloadKey]
	if !exist {
		common.Fail(c, errors.New("payload key not found"))
		return
	}

	terminalPod := client.GenerateTtydTerminal(saUid)
	// TODO: double check pod terminal pod exist
	kubeClient := client.InClusterClient()
	_, err = kubeClient.CoreV1().Pods(namespace).Create(context.TODO(), terminalPod, metav1.CreateOptions{})
	if err != nil {
		common.Fail(c, err)
		return
	}

	// wait for terminal pod ready
	deadlineCtx, _ := context.WithTimeout(context.TODO(), 60*time.Second)
	isPodReady := func(ctx context.Context) (bool, error) {
		pod, err := kubeClient.CoreV1().Pods(namespace).Get(ctx, terminalPod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		// 假设只有一个 Pod，检查该 Pod 是否处于 Ready 状态
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				return true, nil
			}
		}

		return false, nil
	}

	if err = wait.PollUntilContextCancel(deadlineCtx, 5*time.Second, false, isPodReady); err != nil {
		common.Fail(c, err)
		return
	}
	karmadaCfg, _, err := client.GetKarmadaConfig()
	if err != nil {
		common.Fail(c, err)
		return
	}
	kubeconfigOpts := client.KubeconfigOpts{
		Server:                   karmadaCfg.Host,
		InsecureSkipTLSVerify:    karmadaCfg.Insecure,
		CertificateAuthorityData: base64.StdEncoding.EncodeToString(karmadaCfg.CAData),
		Token:                    restConfig.BearerToken,
	}
	if err := client.WriteKubeconfig(namespace, terminalPod.Name, terminalPod.Spec.Containers[0].Name, kubeconfigOpts); err != nil {
		common.Fail(c, err)
		return
	}
	common.Success(c, "ok")
}

func handleExecShell(c *gin.Context) {
	sessionID, err := genTerminalSessionId()
	if err != nil {
		common.Fail(c, err)
		return
	}
	cfg, _, err := client.GetKubeConfig()
	if err != nil {
		common.Fail(c, err)
		return
	}
	terminalSessions.Set(sessionID, TerminalSession{
		id:       sessionID,
		bound:    make(chan error),
		sizeChan: make(chan remotecommand.TerminalSize),
	})
	go WaitForTerminal(client.InClusterClient(), cfg, c, sessionID)
	common.Success(c, v1.TerminalResponse{ID: sessionID})
}

func init() {
	r := router.V1()
	r.POST("/misc/terminal", handleCreateTerminal)
	r.GET("/misc/pod/:namespace/:pod/shell/:container", handleExecShell)
	r.Any("/misc/sockjs/*w", gin.WrapH(CreateAttachHandler("/api/v1/misc/sockjs")))
}
