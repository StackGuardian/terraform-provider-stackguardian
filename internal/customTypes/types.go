package customTypes

import (
	sgclient "github.com/StackGuardian/sg-sdk-go/client"
)

type ProviderInfo struct {
	ApiBaseURL string
	ApiKey     string
	OrgName    string
	Client     *sgclient.Client
}
