package options

import (
	"github.com/spf13/pflag"
	"k8s.io/dashboard/helpers"
	"net"
)

// Options contains everything necessary to create and run api.
type Options struct {
	BindAddress                   net.IP
	Port                          int
	InsecureBindAddress           net.IP
	InsecurePort                  int
	StaticDir                     string
	I18nDir                       string
	EnableApiProxy                bool
	ApiProxyEndpoint              string
	DashboardConfigPath           string
	ArgNamespace                  string
	ArgSettingsConfigMapName      string
	ArgSystemBanner               string
	ArgSystemBannerSeverity       string
	EnableMemberClusterApiProxy   bool
	MemberClusterApiProxyEndpoint string
}

var opts *Options

func NewOptions() *Options {
	if opts == nil {
		opts = &Options{}
	}
	return opts
}

func GetOptions() *Options {
	return NewOptions()
}

// AddFlags adds flags of api to the specified FlagSet
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}
	fs.IPVar(&o.BindAddress, "bind-address", net.IPv4(127, 0, 0, 1), "IP address on which to serve the --port, set to 0.0.0.0 for all interfaces")
	fs.IntVar(&o.Port, "port", 8001, "secure port to listen to for incoming HTTPS requests")
	fs.IPVar(&o.InsecureBindAddress, "insecure-bind-address", net.IPv4(127, 0, 0, 1), "IP address on which to serve the --insecure-port, set to 0.0.0.0 for all interfaces")
	fs.IntVar(&o.InsecurePort, "insecure-port", 8000, "port to listen to for incoming HTTP requests")
	fs.StringVar(&o.StaticDir, "static-dir", "./static", "directory to serve static files")
	fs.StringVar(&o.I18nDir, "i18n-dir", "./i18n", "directory to serve i18n files")
	fs.BoolVar(&o.EnableApiProxy, "enable-api-proxy", true, "whether enable proxy to karmada-dashboard-api, if set true, all requests with /api prefix will be proxyed to karmada-dashboard-api.karmada-system.svc.cluster.local")
	fs.StringVar(&o.ApiProxyEndpoint, "api-proxy-endpoint", "http://karmada-dashboard-api.karmada-system.svc.cluster.local:8000", "karmada-dashboard-api endpoint")
	fs.StringVar(&o.DashboardConfigPath, "dashboard-config-path", "./config/dashboard-config.yaml", "path to dashboard config file")

	fs.StringVar(&o.ArgNamespace, "namespace", helpers.GetEnv("POD_NAMESPACE", "kubernetes-dashboard"), "Namespace to use when creating Dashboard specific resources, i.e. settings config map")
	fs.StringVar(&o.ArgSettingsConfigMapName, "settings-config-map-name", "kubernetes-dashboard-settings", "Name of a config map, that stores settings")
	fs.StringVar(&o.ArgSystemBanner, "system-banner", "", "system banner message displayed in the app if non-empty, it accepts simple HTML")
	fs.StringVar(&o.ArgSystemBannerSeverity, "system-banner-severity", "INFO", "severity of system banner, should be one of 'INFO', 'WARNING' or 'ERROR'")
	fs.BoolVar(&o.EnableMemberClusterApiProxy, "enable-member-cluster-api-proxy", true, "whether enable proxy to member cluster, if set true, all requests with /member/api/v1/member/ prefix will be proxied with karmada unified authentication")
	fs.StringVar(&o.MemberClusterApiProxyEndpoint, "member-cluster-api-proxy-endpoint", "http://kubernetes-dashboard-api.karmada-system.svc.cluster.local:8000", "member-cluster-api-proxy-endpoint")
}
