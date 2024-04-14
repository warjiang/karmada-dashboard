package main

import (
	"k8s.io/klog/v2"
	"os"
	"warjiang/karmada-dashboard/auth/pkg/args"
	"warjiang/karmada-dashboard/auth/pkg/environment"
	"warjiang/karmada-dashboard/auth/pkg/router"
	"warjiang/karmada-dashboard/client"

	// Importing route packages forces route registration
	_ "warjiang/karmada-dashboard/auth/pkg/routes/csrftoken"
	_ "warjiang/karmada-dashboard/auth/pkg/routes/login"
	_ "warjiang/karmada-dashboard/auth/pkg/routes/me"
)

func main() {
	klog.InfoS("Starting Kubernetes Dashboard Auth", "version", environment.Version)

	client.Init(
		client.WithUserAgent(environment.UserAgent()),
		client.WithKubeconfig(args.KubeconfigPath()),
		client.WithInsecureTLSSkipVerify(true),
	)

	klog.V(1).InfoS("Listening and serving insecurely on", "address", args.Address())
	if err := router.Router().Run(args.Address()); err != nil {
		klog.ErrorS(err, "Router error")
		os.Exit(1)
	}
}
