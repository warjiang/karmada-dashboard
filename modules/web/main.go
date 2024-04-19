package main

import (
	"crypto/elliptic"
	"k8s.io/klog/v2"
	"warjiang/karmada-dashboard/certificates"
	"warjiang/karmada-dashboard/certificates/ecdsa"
	"warjiang/karmada-dashboard/client"
	"warjiang/karmada-dashboard/web/pkg/args"
	"warjiang/karmada-dashboard/web/pkg/environment"
	"warjiang/karmada-dashboard/web/pkg/router"

	// Importing route packages forces route registration
	_ "warjiang/karmada-dashboard/web/pkg/locale"
	_ "warjiang/karmada-dashboard/web/pkg/systembanner"
)

func main() {
	klog.InfoS("Starting Karmada Dashboard Web", "version", environment.Version)

	client.InitKubeConfig(
		client.WithUserAgent(environment.UserAgent()),
		client.WithKubeconfig(args.KubeconfigPath()),
		client.WithKubeContext(args.KubeContext()),
	)

	certCreator := ecdsa.NewECDSACreator(args.KeyFile(), args.CertFile(), elliptic.P256())
	certManager := certificates.NewCertManager(certCreator, args.DefaultCertDir(), args.AutoGenerateCertificates())
	certPath, keyPath, err := certManager.GetCertificatePaths()
	if err != nil {
		klog.Fatalf("Error while loading dashboard server certificates. Reason: %s", err)
	}

	if len(certPath) != 0 && len(keyPath) != 0 {
		klog.V(1).InfoS("Listening and serving securely on", "address", args.Address())
		if err := router.Router().RunTLS(args.Address(), certPath, keyPath); err != nil {
			klog.Fatalf("Router error: %s", err)
		}
	} else {
		klog.V(1).InfoS("Listening and serving insecurely on", "address", args.InsecureAddress())
		if err := router.Router().Run(args.InsecureAddress()); err != nil {
			klog.Fatalf("Router error: %s", err)
		}
	}
}
