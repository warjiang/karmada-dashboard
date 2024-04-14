package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"warjiang/karmada-dashboard/client"
)

func main() {
	token := `eyJhbGciOiJSUzI1NiIsImtpZCI6Ik5xbkxVNWd1NlhTYVNzNEZIMGZjY1phT184UGx6TkRlb2NRQnlVWnFOVHcifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJrYXJtYWRhLXN5c3RlbSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJrYXJtYWRhLWRhc2hib2FyZC1zZWNyZXQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoia2FybWFkYS1kYXNoYm9hcmQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiIxNTQ5ZmQxMC1jM2I2LTRmOWUtYTk0ZC1kNjExMGZlYzNlZTUiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6a2FybWFkYS1zeXN0ZW06a2FybWFkYS1kYXNoYm9hcmQifQ.dHvPVFOWlD4E-KlTLEAvvTtB2mnQJIQroYdFhLSed_PuAqVNURhip65SKNbAcms43RCBhoFU5en9WHAPAAajzes6q4RvWHRDldKG-Urnq6b-3jhSJlWHtgQJ_ak_l8h3JqzKcHLXT1O3RG75vufnrW_4KWuiObJpEBCVXtkfwqkfdewQA0fqYaDZCIXXQwMGs2Ne4vHGQzsOqAr3L6SWzgAghIxa9lIudOKtHbwZEjJYqFHjWgfq0ZJDrY9TNIMNzMksskFgzkvJGSu9WoNGla22v3SFxwFOVA0AQTsh2QBe7VfuziOi4oD0A3cXh9jojR0YOwHbcMinxCGqP4-mdg`
	client.Init(
		client.WithUserAgent("dashboard-auth"),
		client.WithKubeconfig("/Users/dingwenjiang/.kube/karmada.config"),
		client.WithInsecureTLSSkipVerify(true),
	)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/login", nil)
	client.SetAuthorizationHeader(req, token)
	karmadaClient, err := client.Client(req)
	if err != nil {
		panic(err)
	}
	version, err := karmadaClient.Discovery().ServerVersion()
	if err != nil {
		panic(err)
	}
	fmt.Println(version)
	fmt.Println(version.GoVersion)
	fmt.Println(version.GitVersion)
	fmt.Println(version.GitCommit)
}
