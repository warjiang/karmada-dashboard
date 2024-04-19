package main

import (
	"crypto/elliptic"
	"crypto/tls"
	"k8s.io/klog/v2"
	"net/http"
	"warjiang/karmada-dashboard/certificates"
	"warjiang/karmada-dashboard/certificates/ecdsa"
	"warjiang/karmada-dashboard/client"
	"warjiang/karmada-dashboard/dashboard-api/pkg/args"
	"warjiang/karmada-dashboard/dashboard-api/pkg/environment"
	"warjiang/karmada-dashboard/dashboard-api/pkg/router"

	// Importing route packages forces route registration
	_ "warjiang/karmada-dashboard/dashboard-api/pkg/routes/cluster"
	_ "warjiang/karmada-dashboard/dashboard-api/pkg/routes/csrftoken"
	_ "warjiang/karmada-dashboard/dashboard-api/pkg/routes/login"
	_ "warjiang/karmada-dashboard/dashboard-api/pkg/routes/me"
)

func main() {
	klog.InfoS("Starting Karmada Dashboard API", "version", environment.Version)
	client.InitKarmadaConfig(
		client.WithUserAgent(environment.UserAgent()),
		client.WithKubeconfig(args.KarmadaConfigPath()),
		client.WithKubeContext(args.KarmadaContext()),
		client.WithInsecureTLSSkipVerify(args.SkipKarmadaApiserverTLSVerify()),
	)

	ensureAPIServerConnectionOrDie()

	certCreator := ecdsa.NewECDSACreator(args.KeyFile(), args.CertFile(), elliptic.P256())
	certManager := certificates.NewCertManager(certCreator, args.DefaultCertDir(), args.AutogenerateCertificates())
	certs, err := certManager.GetCertificates()
	if err != nil {
		handleFatalInitServingCertError(err)
	}

	if args.IsOpenAPIEnabled() {
		// TODO: config swagger handler
		klog.Info("Enabling OpenAPI endpoint on /apidocs.json")
	}

	if err != nil {
		handleFatalInitServingCertError(err)
	}

	if certs != nil {
		serveTLS(certs)
	} else {
		serve()
	}

	select {}
}

func serve() {
	klog.V(1).InfoS("Listening and serving on", "address", args.InsecureAddress())
	go func() {
		klog.Fatal(router.Router().Run(args.InsecureAddress()))
	}()
}

func serveTLS(certificates []tls.Certificate) {
	klog.V(1).InfoS("Listening and serving on", "address", args.Address())
	r := router.Router()
	// Run gin with custom TLSConfig: https://github.com/gin-gonic/gin/issues/1099
	tlsConfig := &tls.Config{
		Certificates: certificates,
		MinVersion:   tls.VersionTLS12,
	}
	server := &http.Server{
		Addr:      args.Address(),
		Handler:   r,
		TLSConfig: tlsConfig,
	}
	go func() { klog.Fatal(server.ListenAndServeTLS("", "")) }()
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
		"This most likely means that the cluster is misconfigured. Reason: %s\n", err)
}

/**
 * Handles fatal init errors encountered during service cert loading.
 */
func handleFatalInitServingCertError(err error) {
	klog.Fatalf("Error while loading dashboard server certificates. Reason: %s", err)
}
