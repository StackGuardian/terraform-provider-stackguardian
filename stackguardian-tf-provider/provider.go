package stackguardian_tf_provider

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_uri": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("STACKGUARDIAN_API_URI", nil),
				Description: "URI of the StackGuradian API endpoint. This serves as the base of all requests.",
			},
			"org_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("STACKGUARDIAN_ORG_NAME", nil),
				Description: "Organization Name created in STACKGUARDIAN",
			},
			"api_key": &schema.Schema{
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
			"stackguardian_tf_provider_workflow": resourceStackGuardianWorkflowAPI(),
			"stackguardian_tf_provider_stack":    resourceStackGuardianStackAPI(),
			"stackguardian_tf_provider_policy":   resourceStackGuardianPolicyAPI(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"stackguardian_tf_provider_workflow": dataSourceStackGuardianAPI(),
			"stackguardian_tf_provider_stack":    dataSourceStackGuardianAPI(),
			"stackguardian_tf_provider_policy":   dataSourceStackGuardianAPI(),
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
