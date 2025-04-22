package customTypes

import (
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
)

type ProviderInfo struct {
	ApiBaseURL string
	ApiKey     string
	Org_name   string
	Client     *sgclient.Client
}
