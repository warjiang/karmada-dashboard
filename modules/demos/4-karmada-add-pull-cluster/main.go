package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	cmdutil "github.com/karmada-io/karmada/pkg/karmadactl/util"
	"github.com/karmada-io/karmada/pkg/karmadactl/util/apiclient"
	"github.com/karmada-io/karmada/pkg/util"
	karmadautil "github.com/karmada-io/karmada/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
	"os"
	"strings"
	"time"
)

const (
	// KarmadaKubeconfigName is the name of karmada kubeconfig
	KarmadaKubeconfigName = "karmada-kubeconfig"
	// KarmadaAgentServiceAccountName is the name of karmada-agent serviceaccount
	KarmadaAgentServiceAccountName = "karmada-agent-sa"
	// KarmadaAgentName is the name of karmada-agent
	KarmadaAgentName = "karmada-agent"
)

var karmadaAgentLabels = map[string]string{"app": KarmadaAgentName}

var (
	kubeConfigPath         string // path-to-karmada.config
	controlPlaneRestConfig *rest.Config
	controlPlaneApiConfig  *clientcmdapi.Config
	karmadaClient          *karmadaclientset.Clientset
	memberKubeConfigPath   string // path-to-member.config
	memberClusterClient    *kubeclient.Clientset
	memberClusterEndpoint  = "https://192.168.10.7:6443"
	namespace              = "example" // karmada-agent will be installed under the scope of the namespace
	clusterName            = "test-cluster"
	karmadaAgentImage      = "fronted-cn-beijing.cr.volces.com/container/karmada/karmada-agent:latest"
	karmadaAgentReplicas   = int32(2)
	clusterNamespace       = "karmada-cluster" // namespace in control-plane for recording member cluster
	clusterProvider        = ""
	clusterRegion          = ""
	clusterZones           []string
	proxyServerAddress     = ""
	timeout                = 5 * time.Minute
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	kubeConfigPath = os.Getenv("kubeConfigPath")
	controlPlaneRestConfig, err = apiclient.RestConfig("karmada-apiserver", kubeConfigPath)
	//controlPlaneApiConfig = ConvertRestConfigToAPIConfig(controlPlaneRestConfig)
	if err != nil {
		klog.Fatalf("Failed to get cluster-api management cluster rest config. kubeconfig: %s, err: %v", kubeConfigPath, err)
		os.Exit(1)
	}
	controlPlaneApiConfig, err = GenerateAPIConfigFromKubeconfigFile(kubeConfigPath)
	if err != nil {
		klog.Fatalf("Failed to get cluster-api management cluster rest config. kubeconfig: %s, err: %v", kubeConfigPath, err)
		os.Exit(1)
	}
	karmadaClient, err = karmadaclientset.NewForConfig(controlPlaneRestConfig)
	if err != nil {
		klog.ErrorS(err, "Could not init kubernetes in-cluster client")
		os.Exit(1)
	}
	memberKubeConfigPath = os.Getenv("memberKubeConfigPath")
	memberClusterClient, err = ClientSetFromFile(memberKubeConfigPath)
	if err != nil {
		klog.ErrorS(err, "Could not init kubernetes in-cluster client")
		os.Exit(1)
	}

	klog.V(1).Infof("Registering cluster. cluster name: %s", clusterName)
	klog.V(1).Infof("Registering cluster. cluster namespace: %s", clusterNamespace)

	fmt.Println("[preflight] Running pre-flight checks")
	errlist := preflight()
	if len(errlist) > 0 {
		fmt.Println("error execution phase preflight: [preflight] Some fatal errors occurred:")
		for _, err := range errlist {
			fmt.Printf("\t[ERROR]: %s\n", err)
		}

		fmt.Printf("\n[preflight] Please check the above errors\n")
		return
	}
	fmt.Println("[preflight] All pre-flight checks were passed")

	fmt.Println("[karmada-agent-start] Waiting to check cluster exists")
	_, exist, err := karmadautil.GetClusterWithKarmadaClient(karmadaClient, clusterName)
	if err != nil {
		panic(err)
	} else if exist {
		//return fmt.Errorf("failed to register as cluster with name %s already exists", o.ClusterName)
		panic(fmt.Errorf("failed to register as cluster with name %s already exists", clusterName))
	}
	// It's necessary to set the label of namespace to make sure that the namespace is created by Karmada.
	labels := map[string]string{
		util.ManagedByKarmadaLabel: util.ManagedByKarmadaLabelValue,
	}
	// ensure namespace where the karmada-agent resources be deployed exists in the member cluster
	if _, err := karmadautil.EnsureNamespaceExistWithLabels(memberClusterClient, namespace, false, labels); err != nil {
		panic(err)
	}

	// create the necessary secret and RBAC in the member cluster
	fmt.Println("[karmada-agent-start] Waiting the necessary secret and RBAC")
	if err := createSecretAndRBACInMemberCluster(controlPlaneApiConfig); err != nil {
		panic(err)
	}

	// create karmada-agent Deployment in the member cluster
	fmt.Println("[karmada-agent-start] Waiting karmada-agent Deployment")
	KarmadaAgentDeployment := makeKarmadaAgentDeployment()

	if _, err := memberClusterClient.AppsV1().Deployments(namespace).Create(context.TODO(), KarmadaAgentDeployment, metav1.CreateOptions{}); err != nil {
		panic(err)
	}

	if err := cmdutil.WaitForDeploymentRollout(memberClusterClient, KarmadaAgentDeployment, int(timeout)); err != nil {
		panic(err)
	}

	fmt.Printf("\ncluster(%s) is joined successfully\n", clusterName)

}

