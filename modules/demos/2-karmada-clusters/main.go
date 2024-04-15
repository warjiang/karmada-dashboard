package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
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

	clusters, err := karmadaClient.ClusterV1alpha1().Clusters().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, cluster := range clusters.Items {
		fmt.Println("=================")
		fmt.Println(cluster.Name)
		fmt.Println(cluster.Status.KubernetesVersion)
		fmt.Println(cluster.Status.NodeSummary.TotalNum, cluster.Status.NodeSummary.ReadyNum)
		fmt.Printf(
			"%+v=>%+v\n",
			cluster.Status.ResourceSummary.Allocatable.Name(v1.ResourceCPU, resource.DecimalSI),
			cluster.Status.ResourceSummary.Allocated.Name(v1.ResourceCPU, resource.DecimalSI),
		)
		fmt.Printf(
			"%+v=>%+v\n",
			cluster.Status.ResourceSummary.Allocatable.Name(v1.ResourceMemory, resource.DecimalSI),
			cluster.Status.ResourceSummary.Allocated.Name(v1.ResourceMemory, resource.DecimalSI),
		)
		fmt.Printf(
			"%+v=>%+v\n",
			cluster.Status.ResourceSummary.Allocatable.Name(v1.ResourcePods, resource.DecimalSI),
			cluster.Status.ResourceSummary.Allocated.Name(v1.ResourcePods, resource.DecimalSI),
		)
		fmt.Printf("mode:%+v\n", cluster.Spec.SyncMode)
		//fmt.Println(cluster.Status.ResourceSummary.Allocatable.Name(v1.ResourceMemory, resource.DecimalSI))
		//fmt.Println(cluster.Status.ResourceSummary.Allocatable.Name(v1.ResourcePods, resource.DecimalSI))
		//fmt.Println(cluster.Status.ResourceSummary.Allocated.Name(v1.ResourceCPU, resource.DecimalSI))
		//fmt.Println(cluster.Status.ResourceSummary.Allocated.Name(v1.ResourceMemory, resource.DecimalSI))
		//fmt.Println(cluster.Status.ResourceSummary.Allocated.Name(v1.ResourcePods, resource.DecimalSI))

		fmt.Println("=================")
	}
}
