package main

import (
	"fmt"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	// DefaultConfigFlags It composes the set of values necessary for obtaining a REST client config with default values set.
	var DefaultConfigFlags = genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().WithDiscoveryBurst(300).WithDiscoveryQPS(50.0)
	fmt.Println(*DefaultConfigFlags.KubeConfig)
}