// preflight checks the deployment environment of the member cluster
func preflight() []error {
	var errlist []error

	// check if relative resources already exist in member cluster
	_, err := memberClusterClient.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err == nil {
		_, err = memberClusterClient.CoreV1().Secrets(namespace).Get(context.TODO(), KarmadaKubeconfigName, metav1.GetOptions{})
		if err == nil {
			errlist = append(errlist, fmt.Errorf("%s/%s Secret already exists", namespace, KarmadaKubeconfigName))
		} else if !apierrors.IsNotFound(err) {
			errlist = append(errlist, err)
		}

		_, err = memberClusterClient.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), KarmadaAgentServiceAccountName, metav1.GetOptions{})
		if err == nil {
			errlist = append(errlist, fmt.Errorf("%s/%s ServiceAccount already exists", namespace, KarmadaAgentServiceAccountName))
		} else if !apierrors.IsNotFound(err) {
			errlist = append(errlist, err)
		}

		_, err = memberClusterClient.AppsV1().Deployments(namespace).Get(context.TODO(), KarmadaAgentName, metav1.GetOptions{})
		if err == nil {
			errlist = append(errlist, fmt.Errorf("%s/%s Deployment already exists", namespace, KarmadaAgentName))
		} else if !apierrors.IsNotFound(err) {
			errlist = append(errlist, err)
		}
	}

	return errlist
}

