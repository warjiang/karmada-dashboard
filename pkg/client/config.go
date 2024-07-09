package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
	"text/template"
)

var (
	kubeconfigTpl = `apiVersion: v1
kind: Config
clusters:
- name: karmada-apiserver
  cluster:
    server: {{ .Server }}
    {{- if .InsecureSkipTLSVerify }}
    insecure-skip-tls-verify: {{ .InsecureSkipTLSVerify }}
	{{- else}}
	certificate-authority-data: {{ .CertificateAuthorityData }}
	{{- end}}
users:
- name: karmada-apiserver
  user:
    token: {{ .Token }}
contexts:
- name: karmada-apiserver
  context:
    cluster: karmada-apiserver
    user: karmada-apiserver
current-context: karmada-apiserver`
	kubeconfigTemplate = template.Must(template.New("kubeconfigTemplate").Parse(kubeconfigTpl))
)

type KubeconfigOpts struct {
	Server                   string
	InsecureSkipTLSVerify    bool
	CertificateAuthorityData string
	Token                    string
}

func GenerateTtydTerminal(saUid string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("ttyd-%s", saUid),
			Namespace: "karmada-system",
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: "karmada-dashboard",
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "fronted-cn-beijing"},
			},
			Containers: []corev1.Container{
				{
					Name:            fmt.Sprintf("ttyd-%s", saUid),
					Image:           "fronted-cn-beijing.cr.volces.com/container/tsl0922/ttyd:1.7.4-dev-v1",
					ImagePullPolicy: corev1.PullIfNotPresent,
					Ports: []corev1.ContainerPort{
						{
							Name:          "tcp",
							ContainerPort: 7681,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					LivenessProbe: &corev1.Probe{
						FailureThreshold:    3,
						InitialDelaySeconds: 5,
						PeriodSeconds:       10,
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.FromInt32(7681),
							},
						},
					},
					ReadinessProbe: &corev1.Probe{
						FailureThreshold:    3,
						InitialDelaySeconds: 5,
						PeriodSeconds:       10,
						ProbeHandler: corev1.ProbeHandler{
							TCPSocket: &corev1.TCPSocketAction{
								Port: intstr.FromInt32(7681),
							},
						},
					},
				},
			},
		},
	}
}

func GenerateKubeconfigForTerminal(opts KubeconfigOpts) (string, error) {
	var tplOutput bytes.Buffer
	if err := kubeconfigTemplate.Execute(&tplOutput, opts); err != nil {
		return "", err
	} else {
		return tplOutput.String(), nil
	}
}

func WriteKubeconfig(namespace, podName, containerName string, opts KubeconfigOpts) error {
	if !isKubeInitialized() {
		return fmt.Errorf("kube client is not initialized")
	}
	tmpKubeConfig, err := GenerateKubeconfigForTerminal(opts)
	if err != nil {
		return errors.Wrap(err, "failed to generate kubeconfig")
	}
	script := fmt.Sprintf("mkdir -p ~/.kube && echo '%s' > ~/.kube/config", tmpKubeConfig)
	req := InClusterClient().CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   []string{"sh", "-c", script},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)
	executor, err := remotecommand.NewSPDYExecutor(kubernetesRestConfig, "POST", req.URL())
	if err != nil {
		return errors.Wrap(err, "failed to init executor")
	}
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    false,
	})
	if err != nil {
		return errors.Wrap(err, "failed to execute remotecmd")
	}
	return nil
}

func UnifyYaml(data string) (string, error) {
	tmpData := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(data), &tmpData)
	if err != nil {
		return "", err
	}
	tmpBuff, err := yaml.Marshal(tmpData)
	if err != nil {
		return "", err
	}
	return string(tmpBuff), nil
}

func ParseToken(token string) (map[string]string, error) {
	payload := make(map[string]string)
	items := strings.Split(token, ".")
	if len(items) != 3 {
		return nil, errors.New("invalid token")
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(items[1]); err != nil {
		return nil, errors.Wrap(err, "failed to decode token")
	} else {
		//return string(decoded), errors.New("invalid token")
		if err := json.Unmarshal(decoded, &payload); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal token")
		}
		return payload, nil
	}
}
