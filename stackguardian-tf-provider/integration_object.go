package stackguardian_tf_provider

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceStackGuardianIntegrationAPI() *schema.Resource {
	// Consider data sensitive if env variables is set to true.
	is_data_sensitive, _ := strconv.ParseBool(GetEnvOrDefault("API_DATA_IS_SENSITIVE", "false"))

	return &schema.Resource{
		Create: resourceresourceStackGuardianIntegrationAPICreate,
		Read:   resourceresourceStackGuardianIntegrationAPIRead,
		Update: resourceresourceStackGuardianIntegrationAPIUpdate,
		Delete: resourceresourceStackGuardianIntegrationAPIDelete,
		Exists: resourceresourceStackGuardianIntegrationAPIExists,

		Importer: &schema.ResourceImporter{
			State: resourceresourceStackGuardianIntegrationAPIImport,
		},

		Schema: map[string]*schema.Schema{
			// "wfgrp": &schema.Schema{
			// 	Type:        schema.TypeString,
			// 	Description: "WorkFlow Group Name",
			// 	Required:    true,
			// },
			"create_path": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Defaults to `path`. The API path that represents where to CREATE (POST) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object if the data contains the `id_attribute`.",
				Optional:    true,
			},
			"read_path": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Defaults to `path/{id}`. The API path that represents where to READ (GET) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object.",
				Optional:    true,
			},
			"update_path": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Defaults to `path/{id}`. The API path that represents where to UPDATE (PUT) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object.",
				Optional:    true,
			},
			"create_method": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Defaults to `create_method` set on the provider. Allows per-resource override of `create_method` (see `create_method` provider config documentation)",
				Optional:    true,
			},
			"read_method": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Defaults to `read_method` set on the provider. Allows per-resource override of `read_method` (see `read_method` provider config documentation)",
				Optional:    true,
			},
			"update_method": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Defaults to `update_method` set on the provider. Allows per-resource override of `update_method` (see `update_method` provider config documentation)",
				Optional:    true,
			},
			"destroy_method": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Defaults to `destroy_method` set on the provider. Allows per-resource override of `destroy_method` (see `destroy_method` provider config documentation)",
				Optional:    true,
			},
			"destroy_path": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Defaults to `path/{id}`. The API path that represents where to DESTROY (DELETE) objects of this type on the API server. The string `{id}` will be replaced with the terraform ID of the object.",
				Optional:    true,
			},
			"destroy_data": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Whether to use the data object as the body for the delete request.",
				Optional:    true,
			},
			"id_attribute": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Defaults to `id_attribute` set on the provider. Allows per-resource override of `id_attribute` (see `id_attribute` provider config documentation)",
				Optional:    true,
			},
			"object_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Defaults to the id learned by the provider during normal operations and `id_attribute`. Allows you to set the id manually. This is used in conjunction with the `*_path` attributes.",
				Optional:    true,
			},
			"data": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Valid JSON data that this provider will manage with the API server.",
				Required:    true,
				Sensitive:   is_data_sensitive,
			},
			"debug": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Whether to emit verbose debug output while working with the API object on the server.",
				Optional:    true,
			},
			"read_search": &schema.Schema{
				Type:        schema.TypeMap,
				Description: "Custom search for `read_path`. This map will take `search_key`, `search_value`, `results_key` and `query_string` (see datasource config documentation)",
				Optional:    true,
			},
			"api_data": &schema.Schema{
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "After data from the API server is read, this map will include k/v pairs usable in other terraform resources as readable objects. Currently the value is the golang fmt package's representation of the value (simple primitives are set as expected, but complex types like arrays and maps contain golang formatting).",
				Computed:    true,
			},
			"api_response": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The raw body of the HTTP response from the last read of the object.",
				Computed:    true,
			},
			"create_response": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The raw body of the HTTP response returned when creating the object.",
				Computed:    true,
			},
			"force_new": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				ForceNew:    true,
				Description: "Any changes to these values will result in recreating the resource instead of updating.",
			},
		}, /* End schema */

	}
}

/*
Since there is nothing in the ResourceData structure other

	than the "id" passed on the command line, we have to use an opinionated
	view of the API paths to figure out how to read that object
	from the API
*/
func resourceresourceStackGuardianIntegrationAPIImport(d *schema.ResourceData, meta interface{}) (imported []*schema.ResourceData, err error) {
	input := d.Id()

	hasTrailingSlash := strings.LastIndex(input, "/") == len(input)-1
	var n int
	if hasTrailingSlash {
		n = strings.LastIndex(input[0:len(input)-1], "/")
	} else {
		n = strings.LastIndex(input, "/")
	}

	if n == -1 {
		return imported, fmt.Errorf("Invalid path to import api_object '%s'. Must be /<full path from server root>/<object id>", input)
	}

	path := input[0:n]
	d.Set("path", path)

	var id string
	if hasTrailingSlash {
		id = input[n+1 : len(input)-1]
	} else {
		id = input[n+1 : len(input)]
	}

	d.Set("data", fmt.Sprintf(`{ "id": "%s" }`, id))
	d.SetId(id)

	/* Troubleshooting is hard enough. Emit log messages so TF_LOG
	   has useful information in case an import isn't working */
	d.Set("debug", true)

	obj, err := make_api_object_stack(d, meta)
	if err != nil {
		return imported, err
	}
	log.Printf("resource_api_object.go: Import routine called. Object built:\n%s\n", obj.toString())

	err = obj.read_object()
	if err == nil {
		set_resource_state(obj, d)
		/* Data that we set in the state above must be passed along
		   as an item in the stack of imported data */
		imported = append(imported, d)
	}

	return imported, err
}

