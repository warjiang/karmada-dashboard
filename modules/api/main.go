package main

import (
	"k8s.io/klog/v2"
	"os"
	"warjiang/karmada-dashboard/api/pkg/args"
	"warjiang/karmada-dashboard/api/pkg/environment"
	"warjiang/karmada-dashboard/api/pkg/router"
	"warjiang/karmada-dashboard/client"

	_ "warjiang/karmada-dashboard/api/pkg/routes/cluster"
)

func main() {
	klog.InfoS("Starting Kubernetes Dashboard API", "version", environment.Version)
	client.Init(
		client.WithUserAgent(environment.UserAgent()),
		client.WithKubeconfig(args.KubeconfigPath()),
		client.WithMasterUrl(args.ApiServerHost()),
		client.WithInsecureTLSSkipVerify(args.ApiServerSkipTLSVerify()),
	)
	if !args.IsProxyEnabled() {
		ensureAPIServerConnectionOrDie()
	} else {
		klog.Info("Running in proxy mode. InClusterClient connections will be disabled.")
	}

	if err := router.Router().Run("127.0.0.1:8000"); err != nil {
		klog.ErrorS(err, "Router error")
		os.Exit(1)
	}
}

func ensureAPIServerConnectionOrDie() {
	versionInfo, err := client.InClusterClient().Discovery().ServerVersion()
	if err != nil {
		handleFatalInitError(err)
	}

	klog.InfoS("Successful initial request to the apiserver", "version", versionInfo.String())
}

/**
 * Handles fatal init error that prevents server from doing any work. Prints verbose error
 * message and quits the server.
 */
func handleFatalInitError(err error) {
	klog.Fatalf("Error while initializing connection to Kubernetes apiserver. "+
		"This most likely means that the cluster is misconfigured (e.g., it has "+
		"invalid apiserver certificates or service account's configuration) or the "+
		"--apiserver-host param points to a server that does not exist. Reason: %s\n"+
		"Refer to our FAQ and wiki pages for more information: "+
		"https://github.com/kubernetes/dashboard/wiki/FAQ", err)
}
