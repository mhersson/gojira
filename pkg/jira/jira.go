/*
Copyright © 2021 Morten Hersson

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"gitlab.com/mhersson/gojira/pkg/types"
	"gitlab.com/mhersson/gojira/pkg/util"
)

func GetIssues(config types.Config, filter string) []types.Issue {
	url := config.JiraURL + "/rest/api/2/search"

	if filter == "" {
		filter = `assignee = ` + config.Username +
			` AND resolution = Unresolved order by priority, updated`
	} else {
		filter += " order by priority, updated"
	}

	payload := []byte(`{"jql": "` + filter + `",
		"startAt":0,
		"maxResults":50,
		"fields":[
		"summary",
		"status",
		"updated",
		"assignee",
		"issuetype",
		"priority"]
	}`)

	jsonResponse := new(struct {
		Issues []types.Issue `json:"issues"`
	})

	getJSONResponse(config, "POST", url, payload, jsonResponse)

	return jsonResponse.Issues
}

func GetTimesheet(config types.Config, date string, showEntireWeek bool) []types.Timesheet {
	url := config.JiraURL + "/rest/timesheet-gadget/1.0/raw-timesheet.json?startDate=" + date + "&endDate=" + date

	if showEntireWeek {
		// Date is already validated, so should be safe
		// to drop the error check here
		t, _ := time.Parse("2006-01-02", date)
		start, end := util.WeekStartEndDate(t.ISOWeek())
		url = config.JiraURL + "/rest/timesheet-gadget/1.0/raw-timesheet.json?startDate=" + start + "&endDate=" + end
	}

	jsonResponse := new(struct {
		Worklog []types.Timesheet `json:"worklog"`
	})

	getJSONResponse(config, http.MethodGet, url, nil, jsonResponse)

	return jsonResponse.Worklog
}

func GetValidProjectsAndIssueType(config types.Config) types.IssueCreateMeta {
	url := config.JiraURL + "/rest/api/2/issue/createmeta"

	jsonResponse := &types.IssueCreateMeta{}

	getJSONResponse(config, "GET", url, nil, jsonResponse)

	return *jsonResponse
}

func GetPriorities(config types.Config) []types.Priority {
	url := config.JiraURL + "/rest/api/2/priority"

	jsonResponse := &[]types.Priority{}

	getJSONResponse(config, "GET", url, nil, jsonResponse)

	return *jsonResponse
}

func GetIssueTypes(config types.Config) *[]types.IssueType {
	url := config.JiraURL + "/rest/api/2/issuetype"

	jsonResponse := &[]types.IssueType{}

	getJSONResponse(config, "GET", url, nil, jsonResponse)

	return jsonResponse
}

func GetIssue(config types.Config, key string) types.IssueDescription {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key)

	jsonResponse := &types.IssueDescription{}

	getJSONResponse(config, "GET", url, nil, jsonResponse)

	return *jsonResponse
}

func GetIssuesInEpic(config types.Config, key string) []types.Issue {
	url := config.JiraURL + "/rest/api/2/search?jql=cf[10500]=" + strings.ToUpper(key)

	jsonResponse := new(struct {
		Issues []types.Issue `json:"issues"`
	})

	getJSONResponse(config, "GET", url, nil, jsonResponse)

	return jsonResponse.Issues
}

func GetTransistions(config types.Config, key string) []types.Transition {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/transitions"

	jsonResponse := new(struct {
		Transitions []types.Transition `json:"transitions"`
	})

	getJSONResponse(config, "GET", url, nil, jsonResponse)

	return jsonResponse.Transitions
}

func GetComments(config types.Config, key string) []types.Comment {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/comment"

	jsonResponse := new(struct {
		Comments []types.Comment `json:"comments"`
	})

	getJSONResponse(config, "GET", url, nil, jsonResponse)

	return jsonResponse.Comments
}

func GetWorklogs(config types.Config, key string) []types.Worklog {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/worklog"

	jsonResponse := new(struct {
		Worklogs []types.Worklog `json:"worklogs"`
	})

	getJSONResponse(config, "GET", url, nil, jsonResponse)

	return jsonResponse.Worklogs
}

func GetRapidViewID(config types.Config, board string) *types.RapidView {
	url := config.JiraURL + "/rest/greenhopper/1.0/rapidview"

	resp := new(struct {
		Views []types.RapidView `json:"views"`
	})

	getJSONResponse(config, http.MethodGet, url, nil, resp)

	for _, x := range resp.Views {
		if strings.EqualFold(board, x.Name) {
			return &x
		}
	}

	return nil
}

func GetSprints(config types.Config, rapidViewID int) ([]types.Sprint, []types.SprintIssue) {
	url := fmt.Sprintf(
		"%s/rest/greenhopper/1.0/xboard/plan/backlog/data.json?rapidViewId=%d",
		config.JiraURL, rapidViewID)

	resp := new(struct {
		Issues  []types.SprintIssue `json:"issues"`
		Sprints []types.Sprint      `json:"sprints"`
	})

	getJSONResponse(config, http.MethodGet, url, nil, resp)

	return resp.Sprints, resp.Issues
}

func IssueExists(config types.Config, issueKey *string) bool {
	url := config.JiraURL + "/rest/api/2/issue/" + *issueKey
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(config.Username, config.Password)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func getJSONResponse(config types.Config, method string, url string, payload []byte, jsonResponse interface{}) {
	// Create request
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(payload))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(config.Username, config.Password)

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(string(body))

	err = json.Unmarshal(body, jsonResponse)
	if err != nil {
		log.Fatalf("Failed to parse json response: %s\n", err)
	}

	defer resp.Body.Close()
}
