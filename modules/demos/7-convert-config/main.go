package main

import (
	"github.com/joho/godotenv"
	"github.com/karmada-io/karmada/pkg/karmadactl/util/apiclient"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	kubeConfigPath := os.Getenv("kubeConfigPath")
	controlPlaneRestConfig, err := apiclient.RestConfig("karmada-apiserver", kubeConfigPath)
	if err != nil {
		panic(err)
	}
	controlPlaneRestConfig = controlPlaneRestConfig
}

func ConvertRestConfigToAPIConfig(restConfig *rest.Config) *clientcmdapi.Config {
	apiConfig := &clientcmdapi.Config{
		Clusters:       make(map[string]*clientcmdapi.Cluster),
		AuthInfos:      make(map[string]*clientcmdapi.AuthInfo),
		Contexts:       make(map[string]*clientcmdapi.Context),
		CurrentContext: "",
	}
	//restConfig.Username

	// 设置集群信息
	apiConfig.Clusters["cluster"] = &clientcmdapi.Cluster{
		Server:                restConfig.Host,
		InsecureSkipTLSVerify: restConfig.Insecure,
	}

	// 设置认证信息
	apiConfig.AuthInfos["authInfo"] = &clientcmdapi.AuthInfo{
		Token: restConfig.BearerToken,
	}

	// 设置上下文信息
	apiConfig.Contexts["context"] = &clientcmdapi.Context{
		Cluster:  "cluster",
		AuthInfo: "authInfo",
	}

	apiConfig.CurrentContext = "context"

	return apiConfig
}

func GenerateAPIConfigFromKubeconfigFile(kubeconfigPath string) (*clientcmdapi.Config, error) {
	// 使用 clientcmd 从 kubeconfig 文件加载配置
	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	currentContext := config.CurrentContext
	context := config.Contexts[currentContext]
	clusterName := context.Cluster
	authInfoName := context.AuthInfo
	cluster := config.Clusters[clusterName]
	authInfo := config.AuthInfos[authInfoName]
	apiConfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName: cluster,
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			authInfoName: authInfo,
		},
		Contexts: map[string]*clientcmdapi.Context{
			currentContext: {
				Cluster:  clusterName,
				AuthInfo: authInfoName,
			},
		},
		CurrentContext: currentContext,
	}
	return apiConfig, nil
}
