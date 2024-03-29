package provider

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceStackGuardianWorkflowGroupAPI() *schema.Resource {
	// Consider data sensitive if env variables is set to true.
	is_data_sensitive, _ := strconv.ParseBool(GetEnvOrDefault("API_DATA_IS_SENSITIVE", "false"))

	return &schema.Resource{
		Create: resourceStackGuardianWorkflowGroupAPICreate,
		Read:   resourceStackGuardianWorkflowGroupAPIRead,
		Update: resourceStackGuardianWorkflowGroupAPIUpdate,
		Delete: resourceStackGuardianWorkflowGroupAPIDelete,
		Exists: resourceStackGuardianWorkflowGroupAPIExists,

		Importer: &schema.ResourceImporter{
			State: resourceStackGuardianWorkflowGroupAPIImport,
		},

		Schema: map[string]*schema.Schema{
			"data": {
				Type:        schema.TypeString,
				Description: "Valid JSON data that this provider will manage with the API server. Please refer to the API Docs: https://docs.stackguardian.io/api#tag/Workflow-Groups",
				Required:    true,
				Sensitive:   is_data_sensitive,
			},
			"api_data": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "After data from the API server is read, this map will include k/v pairs usable in other terraform resources as readable objects. Currently the value is the golang fmt package's representation of the value (simple primitives are set as expected, but complex types like arrays and maps contain golang formatting).",
				Computed:    true,
			},
			"api_response": {
				Type:        schema.TypeString,
				Description: "The raw body of the HTTP response from the last read of the object.",
				Computed:    true,
			},
		},
	}
}

/*
Since there is nothing in the ResourceData structure other

	than the "id" passed on the command line, we have to use an opinionated
	view of the API paths to figure out how to read that object
	from the API
*/
func resourceStackGuardianWorkflowGroupAPIImport(d *schema.ResourceData, meta interface{}) (imported []*schema.ResourceData, err error) {
	input := d.Id()

	hasTrailingSlash := strings.LastIndex(input, "/") == len(input)-1
	var n int
	if hasTrailingSlash {
		n = strings.LastIndex(input[0:len(input)-1], "/")
	} else {
		n = strings.LastIndex(input, "/")
	}

	if n == -1 {
		return imported, fmt.Errorf("invalid path to import api_object '%s'. Must be /<full path from server root>/<object id>", input)
	}

	var id string
	if hasTrailingSlash {
		id = input[n+1 : len(input)-1]
	} else {
		id = input[n+1:]
	}

	d.Set("data", fmt.Sprintf(`{ "id": "%s" }`, id))
	d.SetId(id)

	obj, err := make_api_object_WorkflowGroup(d, meta)
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

func resourceStackGuardianWorkflowGroupAPICreate(d *schema.ResourceData, meta interface{}) error {
	obj, err := make_api_object_WorkflowGroup(d, meta)
	if err != nil {
		return err
	}
	log.Printf("resource_api_object.go: Create routine called. Object built:\n%s\n", obj.toString())

	err = obj.create_object()
	if err == nil {
		/* Setting terraform ID tells terraform the object was created or it exists */
		d.SetId(obj.ResourceName)
		set_resource_state(obj, d)
	}
	return err
}

func resourceStackGuardianWorkflowGroupAPIRead(d *schema.ResourceData, meta interface{}) error {
	obj, err := make_api_object_WorkflowGroup(d, meta)
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

func resourceStackGuardianWorkflowGroupAPIUpdate(d *schema.ResourceData, meta interface{}) error {
	obj, err := make_api_object_WorkflowGroup(d, meta)
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

func resourceStackGuardianWorkflowGroupAPIDelete(d *schema.ResourceData, meta interface{}) error {
	obj, err := make_api_object_WorkflowGroup(d, meta)
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

func resourceStackGuardianWorkflowGroupAPIExists(d *schema.ResourceData, meta interface{}) (exists bool, err error) {
	obj, err := make_api_object_WorkflowGroup(d, meta)
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
func make_api_object_WorkflowGroup(d *schema.ResourceData, meta interface{}) (*api_object, error) {
	opts, err := buildApiObjectOpts_WorkflowGroup(d)
	if err != nil {
		return nil, err
	}

	obj, err := NewAPIObject(meta.(*api_client), opts)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func buildApiObjectOpts_WorkflowGroup(d *schema.ResourceData) (*apiObjectOpts, error) {
	resultPath := "/wfgrps/"

	opts := &apiObjectOpts{
		path: resultPath,
	}

	opts.ResourceName = d.Id()

	log.Printf("common.go: make_api_object routine called for id '%s'\n", opts.ResourceName)

	opts.data = d.Get("data").(string)
	opts.debug = true
	return opts, nil
}
