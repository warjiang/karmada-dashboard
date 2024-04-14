package main

import (
	"k8s.io/klog/v2"
	"warjiang/karmada-dashboard/web/pkg/args"
	"warjiang/karmada-dashboard/web/pkg/environment"
	"warjiang/karmada-dashboard/web/pkg/router"
	_ "warjiang/karmada-dashboard/web/pkg/systembanner"
)

func main() {
	klog.InfoS("Starting Kubernetes Dashboard Web", "version", environment.Version)
	klog.V(1).InfoS("Listening and serving insecurely on", "address", args.InsecureAddress())
	if err := router.Router().Run(args.InsecureAddress()); err != nil {
		klog.Fatalf("Router error: %s", err)
	}
}
