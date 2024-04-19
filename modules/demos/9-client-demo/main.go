package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"warjiang/karmada-dashboard/client"
)

func main() {
	// --karmada-kubeconfig=/Users/dingwenjiang/.kube/karmada.config --skip-karmada-apiserver-tls-verify=true
	client.InitKarmadaConfig(
		client.WithUserAgent("agent"),
		client.WithKubeconfig("/Users/dingwenjiang/.kube/karmada.config"),
		client.WithInsecureTLSSkipVerify(true),
	)
	karmadaClient := client.InClusterKarmadaClient()
	nsList, err := karmadaClient.ClusterV1alpha1().Clusters().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, item := range nsList.Items {
		fmt.Println(item.Name)
	}
}
