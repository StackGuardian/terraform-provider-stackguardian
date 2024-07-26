package provider

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_uri": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Api Uri to set as prefix URL for StackGuardian API. Required if not using environment variable STACKGUARDIAN_API_URI",
			},
			"org_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization Name to use on StackGuardian API. Required if not using environment variable STACKGUARDIAN_API_KEY",
			},
			"api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Api Key to authenticate on StackGuardian API. Required if not using environment variable STACKGUARDIAN_ORG_NAME",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"stackguardian_workflow":        resourceStackGuardianWorkflowAPI(),
			"stackguardian_workflow_group":  resourceStackGuardianWorkflowGroupAPI(),
			"stackguardian_stack":           resourceStackGuardianStackAPI(),
			"stackguardian_policy":          resourceStackGuardianPolicyAPI(),
			"stackguardian_integration":     resourceStackGuardianIntegrationAPI(),
			"stackguardian_role":            resourceStackGuardianRoleAPI(),
			"stackguardian_connector_cloud": resourceStackGuardianConnectorCloudAPI(),
			"stackguardian_connector_vcs":   resourceStackGuardianConnectorVcsAPI(),
			"stackguardian_secret":          resourceStackGuardianSecretAPI(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"stackguardian_workflow":         dataSourceStackGuardianAPI(),
			"stackguardian_workflow_group":   dataSourceStackGuardianAPI(),
			"stackguardian_stack":            dataSourceStackGuardianAPI(),
			"stackguardian_policy":           dataSourceStackGuardianAPI(),
			"stackguardian_integration":      dataSourceStackGuardianAPI(),
			"stackguardian_workflow_outputs": dataSourceStackGuardianWorkflowOutputsAPI(),
			"stackguardian_role":             dataSourceStackGuardianAPI(),
			"stackguardian_connector_cloud":  dataSourceStackGuardianAPI(),
			"stackguardian_connector_vcs":    dataSourceStackGuardianAPI(),
			"stackguardian_secret":           dataSourceStackGuardianAPI(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {

	// Fetch backend api key from provider block, or the environment. If not found
	// return an error
	providerFieldsValue := map[string]string{}

	for providerField, envVar := range map[string]string{
		"api_uri":  "STACKGUARDIAN_API_URI",
		"api_key":  "STACKGUARDIAN_API_KEY",
		"org_name": "STACKGUARDIAN_ORG_NAME"} {

		if d.Get(providerField).(string) != "" {
			providerFieldsValue[providerField] = d.Get(providerField).(string)
		} else if os.Getenv(envVar) != "" {
			providerFieldsValue[providerField] = os.Getenv(envVar)
		} else {
			return nil, fmt.Errorf("define %s in the provider block or %s environment variable", providerField, envVar)
		}
	}

	opt := &apiClientOpt{
		api_uri:  providerFieldsValue["api_uri"],
		org_name: providerFieldsValue["org_name"],
		headers: map[string]string{
			"Authorization": "apikey " + providerFieldsValue["api_key"],
		},
		id_attribute: "ResourceName",
	}
	client, err := NewAPIClient(opt)
	return client, err
}

/// DEBUG /////////////////////////////////////////////////////////////////////////////////////////

func debugProcess() {
	_, found := os.LookupEnv("TF_LOG")
	if !found {
		return
	}

	pid_current := os.Getpid()
	pid_parent := os.Getppid()

	wait := time.Second * 15

	log.Printf("DEBUG: ParentPID = %d | CurrentPID = %d", pid_parent, pid_current)
	log.Printf("DEBUG: Continuing in %s", wait)
	time.Sleep(wait)
}
