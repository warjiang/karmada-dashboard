package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"os"
	"warjiang/karmada-dashboard/client"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	token := os.Getenv("token")
	kubeConfigPath := os.Getenv("kubeConfigPath")
	client.Init(
		client.WithUserAgent("dashboard-auth"),
		client.WithKubeconfig(kubeConfigPath),
		client.WithInsecureTLSSkipVerify(true),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/login", nil)
	client.SetAuthorizationHeader(req, token)
	karmadaClient, err := client.Client(req)
	if err != nil {
		panic(err)
	}
	version, err := karmadaClient.Discovery().ServerVersion()
	if err != nil {
		panic(err)
	}
	fmt.Println(version)
	fmt.Println(version.GoVersion)
	fmt.Println(version.GitVersion)
	fmt.Println(version.GitCommit)
	clusters, err := karmadaClient.ClusterV1alpha1().Clusters().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, cluster := range clusters.Items {
		fmt.Println(cluster.Name)
	}
}
