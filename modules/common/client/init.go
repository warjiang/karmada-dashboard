package client

import (
	"errors"
	"fmt"
	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"strings"
	karmadaerrors "warjiang/karmada-dashboard/errors"
)

func buildConfigFromAuthInfo(authInfo *clientcmdapi.AuthInfo) (*rest.Config, error) {
	cmdCfg := clientcmdapi.NewConfig()

	cmdCfg.Clusters[DefaultCmdConfigName] = &clientcmdapi.Cluster{
		Server:                   karmadaRestConfig.Host,
		CertificateAuthority:     karmadaRestConfig.TLSClientConfig.CAFile,
		CertificateAuthorityData: karmadaRestConfig.TLSClientConfig.CAData,
		InsecureSkipTLSVerify:    karmadaRestConfig.TLSClientConfig.Insecure,
	}

	cmdCfg.AuthInfos[DefaultCmdConfigName] = authInfo

	cmdCfg.Contexts[DefaultCmdConfigName] = &clientcmdapi.Context{
		Cluster:  DefaultCmdConfigName,
		AuthInfo: DefaultCmdConfigName,
	}

	cmdCfg.CurrentContext = DefaultCmdConfigName

	return clientcmd.NewDefaultClientConfig(
		*cmdCfg,
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
}

func buildAuthInfo(request *http.Request) (*clientcmdapi.AuthInfo, error) {
	if !HasAuthorizationHeader(request) {
		return nil, karmadaerrors.NewUnauthorized(karmadaerrors.MsgLoginUnauthorizedError)
	}

	token := GetBearerToken(request)
	authInfo := &clientcmdapi.AuthInfo{
		Token:                token,
		ImpersonateUserExtra: make(map[string][]string),
	}

	handleImpersonation(authInfo, request)
	return authInfo, nil
}

func handleImpersonation(authInfo *clientcmdapi.AuthInfo, request *http.Request) {
	user := request.Header.Get(ImpersonateUserHeader)
	groups := request.Header[ImpersonateGroupHeader]

	if len(user) == 0 {
		return
	}

	// Impersonate user
	authInfo.Impersonate = user

	// Impersonate groups if available
	if len(groups) > 0 {
		authInfo.ImpersonateGroups = groups
	}

	// Add extra impersonation fields if available
	for name, values := range request.Header {
		if strings.HasPrefix(name, ImpersonateUserExtraHeader) {
			extraName := strings.TrimPrefix(name, ImpersonateUserExtraHeader)
			authInfo.ImpersonateUserExtra[extraName] = values
		}
	}
}

func karmadaConfigFromRequest(request *http.Request) (*rest.Config, error) {
	authInfo, err := buildAuthInfo(request)
	if err != nil {
		return nil, err
	}

	return buildConfigFromAuthInfo(authInfo)
}

func karmadaClientFromRequest(request *http.Request) (karmadaclientset.Interface, error) {
	config, err := karmadaConfigFromRequest(request)
	if err != nil {
		return nil, err
	}

	return karmadaclientset.NewForConfig(config)
}

func GenerateAPIConfigFromKubeconfigFile(kubeconfigPath string) (*clientcmdapi.Config, error) {
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

var (
	kubernetesRestConfig   *rest.Config
	kubernetesApiConfig    *clientcmdapi.Config
	inClusterClient        kubeclient.Interface
	karmadaRestConfig      *rest.Config
	karmadaApiConfig       *clientcmdapi.Config
	inClusterKarmadaClient karmadaclientset.Interface
)

type configBuilder struct {
	userAgent      string
	kubeconfigPath string
	kubeContext    string
	masterUrl      string
	insecure       bool
	contentType    string
}

type Option func(*configBuilder)

func WithUserAgent(agent string) Option {
	return func(c *configBuilder) {
		c.userAgent = agent
	}
}

func WithKubeconfig(path string) Option {
	return func(c *configBuilder) {
		c.kubeconfigPath = path
	}
}

func WithKubeContext(kubecontext string) Option {
	return func(c *configBuilder) {
		c.kubeContext = kubecontext
	}
}

func WithMasterUrl(url string) Option {
	return func(c *configBuilder) {
		c.masterUrl = url
	}
}

func WithInsecureTLSSkipVerify(insecure bool) Option {
	return func(c *configBuilder) {
		c.insecure = insecure
	}
}

func WithContentType(contentType string) Option {
	return func(c *configBuilder) {
		c.contentType = contentType
	}
}

func (in *configBuilder) buildRestConfig() (*rest.Config, error) {
	if len(in.kubeconfigPath) == 0 {
		return nil, errors.New("must specify kubeconfig")
	}
	klog.InfoS("Using kubeconfig", "kubeconfig", in.kubeconfigPath)

	restConfig, err := LoadRestConfig(in.kubeconfigPath, in.kubeContext)
	if err != nil {
		return nil, err
	}

	if len(in.masterUrl) > 0 {
		klog.InfoS("Using apiserver-host location", "masterUrl", in.masterUrl)
		restConfig.Host = in.masterUrl
	}

	restConfig.QPS = DefaultQPS
	restConfig.Burst = DefaultBurst
	// TODO: make clear that why karmada apiserver seems only can use application/json, however kubernetest apiserver can use "application/vnd.kubernetes.protobuf"
	restConfig.UserAgent = DefaultUserAgent + "/" + in.userAgent
	restConfig.TLSClientConfig.Insecure = in.insecure

	return restConfig, nil
}

func (in *configBuilder) buildApiConfig() (*clientcmdapi.Config, error) {
	if len(in.kubeconfigPath) == 0 {
		return nil, errors.New("must specify kubeconfig")
	}
	klog.InfoS("Using kubeconfig", "kubeconfig", in.kubeconfigPath)
	apiConfig, err := LoadApiConfig(in.kubeconfigPath, in.kubeContext)
	if err != nil {
		return nil, err
	}
	if len(in.masterUrl) > 0 {
		klog.InfoS("Using apiserver-host location", "masterUrl", in.masterUrl)
		apiConfig.Clusters[apiConfig.CurrentContext].Server = in.masterUrl
	}
	return apiConfig, nil
}

func newConfigBuilder(options ...Option) *configBuilder {
	builder := &configBuilder{}

	for _, opt := range options {
		opt(builder)
	}

	return builder
}

func isKubeInitialized() bool {
	if kubernetesRestConfig == nil || kubernetesApiConfig == nil {
		klog.Errorf(`warjiang/karmada-dashboard/client' package has not been initialized properly. Run 'client.InitKubeConfig(...)' to initialize it. `)
		return false
	}
	return true
}

func isKarmadaInitialized() bool {
	if karmadaRestConfig == nil || karmadaApiConfig == nil {
		klog.Errorf(`warjiang/karmada-dashboard/client' package has not been initialized properly. Run 'client.InitKarmadaConfig(...)' to initialize it. `)
		return false
	}
	return true
}

func InitKubeConfig(options ...Option) {
	builder := newConfigBuilder(options...)

	restConfig, err := builder.buildRestConfig()
	if err != nil {
		klog.Errorf("Could not init client config: %s", err)
		os.Exit(1)
	}
	kubernetesRestConfig = restConfig

	apiConfig, err := builder.buildApiConfig()
	if err != nil {
		klog.Errorf("Could not init api config: %s", err)
		os.Exit(1)
	}
	kubernetesApiConfig = apiConfig
}

func InitKarmadaConfig(options ...Option) {
	builder := newConfigBuilder(options...)

	restConfig, err := builder.buildRestConfig()
	if err != nil {
		klog.Errorf("Could not init client config: %s", err)
		os.Exit(1)
	}
	karmadaRestConfig = restConfig

	apiConfig, err := builder.buildApiConfig()
	if err != nil {
		klog.Errorf("Could not init api config: %s", err)
		os.Exit(1)
	}
	karmadaApiConfig = apiConfig
}

func GetKubeConfig() (*rest.Config, *clientcmdapi.Config, error) {
	if !isKubeInitialized() {
		return nil, nil, fmt.Errorf("client package not initialized")
	}
	return kubernetesRestConfig, kubernetesApiConfig, nil
}

func GetKarmadaConfig() (*rest.Config, *clientcmdapi.Config, error) {
	if !isKarmadaInitialized() {
		return nil, nil, fmt.Errorf("client package not initialized")
	}
	return karmadaRestConfig, karmadaApiConfig, nil
}
