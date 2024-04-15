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
	"fmt"
	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	"k8s.io/klog/v2"
	"net/http"
	"os"
)

func InClusterClient() karmadaclientset.Interface {
	if !isInitialized() {
		return nil
	}

	if inClusterClient != nil {
		return inClusterClient
	}

	// init on-demand only
	c, err := karmadaclientset.NewForConfig(baseConfig)
	if err != nil {
		klog.ErrorS(err, "Could not init kubernetes in-cluster client")
		os.Exit(1)
	}
	// initialize in-memory client
	inClusterClient = c
	return inClusterClient
}

func Client(request *http.Request) (karmadaclientset.Interface, error) {
	if !isInitialized() {
		return nil, fmt.Errorf("client package not initialized")
	}
	return clientFromRequest(request)
}

func InClusterKarmadaClient() karmadaclientset.Interface {
	if !isKarmdaInitialized() {
		return nil
	}
	if inClusterKarmadaClient != nil {
		return inClusterKarmadaClient
	}
	// init on-demand only
	c, err := karmadaclientset.NewForConfig(karmadaBaseConfig)
	if err != nil {
		klog.ErrorS(err, "Could not init kubernetes in-cluster client")
		os.Exit(1)
	}
	// initialize in-memory client
	inClusterKarmadaClient = c
	return inClusterKarmadaClient
}
