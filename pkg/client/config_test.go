package client

import (
	"fmt"
	"github.com/karmada-io/dashboard/pkg/environment"
	"testing"
)

func TestGenerateKubeconfigForTerminal(t *testing.T) {
	cases := []struct {
		Server                   string
		InsecureSkipTLSVerify    bool
		CertificateAuthorityData string
		Token                    string
		expected                 string
	}{
		{
			Server:                   "https://172.18.0.2:5443",
			InsecureSkipTLSVerify:    true,
			CertificateAuthorityData: "",
			Token:                    "xxx token",
			expected: `apiVersion: v1
kind: Config
clusters:
- name: karmada-apiserver
  cluster:
    server: https://172.18.0.2:5443
    insecure-skip-tls-verify: true
users:
- name: karmada-apiserver
  user:
    token: xxx token
contexts:
- name: karmada-apiserver
  context:
    cluster: karmada-apiserver
    user: karmada-apiserver
current-context: karmada-apiserver`,
		},
	}

	for _, c := range cases {
		expectedData, _ := UnifyYaml(c.expected)
		genCfg, genErr := GenerateKubeconfigForTerminal(KubeconfigOpts{
			Server:                   c.Server,
			InsecureSkipTLSVerify:    c.InsecureSkipTLSVerify,
			CertificateAuthorityData: "",
			Token:                    c.Token,
		})
		genCfgData, _ := UnifyYaml(genCfg)

		if genErr != nil {
			t.Error("Cannot generate kubeconfig", genErr)
		}
		if genCfgData != expectedData {
			t.Error("Expected", expectedData, "got", genCfgData)
		}
	}
}

func TestWriteKubeconfig(t *testing.T) {
	InitKubeConfig(
		WithUserAgent(environment.UserAgent()),
		WithKubeconfig("/Users/dingwenjiang/.kube/karmada.config"),
		WithKubeContext("karmada-host"),
		WithInsecureTLSSkipVerify(false),
	)
	err := WriteKubeconfig("karmada-system", "demo-ttyd-6458b64dc6-k8qqn", "ttyd", KubeconfigOpts{
		Server:                   "https://172.18.0.2:5443",
		InsecureSkipTLSVerify:    true,
		CertificateAuthorityData: "",
		Token:                    "eyJhbGciOiJSUzI1NiIsImtpZCI6IlMzcU5LRzlGVlBWdDdJUUFLOEpVYmFBSWRZanlQbGNnYmlpNUpiQ1k5ejAifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJrYXJtYWRhLXN5c3RlbSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJrYXJtYWRhLWRhc2hib2FyZC1zZWNyZXQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoia2FybWFkYS1kYXNoYm9hcmQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiI0YWM4NzAwOC0zNDI2LTRjYWUtOGYzYy0xNDFlMDNhMGUzNGIiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6a2FybWFkYS1zeXN0ZW06a2FybWFkYS1kYXNoYm9hcmQifQ.NN5Zvm2DnlwzLvUtToLHEi9FFQ-r9Di-vY7Axf_7NrAhXyTABkcZsmFMTM8vBG-wrqEiz_dmuLcLOpgM1jHh29QIKTIlQh77ZP7w4p59o4pdWran8ekis7WmtA2CWc_KLPcPt1AUorT_NgYUTxcHhOwrV5JHtznh6whIfRgtp7a-NT7gRi6DnoRjMhy_cSgpRSfwce9iwBIKNvSQdEQT3SjJNQFf4l5BRv0QCmBRwxy5G2IuhZh2N_Tgx_Jvn3N45AoQMoS4OUTz5alSqMW24EyEuukIqJC1UUkr8p7wSI2Z3jf3VQIhMiBo4nYveHL9pKtH6FFjazV9yasqZitlGg",
	})
	if err != nil {
		t.Error("Cannot write kubeconfig", err)
	}
}

func TestParseToken(t *testing.T) {
	token, err := ParseToken("eyJhbGciOiJSUzI1NiIsImtpZCI6IlMzcU5LRzlGVlBWdDdJUUFLOEpVYmFBSWRZanlQbGNnYmlpNUpiQ1k5ejAifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJrYXJtYWRhLXN5c3RlbSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJrYXJtYWRhLWRhc2hib2FyZC1zZWNyZXQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoia2FybWFkYS1kYXNoYm9hcmQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiI0YWM4NzAwOC0zNDI2LTRjYWUtOGYzYy0xNDFlMDNhMGUzNGIiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6a2FybWFkYS1zeXN0ZW06a2FybWFkYS1kYXNoYm9hcmQifQ.NN5Zvm2DnlwzLvUtToLHEi9FFQ-r9Di-vY7Axf_7NrAhXyTABkcZsmFMTM8vBG-wrqEiz_dmuLcLOpgM1jHh29QIKTIlQh77ZP7w4p59o4pdWran8ekis7WmtA2CWc_KLPcPt1AUorT_NgYUTxcHhOwrV5JHtznh6whIfRgtp7a-NT7gRi6DnoRjMhy_cSgpRSfwce9iwBIKNvSQdEQT3SjJNQFf4l5BRv0QCmBRwxy5G2IuhZh2N_Tgx_Jvn3N45AoQMoS4OUTz5alSqMW24EyEuukIqJC1UUkr8p7wSI2Z3jf3VQIhMiBo4nYveHL9pKtH6FFjazV9yasqZitlGg")
	// kubernetes.io/serviceaccount/service-account.uid
	fmt.Println(token)
	fmt.Println(err)
}