// createSecretAndRBACInMemberCluster create required secrets and rbac in member cluster
func createSecretAndRBACInMemberCluster(karmadaAgentCfg *clientcmdapi.Config) error {
	configBytes, err := clientcmd.Write(*karmadaAgentCfg)
	if err != nil {
		return fmt.Errorf("failure while serializing karmada-agent kubeConfig. %w", err)
	}

	kubeConfigSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KarmadaKubeconfigName,
			Namespace: namespace,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{KarmadaKubeconfigName: string(configBytes)},
	}

	// create karmada-kubeconfig secret to be used by karmada-agent component.
	if err := cmdutil.CreateOrUpdateSecret(memberClusterClient, kubeConfigSecret); err != nil {
		return fmt.Errorf("create secret %s failed: %v", kubeConfigSecret.Name, err)
	}

	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: KarmadaAgentName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			{
				NonResourceURLs: []string{"*"},
				Verbs:           []string{"get"},
			},
		},
	}

	// create a karmada-agent ClusterRole in member cluster.
	if err := cmdutil.CreateOrUpdateClusterRole(memberClusterClient, clusterRole); err != nil {
		return err
	}

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KarmadaAgentServiceAccountName,
			Namespace: namespace,
		},
	}

	// create service account for karmada-agent
	_, err = karmadautil.EnsureServiceAccountExist(memberClusterClient, sa, false)
	if err != nil {
		return err
	}

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: KarmadaAgentName,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sa.Name,
				Namespace: sa.Namespace,
			},
		},
	}

	// grant karmada-agent clusterrole to karmada-agent service account
	if err := cmdutil.CreateOrUpdateClusterRoleBinding(memberClusterClient, clusterRoleBinding); err != nil {
		return err
	}

	return nil
}

// makeKarmadaAgentDeployment generate karmada-agent Deployment
func makeKarmadaAgentDeployment() *appsv1.Deployment {
	karmadaAgent := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KarmadaAgentName,
			Namespace: namespace,
			Labels:    karmadaAgentLabels,
		},
	}

	controllers := []string{"*"}

	podSpec := corev1.PodSpec{
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: "fronted-cn-beijing",
			},
		},
		ServiceAccountName: KarmadaAgentServiceAccountName,
		Containers: []corev1.Container{
			{
				Name:  KarmadaAgentName,
				Image: karmadaAgentImage,
				Command: []string{
					"/bin/karmada-agent",
					"--karmada-kubeconfig=/etc/kubeconfig/karmada-kubeconfig",
					fmt.Sprintf("--cluster-name=%s", clusterName),
					fmt.Sprintf("--cluster-api-endpoint=%s", memberClusterEndpoint),
					fmt.Sprintf("--cluster-provider=%s", clusterProvider),
					fmt.Sprintf("--cluster-region=%s", clusterRegion),
					fmt.Sprintf("--cluster-zones=%s", strings.Join(clusterZones, ",")),
					fmt.Sprintf("--controllers=%s", strings.Join(controllers, ",")),
					fmt.Sprintf("--proxy-server-address=%s", proxyServerAddress),
					fmt.Sprintf("--leader-elect-resource-namespace=%s", namespace),
					"--cluster-status-update-frequency=10s",
					"--bind-address=0.0.0.0",
					"--secure-port=10357",
					"--v=4",
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "kubeconfig",
						MountPath: "/etc/kubeconfig",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "kubeconfig",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: KarmadaKubeconfigName,
					},
				},
			},
		},
		Tolerations: []corev1.Toleration{
			{
				Key:      "node-role.kubernetes.io/master",
				Operator: corev1.TolerationOpExists,
			},
		},
	}
	// PodTemplateSpec
	podTemplateSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      KarmadaAgentName,
			Namespace: namespace,
			Labels:    karmadaAgentLabels,
		},
		Spec: podSpec,
	}
	// DeploymentSpec
	karmadaAgent.Spec = appsv1.DeploymentSpec{
		Replicas: &karmadaAgentReplicas,
		Template: podTemplateSpec,
		Selector: &metav1.LabelSelector{
			MatchLabels: karmadaAgentLabels,
		},
	}

	return karmadaAgent
}

// ClientSetFromFile returns a ready-to-use client from a kubeconfig file
func ClientSetFromFile(path string) (*kubeclient.Clientset, error) {
	config, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load admin kubeconfig: %w", err)
	}
	return ToClientSet(config)
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

