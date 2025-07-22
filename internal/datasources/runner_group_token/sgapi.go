package runnergrouptoken

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type apiResponseModel struct {
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

func getAPIToken(runnerGroupID string, apiBaseUrl string, apiKey string, orgName string) (apiResponse *apiResponseModel, err error) {
	url := apiBaseUrl + "/api/v1/orgs/" + orgName + "/api_token/"

	type reqBody struct {
		Regenerate    bool   `json:"regenerate"`
		RunnerGroupId string `json:"runnerGroupId"`
	}
	reqBodyValue := reqBody{
		Regenerate:    false,
		RunnerGroupId: fmt.Sprintf("/runnergroups/%s", runnerGroupID),
	}

	payload, err := json.Marshal(reqBodyValue)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", apiKey)

	reqResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer reqResp.Body.Close()

	reqRespBody, err := io.ReadAll(reqResp.Body)
	if err != nil {
		return nil, err
	}

	if reqResp.StatusCode != 200 {
		return nil, fmt.Errorf("datasources.runner_group_token.getAPIToken: Failed to fetch api token: %s", reqRespBody)
	}

	var respModel apiResponseModel
	err = json.Unmarshal(reqRespBody, &respModel)
	if err != nil {
		return nil, err
	}

	return &respModel, nil
}
