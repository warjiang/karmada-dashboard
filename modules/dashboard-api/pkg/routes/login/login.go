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

package login

import (
	"net/http"

	"warjiang/karmada-dashboard/client"
	v1 "warjiang/karmada-dashboard/dashboard-api/api/v1"
	"warjiang/karmada-dashboard/errors"
)

func login(spec *v1.LoginRequest, request *http.Request) (*v1.LoginResponse, int, error) {
	ensureAuthorizationHeader(spec, request)

	karmadaClient, err := client.KarmadaClient(request)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if _, err = karmadaClient.Discovery().ServerVersion(); err != nil {
		code, err := errors.HandleError(err)
		return nil, code, err
	}

	return &v1.LoginResponse{Token: spec.Token}, http.StatusOK, nil
}

func ensureAuthorizationHeader(spec *v1.LoginRequest, request *http.Request) {
	client.SetAuthorizationHeader(request, spec.Token)
}
