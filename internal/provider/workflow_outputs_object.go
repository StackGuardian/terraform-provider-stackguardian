package provider

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/mapstructure"
)

func dataSourceStackGuardianWorkflowOutputsAPI() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStackGuardianWorkflowOutputsAPIRead,

		Schema: map[string]*schema.Schema{
			"wfgrp": {
				Type:        schema.TypeString,
				Description: "WorkFlow Group Name",
				Required:    true,
			},
			"stack": {
				Type:        schema.TypeString,
				Description: "Stack name",
				Optional:    true,
			},
			"wf": {
				Type:        schema.TypeString,
				Description: "WorkFlow Name",
				Required:    true,
			},
			"msg": {
				Type:        schema.TypeString,
				Description: "Message from API",
				Computed:    true,
			},
			"data": { // TODO: rename it to "outputs"
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "After data from the API server is read, this map will include k/v pairs usable in other terraform resources as readable objects. Currently the value is the golang fmt package's representation of the value (simple primitives are set as expected, but complex types like arrays and maps contain golang formatting).",
				Computed:    true,
			},
			"outputs_json": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outputs_str": {
				Type: schema.TypeMap,
				//Elem:     &schema.Schema{Type: schema.TypeString},
				Elem:     schema.TypeString,
				Computed: true,
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

func dataSourceStackGuardianWorkflowOutputsAPIRead(d *schema.ResourceData, meta interface{}) error {
	// DEBUG: debugProcess()
	log.Printf("DEBUG: dataSourceStackGuardianWorkflowOutputsAPIRead: ...")

	var resultPath string
	stack, stackExists := d.Get("stack").(string)
	if stackExists && stack != "" {
		resultPath = "/wfgrps/" + d.Get("wfgrp").(string) + "/stacks/" + d.Get("stack").(string) + "/wfs/" + d.Get("wf").(string) + "/outputs/"
	} else {
		resultPath = "/wfgrps/" + d.Get("wfgrp").(string) + "/wfs/" + d.Get("wf").(string) + "/outputs/"
	}

	client := meta.(*api_client)

	opts := &apiObjectOpts{
		get_path: resultPath,
	}

	obj, err := NewAPIObject(client, opts)
	if err != nil {
		return err
	}

	if v, ok := d.GetOk("id_attribute"); ok {
		opts.id_attribute = v.(string)
	}
	obj.ResourceName = fmt.Sprintf("%s/%s", d.Get("wfgrp").(string), d.Get("wf").(string))

	/* Back to terraform-specific stuff. Create an api_object with the ID and refresh it object */

	log.Printf("workflow_outputs_object.go: Attempting to construct api_object to refresh data")

	log.Printf("DEBUG: d.SetId(obj.ResourceName) with obj.ResourceName=%v", obj.ResourceName)
	d.SetId(obj.ResourceName)

	err = obj.read_object()
	// TODO: handle error
	if err == nil {
		/* Setting terraform ID tells terraform the object was created or it exists */
		log.Printf("workflow_outputs_object.go: Data resource. Returned id is '%s'\n", obj.ResourceName)

		set_resource_state(obj, d)

		// --- Storing result in the computed field `msg`
		if _, ok := obj.api_data["msg"]; !ok {
			obj.api_data["msg"] = "No message from API"
		}
		ds_msg := fmt.Sprintf("%v", obj.api_data["msg"])
		d.Set("msg", ds_msg)
		log.Printf("workflow_outputs_object.go: message from API: %s", ds_msg)

		var outputs_api_raw outputsAPIResponse
		err := json.Unmarshal([]byte(obj.api_response), &outputs_api_raw)
		if err != nil {
			msg := "failure to Unmarshal obj.api_response"
			log.Printf("ERROR: " + msg)
			return fmt.Errorf("workflow_outputs_object.go: " + msg)
		}

		outputs_str, err := exportOutputsTerraformBasic(outputs_api_raw.Data.Outputs)
		if err != nil {
			return fmt.Errorf("workflow_outputs_object.go: failure to export outputs as map of strings: %w", err)
		}
		d.Set("outputs_str", outputs_str)

		outputs_json, err := exportOutputsJSON(outputs_api_raw.Data.Outputs)
		if err != nil {
			return fmt.Errorf("workflow_outputs_object.go: failure to export outputs as string of JSON: %w", err)
		}
		d.Set("outputs_json", outputs_json)

	}

	log.Printf("DEBUG: dataSourceStackGuardianWorkflowOutputsAPIRead: DONE")
	return err
}

type outputsAPIResponse struct {
	Msg  string `json:"msg"`
	Data struct {
		Outputs json.RawMessage `json:"outputs"` // Might be `map[string]interface{}` or `string`
	} `json:"data"`
}

/// Outputs JSON //////////////////////////////////////////////////////////////////////////////////////////////////////

func exportOutputsJSON(outputs_raw json.RawMessage) (string, error) {
	log.Printf("DEBUG: arg/outputs_raw: %+v", string(outputs_raw))

	outputs_json_raw, err := outputs_raw.MarshalJSON()
	if err != nil {
		errmsg := fmt.Errorf("failure to Unmarshal outputs_raw as JSON: %q : %w", string(outputs_raw), err)
		log.Printf("WARNING: " + errmsg.Error())
		return "{}", errmsg
	}

	return string(outputs_json_raw), nil
}

/// Outputs Basic /////////////////////////////////////////////////////////////////////////////////////////////////////

type outputTF struct {
	Type      string      `mapstructure:"type"`
	Value     interface{} `mapstructure:"value"`
	Sensitive bool        `mapstructure:"sensitive"`
}

func exportOutputsTerraformBasic(outputs_raw json.RawMessage) (map[string]string, error) {
	log.Printf("DEBUG: arg/outputs_raw: %+v", string(outputs_raw))
	outputs_str := make(map[string]string, 0)

	//outputs_map_raw, ok := outputs_raw.(map[string]interface{})
	var outputs_map_raw map[string]interface{}
	err := json.Unmarshal(outputs_raw, &outputs_map_raw)
	if err != nil {
		errmsg := fmt.Errorf("failure to cast as `map[string]interface{}`: %+v : %w", outputs_raw, err)
		log.Printf("WARNING: " + errmsg.Error())
		return nil, errmsg
	}

	for output_key, output_value := range outputs_map_raw {

		// Unmarhalling strings here directly is not supposed to be needed
		output_str, ok := output_value.(string)
		if ok {
			outputs_str[output_key] = output_str
		} else {
			msg := fmt.Sprintf("(expected) failure to cast as string: %+v", output_value)
			log.Printf("WARNING: " + msg)
			//return nil, fmt.Errorf(msg)
		}

		var output_tf_map outputTF
		err := mapstructure.Decode(output_value, &output_tf_map)
		if err != nil {
			err_ := fmt.Errorf("failure to cast as struct of TF output: %+v: %w", output_value, err)
			log.Printf("ERROR: " + err_.Error())
			return nil, err_
		}
		switch output_tf_map.Type {
		// TODO: check that unmarshalling is not needed, use switch output_value.(type)
		case "string":
			output_str, ok := output_tf_map.Value.(string)
			if !ok {
				msg := fmt.Sprintf("failure to cast as string despite being typed as `string`: %+v", output_tf_map.Value)
				log.Printf("ERROR: " + msg)
				return nil, fmt.Errorf(msg)
			}
			outputs_str[output_key] = output_str
		case "number":
			switch output_tf_map.Value.(type) {
			case int64:
				output_str, integer_ok := output_tf_map.Value.(int64)
				if !integer_ok {
					msg := fmt.Sprintf("failure to cast as number(int) despite being typed as `number`: %+v", output_tf_map.Value)
					log.Printf("ERROR: " + msg)
					return nil, fmt.Errorf(msg)
				}
				outputs_str[output_key] = fmt.Sprintf("%v", output_str)
			case float64:
				output_str, integer_ok := output_tf_map.Value.(float64)
				if !integer_ok {
					msg := fmt.Sprintf("failure to cast as number(float) despite being typed as `number`: %+v", output_tf_map.Value)
					log.Printf("ERROR: " + msg)
					return nil, fmt.Errorf(msg)
				}
				outputs_str[output_key] = fmt.Sprintf("%v", output_str)
			}
		case "bool":
			output_str, ok := output_tf_map.Value.(bool)
			if !ok {
				msg := fmt.Sprintf("failure to cast as bool despite being typed as `bool`: %+v", output_tf_map.Value)
				log.Printf("ERROR: " + msg)
				return nil, fmt.Errorf(msg)
			}
			outputs_str[output_key] = fmt.Sprintf("%v", output_str)
		default:
			msg := fmt.Sprintf("output Type is not implemented: `bool`: %+v", output_tf_map.Value)
			log.Printf("ERROR: " + msg)
			return nil, fmt.Errorf(msg)
		}

	}

	return outputs_str, nil
}
