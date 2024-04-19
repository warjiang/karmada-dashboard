package cluster

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/karmada-io/karmada/pkg/apis/cluster/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"net/http"
	"time"
	"warjiang/karmada-dashboard/client"
	commonapi "warjiang/karmada-dashboard/dashboard-api/api/common"
	v1 "warjiang/karmada-dashboard/dashboard-api/api/v1"
	"warjiang/karmada-dashboard/dashboard-api/pkg/parser"
	"warjiang/karmada-dashboard/dashboard-api/pkg/resource/cluster"
	"warjiang/karmada-dashboard/dashboard-api/pkg/router"
)

func handleGetClusterList(c *gin.Context) {
	karmadaClient, err := client.KarmadaClient(c.Request)
	if err != nil {
		klog.ErrorS(err, "Could not read login request")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	dataSelect := parser.ParseDataSelectPathParameter(c)
	result, err := cluster.GetClusterList(karmadaClient, dataSelect)
	if err != nil {
		klog.ErrorS(err, "Could not read login request")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func handleGetClusterDetail(c *gin.Context) {
	karmadaClient, err := client.KarmadaClient(c.Request)
	if err != nil {
		klog.ErrorS(err, "Could not read login request")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	name := c.Param("name")
	karmadaClient = karmadaClient
	name = name
	return
}

func handlePostCluster(c *gin.Context) {
	clusterRequest := new(v1.PostClusterRequest)
	if err := c.ShouldBind(clusterRequest); err != nil {
		klog.ErrorS(err, "Could not read cluster request")
		commonapi.Fail(c, err)
		return
	}
	karmadaClient, err := client.KarmadaClient(c.Request)
	if err != nil {
		klog.ErrorS(err, "Could not init karmada client based request-token")
		commonapi.Fail(c, err)
		return
	}

	if clusterRequest.SyncMode == v1alpha1.Pull {
		memberClusterClient, err := KubeClientSetFromKubeConfig(clusterRequest.MemberClusterKubeConfig)
		if err != nil {
			klog.ErrorS(err, "Generate kubeclient from member_cluster_kubeconfig failed")
			commonapi.Fail(c, err)
			return
		}
		_, apiConfig, err := client.GetKarmadaConfig()
		if err != nil {
			klog.ErrorS(err, "Get api config failed")
			commonapi.Fail(c, err)
			return
		}

		opts := &RegisterClusterInPullOption{
			karmadaClient:          karmadaClient,
			karmadaAgentCfg:        apiConfig,
			memberClusterNamespace: clusterRequest.MemberClusterNamespace,
			memberClusterClient:    memberClusterClient,
			clusterName:            clusterRequest.MemberClusterName,
			memberClusterEndpoint:  clusterRequest.MemberClusterEndpoint,
		}
		if err = RegisterClusterInPullMode(opts); err != nil {
			klog.ErrorS(err, "RegisterClusterInPullMode failed")
			commonapi.Fail(c, err)
		} else {
			klog.Infof("RegisterClusterInPullMode success")
			commonapi.Success(c, "ok")
		}
	} else if clusterRequest.SyncMode == v1alpha1.Push {
		memberClusterRestConfig, err := KubeRestConfigFromKubeConfig(clusterRequest.MemberClusterKubeConfig)
		if err != nil {
			klog.ErrorS(err, "Generate rest config from member_cluster_kubeconfig failed")
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		restConfig, _, err := client.GetKarmadaConfig()
		if err != nil {
			klog.ErrorS(err, "Get rest config failed")
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		opts := &RegisterClusterInPushOption{
			karmadaClient:           karmadaClient,
			clusterName:             clusterRequest.MemberClusterName,
			controlPlaneRestConfig:  restConfig,
			memberClusterRestConfig: memberClusterRestConfig,
		}
		if err = RegisterClusterInPushMode(opts); err != nil {
			klog.ErrorS(err, "RegisterClusterInPushMode failed")
			commonapi.Fail(c, err)
		} else {
			klog.Infof("RegisterClusterInPullMode success")
			commonapi.Success(c, "ok")
		}
	}
}

func handlePutCluster(c *gin.Context) {

}

func handleDeleteCluster(c *gin.Context) {
	ctx := context.Context(c)
	clusterRequest := new(v1.DeleteClusterRequest)
	if err := c.ShouldBindUri(&clusterRequest); err != nil {
		commonapi.Fail(c, err)
		return
	}
	clusterName := clusterRequest.MemberClusterName
	karmadaClient, err := client.KarmadaClient(c.Request)
	if err != nil {
		klog.ErrorS(err, "Could not init karmada client based request-token")
		commonapi.Fail(c, err)
		return
	}
	waitDuration := time.Second * 60

	err = karmadaClient.ClusterV1alpha1().Clusters().Delete(ctx, clusterName, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		commonapi.Fail(c, fmt.Errorf("no cluster object %s found in karmada control Plane", clusterName))
		return
	}
	if err != nil {
		klog.Errorf("Failed to delete cluster object. cluster name: %s, error: %v", clusterName, err)
		commonapi.Fail(c, err)
		return
	}

	// make sure the given cluster object has been deleted
	err = wait.Poll(1*time.Second, waitDuration, func() (done bool, err error) {
		_, err = karmadaClient.ClusterV1alpha1().Clusters().Get(ctx, clusterName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			klog.Errorf("Failed to get cluster %s. err: %v", clusterName, err)
			return false, err
		}
		klog.Infof("Waiting for the cluster object %s to be deleted", clusterName)
		return false, nil
	})
	if err != nil {
		klog.Errorf("Failed to delete cluster object. cluster name: %s, error: %v", clusterName, err)
		commonapi.Fail(c, err)
		return
	}
	commonapi.Success(c, "ok")
	return
}

func init() {
	r := router.V1()
	r.GET("/cluster", handleGetClusterList)
	r.GET("/cluster/:member_cluster_name", handleGetClusterDetail)
	r.POST("/cluster", handlePostCluster)
	r.PUT("/cluster/:member_cluster_name", handlePutCluster)
	r.DELETE("/cluster/:member_cluster_name", handleDeleteCluster)
}