// ToKarmadaClient converts a KubeConfig object to a client
func ToKarmadaClient(config *clientcmdapi.Config) (*karmadaclientset.Clientset, error) {
	overrides := clientcmd.ConfigOverrides{Timeout: "10s"}
	clientConfig, err := clientcmd.NewDefaultClientConfig(*config, &overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create API client configuration from kubeconfig: %w", err)
	}

	karmadaClient, err := karmadaclientset.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return karmadaClient, nil
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

/*
  # Register cluster into karmada control plane with Pull mode.
  # If '--cluster-name' isn't specified, the cluster of current-context will be used by default.
  karmadactl register [karmada-apiserver-endpoint] --cluster-name=<CLUSTER_NAME> --token=<TOKEN>  --discovery-token-ca-cert-hash=<CA-CERT-HASH>

  # UnsafeSkipCAVerification allows token-based discovery without CA verification via CACertHashes. This can weaken
  # the security of register command since other clusters can impersonate the control-plane.
  karmadactl register [karmada-apiserver-endpoint] --token=<TOKEN>  --discovery-token-unsafe-skip-ca-verification=true
*/
/*
// BootstrapToken is used to set the options for bootstrap token based discovery
BootstrapToken *BootstrapTokenDiscovery
// BootstrapTokenDiscovery is used to set the options for bootstrap token based discovery
type BootstrapTokenDiscovery struct {
	// Token is a token used to validate cluster information
	// fetched from the control-plane.
	Token string

	// APIServerEndpoint is an IP or domain name to the API server from which info will be fetched.
	APIServerEndpoint string

	// CACertHashes specifies a set of public key pins to verify
	// when token-based discovery is used. The root CA found during discovery
	// must match one of these values. Specifying an empty set disables root CA
	// pinning, which can be unsafe. Each hash is specified as "<type>:<value>",
	// where the only currently supported type is "sha256". This is a hex-encoded
	// SHA-256 hash of the Subject Public Key Info (SPKI) object in DER-encoded
	// ASN.1. These hashes can be calculated using, for example, OpenSSL.
	CACertHashes []string

	// UnsafeSkipCAVerification allows token-based discovery
	// without CA verification via CACertHashes. This can weaken
	// the security of register command since other clusters can impersonate the control-plane.
	UnsafeSkipCAVerification bool
}

// CreateBasic creates a basic, general KubeConfig object that then can be extended
func CreateBasic(serverURL, clusterName, userName string, caCert []byte) *clientcmdapi.Config {
	// Use the cluster and the username as the context name
	contextName := fmt.Sprintf("%s@%s", userName, clusterName)

	return &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName: {
				Server:                   serverURL,
				CertificateAuthorityData: caCert,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			contextName: {
				Cluster:  clusterName,
				AuthInfo: userName,
			},
		},
		AuthInfos:      map[string]*clientcmdapi.AuthInfo{},
		CurrentContext: contextName,
	}
}

// CreateWithToken creates a KubeConfig object with access to the API server with a token
func CreateWithToken(serverURL, clusterName, userName string, caCert []byte, token string) *clientcmdapi.Config {
	config := CreateBasic(serverURL, clusterName, userName, caCert)
	config.AuthInfos[userName] = &clientcmdapi.AuthInfo{
		Token: token,
	}
	return config
}

// getClusterInfoFromControlPlane creates a client from the given kubeconfig if the given client is nil,
// and requests the cluster info ConfigMap using PollImmediate.
func getClusterInfoFromControlPlane(client kubeclient.Interface, kubeconfig *clientcmdapi.Config, token *tokenutil.Token, interval, duration time.Duration) (*corev1.ConfigMap, error) {
	var cm *corev1.ConfigMap
	var err error

	// Create client from kubeconfig
	if client == nil {
		client, err = ToClientSet(kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithTimeout(context.TODO(), duration)
	defer cancel()

	wait.JitterUntil(func() {
		cm, err = client.CoreV1().ConfigMaps(metav1.NamespacePublic).Get(context.TODO(), bootstrapapi.ConfigMapClusterInfo, metav1.GetOptions{})
		if err != nil {
			klog.V(1).Infof("[discovery] Failed to request cluster-info, will try again: %v", err)
			return
		}
		// Even if the ConfigMap is available the JWS signature is patched-in a bit later.
		// Make sure we retry util then.
		if _, ok := cm.Data[bootstrapapi.JWSSignatureKeyPrefix+token.ID]; !ok {
			klog.V(1).Infof("[discovery] The cluster-info ConfigMap does not yet contain a JWS signature for token ID %q, will try again", token.ID)
			err = fmt.Errorf("could not find a JWS signature in the cluster-info ConfigMap for token ID %q", token.ID)
			return
		}
		// Cancel the context on success
		cancel()
	}, interval, 0.3, true, ctx.Done())

	if err != nil {
		return nil, err
	}

	return cm, nil
}
// appendError append err to errlist
func appendError(errlist []error, err error) []error {
	if err == nil {
		return errlist
	}
	errlist = append(errlist, err)
	return errlist
}

// validateClusterInfoToken validates that the JWS token present in the cluster info ConfigMap is valid
func validateClusterInfoToken(insecureClusterInfo *corev1.ConfigMap, token *tokenutil.Token, parentCommand string) ([]byte, error) {
	insecureKubeconfigString, ok := insecureClusterInfo.Data[bootstrapapi.KubeConfigKey]
	if !ok || len(insecureKubeconfigString) == 0 {
		return nil, fmt.Errorf("there is no %s key in the %s ConfigMap. This API Server isn't set up for token bootstrapping, can't connect",
			bootstrapapi.KubeConfigKey, bootstrapapi.ConfigMapClusterInfo)
	}

	detachedJWSToken, ok := insecureClusterInfo.Data[bootstrapapi.JWSSignatureKeyPrefix+token.ID]
	if !ok || len(detachedJWSToken) == 0 {
		return nil, fmt.Errorf("token id %q is invalid for this cluster or it has expired. Use \"%s token create\" on the karmada-control-plane to create a new valid token", token.ID, parentCommand)
	}

	if !bootstrap.DetachedTokenIsValid(detachedJWSToken, insecureKubeconfigString, token.ID, token.Secret) {
		return nil, fmt.Errorf("failed to verify JWS signature of received cluster info object, can't trust this API Server")
	}

	return []byte(insecureKubeconfigString), nil
}

// validateClusterCA validates the cluster CA found in the insecure kubeconfig
func validateClusterCA(insecureConfig *clientcmdapi.Config, pubKeyPins *pubkeypin.Set) ([]byte, error) {
	var clusterCABytes []byte
	for _, cluster := range insecureConfig.Clusters {
		clusterCABytes = cluster.CertificateAuthorityData
	}

	clusterCAs, err := certutil.ParseCertsPEM(clusterCABytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cluster CA from the %s ConfigMap: %w", bootstrapapi.ConfigMapClusterInfo, err)
	}

	// Validate the cluster CA public key against the pinned set
	err = pubKeyPins.CheckAny(clusterCAs)
	if err != nil {
		return nil, fmt.Errorf("cluster CA found in %s ConfigMap is invalid: %w", bootstrapapi.ConfigMapClusterInfo, err)
	}

	return clusterCABytes, nil
}

// buildSecureBootstrapKubeConfig makes a kubeconfig object that connects securely to the API Server for bootstrapping purposes (validating with the specified CA)
func buildSecureBootstrapKubeConfig(endpoint string, caCert []byte, clustername string) *clientcmdapi.Config {
	controlPlaneEndpoint := fmt.Sprintf("https://%s", endpoint)
	bootstrapConfig := CreateBasic(controlPlaneEndpoint, clustername, BootstrapUserName, caCert)
	return bootstrapConfig
}

// retrieveValidatedConfigInfo is a private implementation of RetrieveValidatedConfigInfo.
func retrieveValidatedConfigInfo(client kubeclient.Interface, bootstrapTokenDiscovery *BootstrapTokenDiscovery, duration, interval time.Duration, parentCommand string) (*clientcmdapi.Config, error) {
	token, err := tokenutil.NewToken(bootstrapTokenDiscovery.Token)
	if err != nil {
		return nil, err
	}

	// Load the CACertHashes into a pubkeypin.Set
	pubKeyPins := pubkeypin.NewSet()
	if err = pubKeyPins.Allow(bootstrapTokenDiscovery.CACertHashes...); err != nil {
		return nil, fmt.Errorf("invalid discovery token CA certificate hash: %v", err)
	}

	// Make sure the interval is not bigger than the duration
	if interval > duration {
		interval = duration
	}

	endpoint := bootstrapTokenDiscovery.APIServerEndpoint
	insecureBootstrapConfig := buildInsecureBootstrapKubeConfig(endpoint, DefaultClusterName)
	clusterName := insecureBootstrapConfig.Contexts[insecureBootstrapConfig.CurrentContext].Cluster

	klog.V(1).Infof("[discovery] Created cluster-info discovery client, requesting info from %q", endpoint)
	insecureClusterInfo, err := getClusterInfoFromControlPlane(client, insecureBootstrapConfig, token, interval, duration)
	if err != nil {
		return nil, err
	}

	// Validate the token in the cluster info
	insecureKubeconfigBytes, err := validateClusterInfoToken(insecureClusterInfo, token, parentCommand)
	if err != nil {
		return nil, err
	}

	// Load the insecure config
	insecureConfig, err := clientcmd.Load(insecureKubeconfigBytes)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse the kubeconfig file in the %s ConfigMap: %w", bootstrapapi.ConfigMapClusterInfo, err)
	}

	// The ConfigMap should contain a single cluster
	if len(insecureConfig.Clusters) != 1 {
		return nil, fmt.Errorf("expected the kubeconfig file in the %s ConfigMap to have a single cluster, but it had %d", bootstrapapi.ConfigMapClusterInfo, len(insecureConfig.Clusters))
	}

	// If no TLS root CA pinning was specified, we're done
	if pubKeyPins.Empty() {
		klog.V(1).Infof("[discovery] Cluster info signature and contents are valid and no TLS pinning was specified, will use API Server %q", endpoint)
		return insecureConfig, nil
	}

	// Load and validate the cluster CA from the insecure kubeconfig
	clusterCABytes, err := validateClusterCA(insecureConfig, pubKeyPins)
	if err != nil {
		return nil, err
	}

	// Now that we know the cluster CA, connect back a second time validating with that CA
	secureBootstrapConfig := buildSecureBootstrapKubeConfig(endpoint, clusterCABytes, clusterName)

	klog.V(1).Infof("[discovery] Requesting info from %q again to validate TLS against the pinned public key", endpoint)
	secureClusterInfo, err := getClusterInfoFromControlPlane(client, secureBootstrapConfig, token, interval, duration)
	if err != nil {
		return nil, err
	}

	// Pull the kubeconfig from the securely-obtained ConfigMap and validate that it's the same as what we found the first time
	secureKubeconfigBytes := []byte(secureClusterInfo.Data[bootstrapapi.KubeConfigKey])
	if !bytes.Equal(secureKubeconfigBytes, insecureKubeconfigBytes) {
		return nil, fmt.Errorf("the second kubeconfig from the %s ConfigMap (using validated TLS) was different from the first", bootstrapapi.ConfigMapClusterInfo)
	}

	secureKubeconfig, err := clientcmd.Load(secureKubeconfigBytes)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse the kubeconfig file in the %s ConfigMap: %w", bootstrapapi.ConfigMapClusterInfo, err)
	}

	klog.V(1).Infof("[discovery] Cluster info signature and contents are valid and TLS certificate validates against pinned roots, will use API Server %q", endpoint)

	return secureKubeconfig, nil
}
// generateKeyAndCSR generate private key and csr
func generateKeyAndCSR(clusterName string) (*rsa.PrivateKey, []byte, error) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   ClusterPermissionPrefix + clusterName,
			Organization: []string{ClusterPermissionGroups},
		},
	}, pk)
	if err != nil {
		return nil, nil, err
	}

	return pk, csr, nil
}

// CreateWithCert creates a KubeConfig object with access to the API server with a cert
func CreateWithCert(serverURL, clusterName, userName string, caCert []byte, cert []byte, key []byte) *clientcmdapi.Config {
	config := CreateBasic(serverURL, clusterName, userName, caCert)
	config.AuthInfos[userName] = &clientcmdapi.AuthInfo{
		ClientCertificateData: cert,
		ClientKeyData:         key,
	}
	return config
}
// checkFileIfExist validates if the given file already exist.
func checkFileIfExist(filePath string) error {
	klog.V(1).Infof("Validating the existence of file %s", filePath)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("%s already exists", filePath)
	}
	return nil
}

// buildInsecureBootstrapKubeConfig makes a kubeconfig object that connects insecurely to the API Server for bootstrapping purposes
func buildInsecureBootstrapKubeConfig(endpoint, clustername string) *clientcmdapi.Config {
	controlPlaneEndpoint := fmt.Sprintf("https://%s", endpoint)
	bootstrapConfig := CreateBasic(controlPlaneEndpoint, clustername, BootstrapUserName, []byte{})
	bootstrapConfig.Clusters[clustername].InsecureSkipTLSVerify = true
	return bootstrapConfig
}

// WriteToDisk writes a KubeConfig object down to disk with mode 0600
func WriteToDisk(filename string, kubeconfig *clientcmdapi.Config) error {
	err := clientcmd.WriteToFile(*kubeconfig, filename)
	if err != nil {
		return err
	}

	return nil
}

*/
//// discoveryBootstrapConfigAndClusterInfo get bootstrap-config and cluster-info from control plane
//func (o *CommandRegisterOption) discoveryBootstrapConfigAndClusterInfo(bootstrapKubeConfigFile, parentCommand string) (*kubeclient.Clientset, *clientcmdapi.Cluster, error) {
//	/*
//		config, err := retrieveValidatedConfigInfo(nil, o.BootstrapToken, o.Timeout, DiscoveryRetryInterval, parentCommand)
//		if err != nil {
//
//			return nil, nil, fmt.Errorf("couldn't validate the identity of the API Server: %w", err)
//		}
//
//		klog.V(1).Info("[discovery] Using provided TLSBootstrapToken as authentication credentials for the join process")
//	*/
//	//clusterinfo := tokenutil.GetClusterFromKubeConfig(config, "")
//	//tlsBootstrapCfg := CreateWithToken(
//	//	clusterinfo.Server,
//	//	DefaultClusterName,
//	//	TokenUserName,
//	//	clusterinfo.CertificateAuthorityData,
//	//	o.BootstrapToken.Token,
//	//)
//	// Write the TLS-Bootstrapped karmada-agent config file down to disk
//	//klog.V(1).Infof("[discovery] writing bootstrap karmada-agent config file at %s", bootstrapKubeConfigFile)
//	//if err := WriteToDisk(bootstrapKubeConfigFile, tlsBootstrapCfg); err != nil {
//	//	return nil, nil, fmt.Errorf("couldn't save %s to disk: %w", KarmadaAgentBootstrapKubeConfigFileName, err)
//	//}
//
//	// Write the ca certificate to disk so karmada-agent can use it for authentication
//	//cluster := tlsBootstrapCfg.Contexts[tlsBootstrapCfg.CurrentContext].Cluster
//	//caPath := o.CACertPath
//	//if _, err := os.Stat(caPath); os.IsNotExist(err) {
//	//	klog.V(1).Infof("[discovery] writing CA certificate at %s", caPath)
//	//	if err := certutil.WriteCert(caPath, tlsBootstrapCfg.Clusters[cluster].CertificateAuthorityData); err != nil {
//	//		return nil, nil, fmt.Errorf("couldn't save the CA certificate to disk: %w", err)
//	//	}
//	//}
//
//	//bootstrapClient, err := ClientSetFromFile(bootstrapKubeConfigFile)
//	//if err != nil {
//	//	return nil, nil, fmt.Errorf("couldn't create client from kubeconfig file %q", bootstrapKubeConfigFile)
//	//}
//
//	return nil, nil, nil
//}
/*
// constructKarmadaAgentConfig construct the final kubeconfig used by karmada-agent
func (o *CommandRegisterOption) constructKarmadaAgentConfig(bootstrapClient *kubeclient.Clientset, karmadaClusterInfo *clientcmdapi.Cluster) (*clientcmdapi.Config, error) {
	//var cert []byte
	//
	//pk, csr, err := generateKeyAndCSR(o.ClusterName)
	//if err != nil {
	//	return nil, err
	//}
	//
	//pkData, err := keyutil.MarshalPrivateKeyToPEM(pk)
	//if err != nil {
	//	return nil, err
	//}
	//
	//csrName := o.ClusterName + "-" + k8srand.String(5)
	//
	//certificateSigningRequest := &certificatesv1.CertificateSigningRequest{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: csrName,
	//	},
	//	Spec: certificatesv1.CertificateSigningRequestSpec{
	//		Request: pem.EncodeToMemory(&pem.Block{
	//			Type:  certutil.CertificateRequestBlockType,
	//			Bytes: csr,
	//		}),
	//		SignerName:        SignerName,
	//		ExpirationSeconds: &o.CertExpirationSeconds,
	//		Usages: []certificatesv1.KeyUsage{
	//			certificatesv1.UsageDigitalSignature,
	//			certificatesv1.UsageKeyEncipherment,
	//			certificatesv1.UsageClientAuth,
	//		},
	//	},
	//}
	//
	//_, err = bootstrapClient.CertificatesV1().CertificateSigningRequests().Create(context.TODO(), certificateSigningRequest, metav1.CreateOptions{})
	//if err != nil {
	//	return nil, err
	//}

	//klog.V(1).Infof("Waiting for the client certificate to be issued")
	//err = wait.Poll(1*time.Second, o.Timeout, func() (done bool, err error) {
	//	csrOK, err := bootstrapClient.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), csrName, metav1.GetOptions{})
	//	if err != nil {
	//		return false, fmt.Errorf("failed to get the cluster csr %s. err: %v", o.ClusterName, err)
	//	}
	//
	//	if csrOK.Status.Certificate != nil {
	//		klog.V(1).Infof("Signing certificate successfully")
	//		cert = csrOK.Status.Certificate
	//		return true, nil
	//	}
	//
	//	klog.V(1).Infof("Waiting for the client certificate to be issued")
	//	return false, nil
	//})
	//if err != nil {
	//	return nil, err
	//}

	//karmadaAgentCfg := CreateWithCert(
	//	karmadaClusterInfo.Server,
	//	DefaultClusterName,
	//	o.ClusterName,
	//	karmadaClusterInfo.CertificateAuthorityData,
	//	cert,
	//	pkData,
	//)
	//
	//kubeConfigFile := filepath.Join(KarmadaDir, KarmadaAgentKubeConfigFileName)

	// Write the karmada-agent config file down to disk
	//klog.V(1).Infof("writing bootstrap karmada-agent config file at %s", kubeConfigFile)
	//if err := WriteToDisk(kubeConfigFile, karmadaAgentCfg); err != nil {
	//	return nil, fmt.Errorf("couldn't save %s to disk: %w", KarmadaAgentKubeConfigFileName, err)
	//}

	return nil, nil
}
*/

/*
- /bin/karmada-agent
- --karmada-kubeconfig=/etc/kubeconfig/karmada-kubeconfig
- --cluster-name=test-cluster
- --cluster-api-endpoint=https://192.168.10.7:6443
- --cluster-status-update-frequency=10s
- --bind-address=0.0.0.0
- --secure-port=10357
- --feature-gates=CustomizedClusterResourceModeling=true,MultiClusterService=true
- --v=4
*/
