package main

import (
	"fmt"
	"github.com/joho/godotenv"
	clusterv1alpha1 "github.com/karmada-io/karmada/pkg/apis/cluster/v1alpha1"
	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	"github.com/karmada-io/karmada/pkg/karmadactl/util/apiclient"
	"github.com/karmada-io/karmada/pkg/util"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
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
	controlPlaneRestConfig, err := apiclient.RestConfig("karmada-apiserver", kubeConfigPath)
	if err != nil {
		klog.Fatalf("Failed to get cluster-api management cluster rest config. kubeconfig: %s, err: %v", kubeConfigPath, err)
	}
	controlPlaneKubeClient := kubeclient.NewForConfigOrDie(controlPlaneRestConfig)

	client.InitKarmada(
		client.WithKarmadaUserAgent("dashboard-auth"),
		client.WithKarmadaKubeconfig(kubeConfigPath),
		client.WithKarmadaContext("karmada-apiserver"),
		client.WithKarmadaInsecureTLSSkipVerify(true),
	)
	karmadaClient := client.InClusterKarmadaClient()

	kubeconfigPath := "/Users/dingwenjiang/.kube/1-60-config"
	clusterRestConfig, err := apiclient.RestConfig("", kubeconfigPath)
	if err != nil {
		klog.Fatalf("Failed to get cluster-api management cluster rest config. kubeconfig: %s, err: %v", kubeconfigPath, err)
	}
	clusterName := "test-cluster"
	clusterKubeClient := kubeclient.NewForConfigOrDie(clusterRestConfig)
	registerOption := util.ClusterRegisterOption{
		ClusterNamespace:   "karmada-cluster",
		ClusterName:        clusterName,
		ReportSecrets:      []string{util.KubeCredentials, util.KubeImpersonator},
		ControlPlaneConfig: controlPlaneRestConfig,
		ClusterConfig:      clusterRestConfig,
	}

	id, err := util.ObtainClusterID(clusterKubeClient)
	if err != nil {
		panic(err)
	}

	ok, name, err := util.IsClusterIdentifyUnique(karmadaClient, id)
	if err != nil {
		panic(err)
	}
	if !ok {
		panic(fmt.Errorf("the same cluster has been registered with name %s", name))
	}
	registerOption.ClusterID = id

	clusterSecret, impersonatorSecret, err := util.ObtainCredentialsFromMemberCluster(clusterKubeClient, registerOption)
	if err != nil {
		panic(err)
	}
	registerOption.Secret = *clusterSecret
	registerOption.ImpersonatorSecret = *impersonatorSecret
	err = util.RegisterClusterInControllerPlane(registerOption, controlPlaneKubeClient, generateClusterInControllerPlane)
	if err != nil {
		panic(err)
	}

	fmt.Printf("cluster(%s) is joined successfully\n", clusterName)
	/*
		clusterInfo := &v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test-cluster",
				Labels:      map[string]string{},
				Annotations: map[string]string{},
			},
			Spec: v1alpha1.ClusterSpec{
				SyncMode: v1alpha1.Push,
				Taints: []v1.Taint{
					{
						Key:    "key",
						Value:  "value",
						Effect: v1.TaintEffectNoSchedule,
					},
				},
			},
		}

		createResp, err := karmadaClient.ClusterV1alpha1().Clusters().Create(context.TODO(), clusterInfo, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
		fmt.Printf("create cluster resp %+v", createResp)
	*/

	// push -> karmadactl join
	// pull -> karmadactl register
	// 移除统一使用 unregister
}

func generateClusterInControllerPlane(opts util.ClusterRegisterOption) (*clusterv1alpha1.Cluster, error) {
	clusterObj := &clusterv1alpha1.Cluster{}
	clusterObj.Name = opts.ClusterName
	clusterObj.Spec.SyncMode = clusterv1alpha1.Push
	clusterObj.Spec.APIEndpoint = opts.ClusterConfig.Host
	clusterObj.Spec.ID = opts.ClusterID
	clusterObj.Spec.SecretRef = &clusterv1alpha1.LocalSecretReference{
		Namespace: opts.Secret.Namespace,
		Name:      opts.Secret.Name,
	}
	clusterObj.Spec.ImpersonatorSecretRef = &clusterv1alpha1.LocalSecretReference{
		Namespace: opts.ImpersonatorSecret.Namespace,
		Name:      opts.ImpersonatorSecret.Name,
	}

	if opts.ClusterProvider != "" {
		clusterObj.Spec.Provider = opts.ClusterProvider
	}

	if opts.ClusterZone != "" {
		clusterObj.Spec.Zone = opts.ClusterZone
	}

	if len(opts.ClusterZones) > 0 {
		clusterObj.Spec.Zones = opts.ClusterZones
	}

	if opts.ClusterRegion != "" {
		clusterObj.Spec.Region = opts.ClusterRegion
	}

	clusterObj.Spec.InsecureSkipTLSVerification = opts.ClusterConfig.TLSClientConfig.Insecure

	if opts.ClusterConfig.Proxy != nil {
		url, err := opts.ClusterConfig.Proxy(nil)
		if err != nil {
			return nil, fmt.Errorf("clusterConfig.Proxy error, %v", err)
		}
		clusterObj.Spec.ProxyURL = url.String()
	}

	controlPlaneKarmadaClient := karmadaclientset.NewForConfigOrDie(opts.ControlPlaneConfig)
	cluster, err := util.CreateClusterObject(controlPlaneKarmadaClient, clusterObj)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster(%s) object. error: %v", opts.ClusterName, err)
	}

	return cluster, nil
}
