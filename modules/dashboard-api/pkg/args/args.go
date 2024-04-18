package args

import (
	"flag"
	"fmt"
	"net"
	"strconv"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"warjiang/karmada-dashboard/csrf"
	"warjiang/karmada-dashboard/helpers"
)

const (
	LogLevelDefault  = klog.Level(0)
	LogLevelMinimal  = LogLevelDefault
	LogLevelInfo     = klog.Level(1)
	LogLevelVerbose  = klog.Level(2)
	LogLevelExtended = klog.Level(3)
	LogLevelDebug    = klog.Level(4)
	LogLevelTrace    = klog.Level(5)
)

var (
	argKubeConfig                    = pflag.String("kubeconfig", "", "Path to the host cluster kubeconfig file.")
	argKubeContext                   = pflag.String("context", "", "The name of the kubeconfig context to use.")
	argSkipKubeApiserverTLSVerify    = pflag.Bool("skip-kube-apiserver-tls-verify", false, "enable if connection with remote Kubernetes API server should skip TLS verify")
	argKarmadaConfig                 = pflag.String("karmada-kubeconfig", "/etc/karmada/karmada-apiserver.config", "Path to the karmada control plane kubeconfig file.")
	argKarmadaContext                = pflag.String("karmada-context", "", "The name of the karmada control plane kubeconfig context to use.")
	argSkipKarmadaApiserverTLSVerify = pflag.Bool("skip-karmada-apiserver-tls-verify", false, "enable if connection with remote Karmada API server should skip TLS verify")
	argInsecurePort                  = pflag.Int("insecure-port", 8000, "port to listen to for incoming HTTP requests")
	argPort                          = pflag.Int("port", 8001, "secure port to listen to for incoming HTTPS requests")
	argInsecureBindAddress           = pflag.IP("insecure-bind-address", net.IPv4(127, 0, 0, 1), "IP address on which to serve the --insecure-port, set to 127.0.0.1 for all interfaces")
	argBindAddress                   = pflag.IP("bind-address", net.IPv4(127, 0, 0, 1), "IP address on which to serve the --port, set to 0.0.0.0 for all interfaces")
	argDefaultCertDir                = pflag.String("default-cert-dir", "/certs", "directory path containing files from --tls-cert-file and --tls-key-file, used also when auto-generating certificates flag is set")
	argCertFile                      = pflag.String("tls-cert-file", "", "file containing the default x509 certificate for HTTPS")
	argKeyFile                       = pflag.String("tls-key-file", "", "file containing the default x509 private key matching --tls-cert-file")
	argAutoGenerateCertificates      = pflag.Bool("auto-generate-certificates", false, "enables automatic certificates generation used to serve HTTPS")
	argNamespace                     = pflag.String("namespace", helpers.GetEnv("POD_NAMESPACE", "karmada-dashboard"), "Namespace to use when accessing Dashboard specific resources, i.e. configmap")
	argDisableCSRFProtection         = pflag.Bool("disable-csrf-protection", false, "allows disabling CSRF protection")
	argOpenAPIEnabled                = pflag.Bool("openapi-enabled", false, "enables OpenAPI v2 endpoint under '/apidocs.json'")
)

func init() {
	// Init klog
	fs := flag.NewFlagSet("", flag.PanicOnError)
	klog.InitFlags(fs)

	// Default log level to 1
	_ = fs.Set("v", "1")

	pflag.CommandLine.AddGoFlagSet(fs)
	pflag.Parse()

	if IsCSRFProtectionEnabled() {
		csrf.Ensure()
	}
}

func KubeConfigPath() string {
	return *argKubeConfig
}

func KubeContext() string {
	return *argKubeContext
}

func SkipKubeApiserverTLSVerify() bool {
	return *argSkipKubeApiserverTLSVerify
}

func KarmadaConfigPath() string {
	return *argKarmadaConfig
}

func KarmadaContext() string {
	return *argKarmadaContext
}
func SkipKarmadaApiserverTLSVerify() bool {
	return *argSkipKarmadaApiserverTLSVerify
}

func Address() string {
	return fmt.Sprintf("%s:%d", *argBindAddress, *argPort)
}

func InsecureAddress() string {
	return fmt.Sprintf("%s:%d", *argInsecureBindAddress, *argInsecurePort)
}

func DefaultCertDir() string {
	return *argDefaultCertDir
}

func CertFile() string {
	return *argCertFile
}

func KeyFile() string {
	return *argKeyFile
}
func AutogenerateCertificates() bool {
	return *argAutoGenerateCertificates
}

func APILogLevel() klog.Level {
	v := pflag.Lookup("v")
	if v == nil {
		return LogLevelDefault
	}

	level, err := strconv.ParseInt(v.Value.String(), 10, 32)
	if err != nil {
		klog.ErrorS(err, "Could not parse log level", "level", v.Value.String())
		return LogLevelDefault
	}

	return klog.Level(level)
}

func Namespace() string {
	return *argNamespace
}

func IsCSRFProtectionEnabled() bool {
	return !*argDisableCSRFProtection
}

func IsOpenAPIEnabled() bool {
	return *argOpenAPIEnabled
}
