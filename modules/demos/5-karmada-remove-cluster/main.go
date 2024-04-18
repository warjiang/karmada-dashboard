package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"os"
	"time"
	"warjiang/karmada-dashboard/client"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	// karmada clientset cannot use protobuf why?
	kubeConfigPath := os.Getenv("kubeConfigPath")

	client.InitKarmada(
		client.WithKarmadaUserAgent("dashboard-auth"),
		client.WithKarmadaKubeconfig(kubeConfigPath),
		client.WithKarmadaContext("karmada-apiserver"),
		client.WithKarmadaInsecureTLSSkipVerify(true),
	)
	karmadaClient := client.InClusterKarmadaClient()
	clusterName := "test-cluster"
	waitDuration := time.Second * 60

	err = karmadaClient.ClusterV1alpha1().Clusters().Delete(context.TODO(), clusterName, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		panic(fmt.Errorf("no cluster object %s found in karmada control Plane", clusterName))
	}
	if err != nil {
		klog.Errorf("Failed to delete cluster object. cluster name: %s, error: %v", clusterName, err)
		panic(err)
	}

	// make sure the given cluster object has been deleted
	err = wait.Poll(1*time.Second, waitDuration, func() (done bool, err error) {
		_, err = karmadaClient.ClusterV1alpha1().Clusters().Get(context.TODO(), clusterName, metav1.GetOptions{})
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
		panic(err)
	}
	// kind 环境测试大概用了 21 second 左右可以删除掉
	klog.Infof("Cluster %s delete successfully", clusterName)

	// push -> karmadactl join
	// pull -> karmadactl register
	// 移除统一使用 unregister
}
