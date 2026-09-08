package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

const (
	StackguardianApiKey        = "STACKGUARDIAN_API_KEY"
	StackguardianApiUri        = "STACKGUARDIAN_API_URI"
	StackguardianOrgName       = "STACKGUARDIAN_ORG_NAME"
	TestAzureStorageAccessKey  = "TEST_AZURE_STORAGE_BACKEND_ACCESS_KEY"
	DefaultStackguardianApiUri = "https://api.app.stackguardian.io"
)

var (
	instance *Config
	once     sync.Once
)

type Config struct {
	ApiKey                string
	ApiUri                string
	OrgName               string
	AzureStorageAccessKey string
}

func Get() *Config {
	once.Do(func() {
		instance = load()
	})
	return instance
}

func load() *Config {
	v := viper.New()

	v.SetEnvPrefix("")
	v.AutomaticEnv()

	v.SetDefault(StackguardianApiUri, DefaultStackguardianApiUri)

	cfg := &Config{
		ApiKey:                v.GetString(StackguardianApiKey),
		ApiUri:                v.GetString(StackguardianApiUri),
		OrgName:               v.GetString(StackguardianOrgName),
		AzureStorageAccessKey: v.GetString(TestAzureStorageAccessKey),
	}

	return cfg
}

func (c *Config) FormatApiKey() string {
	return fmt.Sprintf("apikey %s", c.ApiKey)
}
