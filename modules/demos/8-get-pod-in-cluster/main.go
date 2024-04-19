package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
	"net/url"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	kubeConfigPath := os.Getenv("kubeConfigPath")
	apiConfig, err := GenerateAPIConfigFromKubeconfigFile(kubeConfigPath)
	if err != nil {
		panic(err)
	}
	originApiServer := apiConfig.Clusters[apiConfig.CurrentContext].Server
	originApiServerURL, err := url.Parse(originApiServer)
	if err != nil {
		panic(err)
	}
	clusterName := "member3"
	byApiConifg := false
	var kubeClient *kubeclient.Clientset
	memberAPIServer := fmt.Sprintf("%s://%s/apis/cluster.karmada.io/v1alpha1/clusters/%s/proxy", originApiServerURL.Scheme, originApiServerURL.Host, clusterName)

	if byApiConifg {

		apiConfig.Clusters[apiConfig.CurrentContext].Server = memberAPIServer
		kubeClient, err = ToClientSet(apiConfig)
		if err != nil {
			panic(err)
		}
	} else {
		restConfig, err := LoadRestConfig(kubeConfigPath, "")
		if err != nil {
			panic(err)
		}
		restConfig.Host = memberAPIServer
		kubeClient, err = kubeclient.NewForConfig(restConfig)
		fmt.Printf("Host: %s, apiPath:%s", restConfig.Host, restConfig.APIPath)
	}

	nsList, err := kubeClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, item := range nsList.Items {
		fmt.Println(item.Name)
	}
}
func LoadRestConfig(kubeconfig string, context string) (*rest.Config, error) {
	loader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	loadedConfig, err := loader.Load()
	if err != nil {
		return nil, err
	}

	if context == "" {
		context = loadedConfig.CurrentContext
	}
	klog.Infof("Use context %v", context)

	d := clientcmd.NewNonInteractiveClientConfig(
		*loadedConfig,
		context,
		&clientcmd.ConfigOverrides{},
		loader,
	)
	// config, err := d.ClientConfig()
	//rawConfig, err := d.RawConfig()
	//rawConfig.Clusters[rawConfig.CurrentContext].Server = "https://192.168.10.2:5443/apis/cluster.karmada.io/v1alpha1/clusters/member3/proxy"
	return d.ClientConfig()
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

// ToClientSet converts a KubeConfig object to a client
func ToClientSet(config *clientcmdapi.Config) (*kubeclient.Clientset, error) {
	overrides := clientcmd.ConfigOverrides{Timeout: "10s"}
	clientConfig, err := clientcmd.NewDefaultClientConfig(*config, &overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create API client configuration from kubeconfig: %w", err)
	}

	client, err := kubeclient.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}
	return client, nil
}
