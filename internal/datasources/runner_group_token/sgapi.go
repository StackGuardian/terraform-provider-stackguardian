package runnergrouptoken

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type apiResponseModel struct {
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func getAPIToken(runnerGroupID string, apiBaseUrl string, apiKey string, orgName string) (apiResponse *apiResponseModel, err error) {
	tokenURL := apiBaseUrl + "/api/v1/orgs/" + url.PathEscape(orgName) + "/api_token/"

	type reqBody struct {
		Regenerate    bool   `json:"regenerate"`
		RunnerGroupId string `json:"runnerGroupId"`
	}
	reqBodyValue := reqBody{
		Regenerate:    false,
		RunnerGroupId: fmt.Sprintf("/runnergroups/%s", url.PathEscape(runnerGroupID)),
	}

	payload, err := json.Marshal(reqBodyValue)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", tokenURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", apiKey)

	reqResp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = reqResp.Body.Close() }()

	reqRespBody, err := io.ReadAll(io.LimitReader(reqResp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	if reqResp.StatusCode != 200 {
		return nil, fmt.Errorf("datasources.runner_group_token.getAPIToken: API returned status %d", reqResp.StatusCode)
	}

	var respModel apiResponseModel
	err = json.Unmarshal(reqRespBody, &respModel)
	if err != nil {
		return nil, err
	}

	return &respModel, nil
}
