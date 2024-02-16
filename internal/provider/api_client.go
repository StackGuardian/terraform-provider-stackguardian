// SPDX-License-Identifier: Apache-2.0
// Copyright 2015 Mastercard
// Copyright 2023 StackGuardian

package provider

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type apiClientOpt struct {
	api_uri      string
	org_name     string
	headers      map[string]string
	id_attribute string
}

type api_client struct {
	http_client  *http.Client
	api_uri      string
	org_name     string
	headers      map[string]string
	id_attribute string
}

// Make a new api client
func NewAPIClient(opt *apiClientOpt) (*api_client, error) {
	if true {
		log.Printf("api_client.go: Constructing debug api_client\n")
	}

	if opt.api_uri == "" {
		return nil, errors.New("uri must be set to construct an API client")
	} else {
		opt.api_uri = opt.api_uri + "orgs/" + opt.org_name + "/"
	}

	/* Sane default */
	if opt.id_attribute == "" {
		opt.id_attribute = "ResourceName"
	}

	/* Remove any trailing slashes since we will append
	   to this URL with our own root-prefixed location */
	opt.api_uri = strings.TrimSuffix(opt.api_uri, "/")

	// opt.create_method = "POST"

	// opt.read_method = "GET"

	// opt.update_method = "PATCH"

	// opt.destroy_method = "DELETE"

	var cookieJar http.CookieJar

	client := api_client{
		http_client: &http.Client{
			Jar: cookieJar,
		},
		api_uri:      opt.api_uri,
		headers:      opt.headers,
		id_attribute: opt.id_attribute,
	}

	if true {
		log.Printf("api_client.go: Constructed object:\n%s", client.toString())
	}
	return &client, nil
}

// Convert the important bits about this object to string representation
// This is useful for debugging.
func (obj *api_client) toString() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("api_uri: %s\n", obj.api_uri))
	buffer.WriteString(fmt.Sprintf("org_name: %s\n", obj.org_name))
	buffer.WriteString(fmt.Sprintf("id_attribute: %s\n", obj.id_attribute))
	buffer.WriteString(fmt.Sprintf("write_returns_object: %t\n", true))
	buffer.WriteString("headers:\n")
	for k, v := range obj.headers {
		buffer.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
	}
	return buffer.String()
}

/*
Helper function that handles sending/receiving and handling

	of HTTP data in and out.
*/
func (client *api_client) send_request(method string, path string, data string) (string, error) {
	full_uri := client.api_uri + path
	var req *http.Request
	var err error

	if true {
		log.Printf("api_client.go: method='%s', path='%s', full uri (derived)='%s', data='%s'\n", method, path, full_uri, data)
	}

	buffer := bytes.NewBuffer([]byte(data))

	if data == "" {
		req, err = http.NewRequest(method, full_uri, nil)
	} else {
		req, err = http.NewRequest(method, full_uri, buffer)

		/* Default of application/json, but allow headers array to overwrite later */

		req.Header.Set("Content-Type", "application/json")

	}

	if err != nil {
		log.Fatal(err)
		return "", err
	}

	if true {
		log.Printf("api_client.go: Sending HTTP request to %s...\n", req.URL)
	}

	/* Allow for tokens or other pre-created secrets */
	if len(client.headers) > 0 {
		for n, v := range client.headers {
			req.Header.Set(n, v)
		}
	}

	if true {
		log.Printf("api_client.go: Request headers:\n")
		for name, headers := range req.Header {
			for _, h := range headers {
				log.Printf("api_client.go:   %v: %v", name, h)
			}
		}

		log.Printf("api_client.go: BODY:\n")
		body := "<none>"
		if req.Body != nil {
			body = string(data)
		}
		log.Printf("%s\n", body)
	}

	resp, err := client.http_client.Do(req)

	if err != nil {
		//log.Printf("api_client.go: Error detected: %s\n", err)
		return "", err
	}

	if true {
		log.Printf("api_client.go: Response code: %d\n", resp.StatusCode)
		log.Printf("api_client.go: Response headers:\n")
		for name, headers := range resp.Header {
			for _, h := range headers {
				log.Printf("api_client.go:   %v: %v", name, h)
			}
		}
	}

	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err2 != nil {
		return "", err2
	}
	body := strings.TrimPrefix(string(bodyBytes), "")
	if true {
		log.Printf("api_client.go: BODY:\n%s\n", body)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, fmt.Errorf("unexpected response code '%d': %s", resp.StatusCode, body)
	}

	return body, nil

}
