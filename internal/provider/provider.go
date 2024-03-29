package provider

import (
	"log"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const default_api_uri = "https://api.app.stackguardian.io/api/v1/"

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_uri": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("STACKGUARDIAN_API_URI", default_api_uri),
				Description: "Api Uri to use as base for StackGuardian API",
			},
			"org_name": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("STACKGUARDIAN_ORG_NAME", nil),
				Description: "Organization Name created in STACKGUARDIAN",
			},
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("STACKGUARDIAN_API_KEY", nil),
				Description: "Api Key to Authenticate to StackGuardian API",
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
	opt := &apiClientOpt{
		api_uri:  d.Get("api_uri").(string),
		org_name: d.Get("org_name").(string),
		headers: map[string]string{
			"Authorization": "apikey " + d.Get("api_key").(string),
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
