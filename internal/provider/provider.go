package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
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
			/* Could only get terraform to recognize this resource if
			         the name began with the provider's name and had at least
				 one underscore. This is not documented anywhere I could find */
			"stackguardian_workflow":    resourceStackGuardianWorkflowAPI(),
			"stackguardian_stack":       resourceStackGuardianStackAPI(),
			"stackguardian_policy":      resourceStackGuardianPolicyAPI(),
			"stackguardian_integration": resourceStackGuardianIntegrationAPI(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"stackguardian_workflow":    dataSourceStackGuardianAPI(),
			"stackguardian_stack":       dataSourceStackGuardianAPI(),
			"stackguardian_policy":      dataSourceStackGuardianAPI(),
			"stackguardian_integration": dataSourceStackGuardianAPI(),
			"stackguardian_wf_output":   dataSourceStackGuardianWorkflowOutputsAPI(),
		},
		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {

	/* As "data-safe" as terraform says it is, you'd think
	   it would have already coaxed this to a slice FOR me */
	copy_keys := make([]string, 0)
	if i_copy_keys := d.Get("copy_keys"); i_copy_keys != nil {
		for _, v := range i_copy_keys.([]interface{}) {
			copy_keys = append(copy_keys, v.(string))
		}
	}

	// headers := make(map[string]string)
	// if i_headers := d.Get("headers"); i_headers != nil {
	// 	for k, v := range i_headers.(map[string]interface{}) {
	// 		headers[k] = v.(string)
	// 	}
	// }

	opt := &apiClientOpt{
		api_uri:  "https://api.app.stackguardian.io/api/v1/",
		org_name: d.Get("org_name").(string),
		headers: map[string]string{
			"Authorization": "apikey " + d.Get("api_key").(string),
		},
		id_attribute: "ResourceName",
	}

	client, err := NewAPIClient(opt)
	return client, err
}