func resourceresourceStackGuardianIntegrationAPICreate(d *schema.ResourceData, meta interface{}) error {
	obj, err := make_api_object_integration(d, meta)
	if err != nil {
		return err
	}
	log.Printf("resource_api_object.go: Create routine called. Object built:\n%s\n", obj.toString())

	err = obj.create_object()
	if err == nil {
		/* Setting terraform ID tells terraform the object was created or it exists */
		d.SetId(obj.ResourceName)
		set_resource_state(obj, d)
		/* Only set during create for APIs that don't return sensitive data on subsequent retrieval */
		d.Set("create_response", obj.api_response)
	}
	return err
}

func resourceresourceStackGuardianIntegrationAPIRead(d *schema.ResourceData, meta interface{}) error {
	obj, err := make_api_object_integration(d, meta)
	if err != nil {
		return err
	}
	log.Printf("resource_api_object.go: Read routine called. Object built:\n%s\n", obj.toString())

	err = obj.read_object()
	if err == nil {
		/* Setting terraform ID tells terraform the object was created or it exists */
		log.Printf("resource_api_object.go: Read resource. Returned id is '%s'\n", obj.ResourceName)
		d.SetId(obj.ResourceName)
		set_resource_state(obj, d)
	}
	return err
}

func resourceresourceStackGuardianIntegrationAPIUpdate(d *schema.ResourceData, meta interface{}) error {
	obj, err := make_api_object_integration(d, meta)
	if err != nil {
		return err
	}

	log.Printf("resource_api_object.go: Update routine called. Object built:\n%s\n", obj.toString())

	err = obj.update_object()
	if err == nil {
		set_resource_state(obj, d)
	}
	return err
}

func resourceresourceStackGuardianIntegrationAPIDelete(d *schema.ResourceData, meta interface{}) error {
	obj, err := make_api_object_integration(d, meta)
	if err != nil {
		return err
	}
	log.Printf("resource_api_object.go: Delete routine called. Object built:\n%s\n", obj.toString())

	err = obj.delete_object()
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			/* 404 means it doesn't exist. Call that good enough */
			err = nil
		}
	}
	return err
}

func resourceresourceStackGuardianIntegrationAPIExists(d *schema.ResourceData, meta interface{}) (exists bool, err error) {
	obj, err := make_api_object_integration(d, meta)
	if err != nil {
		return exists, err
	}
	log.Printf("resource_api_object.go: Exists routine called. Object built: %s\n", obj.toString())

	/* Assume all errors indicate the object just doesn't exist.
	This may not be a good assumption... */
	err = obj.read_object()
	if err == nil {
		exists = true
	}
	return exists, err
}

/*
Simple helper routine to build an api_object struct

	for the various calls terraform will use. Unfortunately,
	terraform cannot just reuse objects, so each CRUD operation
	results in a new object created
*/
func make_api_object_integration(d *schema.ResourceData, meta interface{}) (*api_object, error) {
	opts, err := buildApiObjectIntegrationOpts(d)
	if err != nil {
		return nil, err
	}

	obj, err := NewAPIObject(meta.(*api_client), opts)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func buildApiObjectIntegrationOpts(d *schema.ResourceData) (*apiObjectOpts, error) {
	// var resultPath string
	// resultPath = "/wfgrps/" + d.Get("wfgrp").(string) + "/stacks/"

	opts := &apiObjectOpts{
		// path: d.Get("path").(string),
		path: "",
	}

	/* Allow user to override provider-level id_attribute */
	if v, ok := d.GetOk("id_attribute"); ok {
		opts.id_attribute = v.(string)
	}

	/* Allow user to specify the ID manually */
	if v, ok := d.GetOk("object_id"); ok {
		opts.ResourceName = v.(string)
	} else {
		/* If not specified, see if terraform has an ID */
		opts.ResourceName = d.Id()
	}

	log.Printf("common.go: make_api_object routine called for id '%s'\n", opts.ResourceName)

	log.Printf("create_path: %s", d.Get("create_path"))
	if v, ok := d.GetOk("create_path"); ok {
		opts.post_path = v.(string)
	}
	if v, ok := d.GetOk("read_path"); ok {
		opts.get_path = v.(string)
	}
	if v, ok := d.GetOk("update_path"); ok {
		opts.put_path = v.(string)
	}
	if v, ok := d.GetOk("create_method"); ok {
		opts.create_method = v.(string)
	}
	if v, ok := d.GetOk("read_method"); ok {
		opts.read_method = v.(string)
	}
	if v, ok := d.GetOk("update_method"); ok {
		opts.update_method = v.(string)
	}
	if v, ok := d.GetOk("destroy_method"); ok {
		opts.destroy_method = v.(string)
	}
	if v, ok := d.GetOk("destroy_path"); ok {
		opts.delete_path = v.(string)
	}
	if v, ok := d.GetOk("destroy_data"); ok {
		opts.delete_data = v.(bool)
	}

	read_search := expandReadSearch(d.Get("read_search").(map[string]interface{}))
	opts.read_search = read_search

	opts.data = d.Get("data").(string)
	opts.debug = true
	return opts, nil
}
