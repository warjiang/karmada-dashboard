// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"strings"
	"warjiang/karmada-dashboard/errors"
)

var (
	inClusterClient karmadaclientset.Interface
	baseConfig      *rest.Config
)

func buildConfigFromAuthInfo(authInfo *api.AuthInfo) (*rest.Config, error) {
	cmdCfg := api.NewConfig()

	cmdCfg.Clusters[DefaultCmdConfigName] = &api.Cluster{
		Server:                   baseConfig.Host,
		CertificateAuthority:     baseConfig.TLSClientConfig.CAFile,
		CertificateAuthorityData: baseConfig.TLSClientConfig.CAData,
		InsecureSkipTLSVerify:    baseConfig.TLSClientConfig.Insecure,
	}

	cmdCfg.AuthInfos[DefaultCmdConfigName] = authInfo

	cmdCfg.Contexts[DefaultCmdConfigName] = &api.Context{
		Cluster:  DefaultCmdConfigName,
		AuthInfo: DefaultCmdConfigName,
	}

	cmdCfg.CurrentContext = DefaultCmdConfigName

	return clientcmd.NewDefaultClientConfig(
		*cmdCfg,
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
}

func buildAuthInfo(request *http.Request) (*api.AuthInfo, error) {
	if !HasAuthorizationHeader(request) {
		return nil, errors.NewUnauthorized(errors.MsgLoginUnauthorizedError)
	}

	token := GetBearerToken(request)
	authInfo := &api.AuthInfo{
		Token:                token,
		ImpersonateUserExtra: make(map[string][]string),
	}

	handleImpersonation(authInfo, request)
	return authInfo, nil
}

func handleImpersonation(authInfo *api.AuthInfo, request *http.Request) {
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

func configFromRequest(request *http.Request) (*rest.Config, error) {
	authInfo, err := buildAuthInfo(request)
	if err != nil {
		return nil, err
	}

	return buildConfigFromAuthInfo(authInfo)
}

func clientFromRequest(request *http.Request) (karmadaclientset.Interface, error) {
	config, err := configFromRequest(request)
	if err != nil {
		return nil, err
	}

	return karmadaclientset.NewForConfig(config)
}

func isInitialized() bool {
	if baseConfig == nil {
		klog.Errorf(`k8s.io/dasboard/client' package has not been initialized properly. Run 'client.Init(...)' to initialize it. `)
		return false
	}

	return true
}

type configBuilder struct {
	userAgent      string
	kubeconfigPath string
	masterUrl      string
	insecure       bool
}

func (in *configBuilder) buildBaseConfig() (*rest.Config, error) {
	if len(in.kubeconfigPath) == 0 && len(in.masterUrl) == 0 {
		klog.Info("Using in-cluster config")
		return rest.InClusterConfig()
	}

	if len(in.kubeconfigPath) > 0 {
		klog.InfoS("Using kubeconfig", "kubeconfig", in.kubeconfigPath)
	}

	if len(in.masterUrl) > 0 {
		klog.InfoS("Using apiserver-host location", "masterUrl", in.masterUrl)
	}

	config, err := clientcmd.BuildConfigFromFlags(in.masterUrl, in.kubeconfigPath)
	if err != nil {
		return nil, err
	}

	config.QPS = DefaultQPS
	config.Burst = DefaultBurst
	config.ContentType = DefaultContentType
	config.UserAgent = DefaultUserAgent + "/" + in.userAgent
	config.TLSClientConfig.Insecure = in.insecure

	return config, nil
}

type Option func(*configBuilder)

func newConfigBuilder(options ...Option) *configBuilder {
	builder := &configBuilder{}

	for _, opt := range options {
		opt(builder)
	}

	return builder
}

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

func Init(options ...Option) {
	builder := newConfigBuilder(options...)

	config, err := builder.buildBaseConfig()
	if err != nil {
		klog.Errorf("Could not init kubernetes client config: %s", err)
		os.Exit(1)
	}

	baseConfig = config
}