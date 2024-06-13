/*
Copyright © 2020-2022 Morten Hersson

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
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mhersson/gojira/pkg/types"
	"github.com/mhersson/gojira/pkg/util"
	"github.com/mhersson/gojira/pkg/util/validate"
)

var jcfg types.JiraConfig

const restAPIIssueURL = "/rest/api/2/issue/"

func Configure(config types.Config) {
	jcfg.Server = config.JiraURL
	jcfg.Username = config.Username
	jcfg.Password = config.Password
	jcfg.PasswordType = config.PasswordType
	jcfg.Decrypted = false
}

func GetIssues(filter string) []types.Issue {
	url := jcfg.Server + "/rest/api/2/search"

	if filter == "" {
		filter = `assignee = ` + jcfg.Username +
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

	query(http.MethodPost, url, payload, jsonResponse)

	return jsonResponse.Issues
}

func GetTimesheet(fromDate, toDate string, showEntireWeek bool) []types.Timesheet {
	url := jcfg.Server + "/rest/timesheet-gadget/1.0/raw-timesheet.json?startDate=" + fromDate + "&endDate=" + toDate

	if showEntireWeek {
		// Date is already validated, so should be safe
		// to drop the error check here
		t, _ := time.Parse("2006-01-02", fromDate)
		start, end := util.WeekStartEndDate(t.ISOWeek())
		url = jcfg.Server + "/rest/timesheet-gadget/1.0/raw-timesheet.json?startDate=" + start + "&endDate=" + end
	}

	jsonResponse := new(struct {
		Worklog []types.Timesheet `json:"worklog"`
	})

	query(http.MethodGet, url, nil, jsonResponse)

	return jsonResponse.Worklog
}

func GetTimesheetForUser(date, username string) []types.Timesheet {
	url := jcfg.Server + "/rest/timesheet-gadget/1.0/raw-timesheet.json?startDate=" +
		date + "&endDate=" + date + "&targetUser=" + username

	jsonResponse := new(struct {
		Worklog []types.Timesheet `json:"worklog"`
	})

	query(http.MethodGet, url, nil, jsonResponse)

	return jsonResponse.Worklog
}

func GetValidProjects() []types.Project {
	url := jcfg.Server + "/rest/api/2/project"

	jsonResponse := new([]types.Project)

	query(http.MethodGet, url, nil, jsonResponse)

	return *jsonResponse
}

func GetProjectIssueTypes(projectKey string) []types.IssueType {
	url := jcfg.Server + "/rest/api/2/issue/createmeta/" + projectKey + "/issuetypes"

	jsonResponse := new(struct {
		Values []types.IssueType `json:"values"`
	})

	query(http.MethodGet, url, nil, jsonResponse)

	return jsonResponse.Values
}

func GetPriorities() []types.Priority {
	url := jcfg.Server + "/rest/api/2/priority"

	jsonResponse := &[]types.Priority{}

	query(http.MethodGet, url, nil, jsonResponse)

	return *jsonResponse
}

func GetIssueTypes() *[]types.IssueType {
	url := jcfg.Server + "/rest/api/2/issuetype"

	jsonResponse := &[]types.IssueType{}

	query(http.MethodGet, url, nil, jsonResponse)

	return jsonResponse
}

func GetIssue(key string) types.IssueDescription {
	url := jcfg.Server + restAPIIssueURL + strings.ToUpper(key)

	jsonResponse := &types.IssueDescription{}

	query(http.MethodGet, url, nil, jsonResponse)

	return *jsonResponse
}

func GetIssuesInEpic(key string) []types.Issue {
	url := jcfg.Server + "/rest/api/2/search?jql=cf[10500]=" + strings.ToUpper(key)

	jsonResponse := new(struct {
		Issues []types.Issue `json:"issues"`
	})

	query(http.MethodGet, url, nil, jsonResponse)

	return jsonResponse.Issues
}

func GetTransistions(key string) []types.Transition {
	url := jcfg.Server + restAPIIssueURL + strings.ToUpper(key) + "/transitions"

	jsonResponse := new(struct {
		Transitions []types.Transition `json:"transitions"`
	})

	query(http.MethodGet, url, nil, jsonResponse)

	return jsonResponse.Transitions
}

func GetComments(key string) []types.Comment {
	url := jcfg.Server + restAPIIssueURL + strings.ToUpper(key) + "/comment"

	jsonResponse := new(struct {
		Comments []types.Comment `json:"comments"`
	})

	query(http.MethodGet, url, nil, jsonResponse)

	return jsonResponse.Comments
}

func GetWorklogs(key string) []types.Worklog {
	url := jcfg.Server + restAPIIssueURL + strings.ToUpper(key) + "/worklog"

	jsonResponse := new(struct {
		Worklogs []types.Worklog `json:"worklogs"`
	})

	query(http.MethodGet, url, nil, jsonResponse)

	return jsonResponse.Worklogs
}

func GetRapidViewID(board string) *types.RapidView {
	url := jcfg.Server + "/rest/greenhopper/1.0/rapidview"

	resp := new(struct {
		Views []types.RapidView `json:"views"`
	})

	query(http.MethodGet, url, nil, resp)

	for _, x := range resp.Views {
		if strings.EqualFold(board, x.Name) {
			return &x
		}
	}

	return nil
}

func GetSprints(rapidViewID int) ([]types.Sprint, []types.SprintIssue) {
	url := fmt.Sprintf(
		"%s/rest/greenhopper/1.0/xboard/plan/backlog/data.json?rapidViewId=%d",
		jcfg.Server, rapidViewID)

	resp := new(struct {
		Issues  []types.SprintIssue `json:"issues"`
		Sprints []types.Sprint      `json:"sprints"`
	})

	query(http.MethodGet, url, nil, resp)

	return resp.Sprints, resp.Issues
}

func GetKanbanIssues(boardID int) []types.Issue {
	url := fmt.Sprintf("%s/rest/agile/1.0/board/%d/issue", jcfg.Server, boardID)

	resp := new(struct {
		Issues []types.Issue `json:"issues"`
	})

	query(http.MethodGet, url, nil, resp)

	return resp.Issues
}

func CheckIssueKey(key *string, issueFile string) {
	if *key != "" {
		if !validate.IssueKey(key) {
			fmt.Println("Invalid key")
			os.Exit(1)
		}

		if !IssueExists(key) {
			fmt.Printf("%s does not exist\n", *key)
			os.Exit(1)
		}
	} else {
		*key = util.GetActiveIssue(issueFile)
	}
}

func IssueExists(issueKey *string) bool {
	url := jcfg.Server + restAPIIssueURL + *issueKey

	return exists(url)
}

func UserExists(username string) bool {
	url := jcfg.Server + "/rest/api/2/user/?username=" + username

	return exists(url)
}

func UpdateStatus(key string, transitions []types.Transition) error {
	r := fmt.Sprintf("^([0-%d])$", len(transitions)-1)
	index := util.GetUserInput("", r)

	i, err := strconv.Atoi(index)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	url := jcfg.Server + restAPIIssueURL + strings.ToUpper(key) + "/transitions"
	id := transitions[i].ID

	payload := []byte(`{
		"update": {
			"comment": [
				{
					"add": {
						"body": "Status updated by Gojira"
					}
				}
			]
		},
		"transition": {
			"id": "` + id + `"
		}
	}`)

	resp, err := update(http.MethodPost, url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func UpdateAssignee(key string, user string) error {
	url := jcfg.Server + restAPIIssueURL + strings.ToUpper(key) + "/assignee"
	payload := []byte(`{"name":"` + user + `"}`)

	resp, err := update(http.MethodPut, url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func CreateNewIssue(project types.Project, issueTypeID,
	priorityID, summary, description string,
) (string, error) {
	url := jcfg.Server + "/rest/api/2/issue"
	method := http.MethodPost

	payload := []byte(`{
		"fields":{
			"project": {
				"id": "` + project.ID + `"
			},
			"summary": "` + summary + `",
			"description": "` + description + `",
			"issuetype": {
				"id": "` + issueTypeID + `"
			},
			"priority": {
				"id": "` + priorityID + `"
			}
		}
	}`)

	// If issueType is Task or Improvement add the
	// Change visibility to Exclude change in release notes
	if issueTypeID == "3" || issueTypeID == "4" {
		re := regexp.MustCompile(`},(\n|.)+?"summary"`)
		payload = re.ReplaceAll(payload, []byte(`},
				"customfield_10707": {
					"value": "Exclude change in release notes"
				},
				"summary"`))
	}

	body, err := update(method, url, payload)
	if err != nil {
		return string(body), err
	}

	var resp struct {
		Key string `json:"key"`
	}

	err = json.Unmarshal(body, &resp)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	return resp.Key, nil
}

func AddWorklog(wDate, wTime, key, seconds, comment string) error {
	url := jcfg.Server + restAPIIssueURL + strings.ToUpper(key) + "/worklog"
	payload := []byte(`{
		"comment": "` + comment + `",
		"started": "` + setWorkStarttime(wDate, wTime) + `",
		"timeSpentSeconds": ` + seconds +
		`}`)

	resp, err := update(http.MethodPost, url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func AddComment(key string, comment []byte) error {
	url := jcfg.Server + restAPIIssueURL + strings.ToUpper(key) + "/comment"

	escaped := util.MakeStringJSONSafe(string(comment))

	payload := []byte(`{
		"body": "` + escaped + `",
		"visibility": {
			"type": "group",
			"value": "Internal users"
		}
	}`)

	resp, err := update(http.MethodPost, url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func UpdateDescription(key string, desc []byte) error {
	url := jcfg.Server + restAPIIssueURL + strings.ToUpper(key)

	jsonDesc := util.MakeStringJSONSafe(string(desc))

	payload := []byte(`{"fields":{"description":"` + jsonDesc + `"}}`)

	resp, err := update(http.MethodPut, url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func UpdateComment(key string, comment []byte, id string) error {
	url := jcfg.Server + restAPIIssueURL + strings.ToUpper(key) + "/comment/" + id

	escaped := util.MakeStringJSONSafe(string(comment))

	payload := []byte(`{
		"body": "` + escaped + `",
		"visibility": {
			"type": "group",
			"value": "Internal users"
		}
	}`)

	resp, err := update(http.MethodPut, url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func UpdateWorklog(worklog types.SimplifiedTimesheet) error {
	dateAndTime := strings.Split(worklog.StartDate, " ")
	if len(dateAndTime) != 2 {
		return &types.Error{Message: "invalid date and time"}
	}

	url := jcfg.Server + restAPIIssueURL +
		strings.ToUpper(worklog.Key) + "/worklog/" + strconv.Itoa(worklog.ID) + "/"

	payload := []byte(`{
		"id": "` + strconv.Itoa(worklog.ID) + `",
		"comment": "` + worklog.Comment + `",
		"started": "` + setWorkStarttime(dateAndTime[0], dateAndTime[1]) + `",
		"timeSpentSeconds": ` + strconv.Itoa(worklog.TimeSpent) +
		`}`)

	resp, err := update(http.MethodPut, url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func setWorkStarttime(wDate, wTime string) string {
	now := time.Now()
	zone, _ := now.Zone()

	// jira time format - "started": "2017-12-07T09:23:19.552+0000"
	startTime := now.UTC().Format("2006-01-02T15:04:05.000+0000")

	switch {
	case wDate == "" && wTime == "":
		return startTime
	case wDate != "" && wTime == "":
		wTime = time.Now().Format("15:04")
	case wDate == "" && wTime != "":
		wDate = now.Format("2006-01-02")
	}

	t, _ := time.Parse("2006-01-02 15:04 MST", fmt.Sprintf("%s %s %s", wDate, wTime, zone))

	return t.UTC().Format("2006-01-02T15:04:05.000+0000")
}

func update(method, url string, payload []byte) ([]byte, error) {
	jcfg.DecryptPassword()

	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(jcfg.Username, jcfg.Password)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusNoContent {
		return body, &types.Error{Message: checkResponseCode(resp)}
	}

	return body, nil
}

func query(method string, url string, payload []byte, jsonResponse interface{}) {
	// Create request
	jcfg.DecryptPassword()

	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(payload))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(jcfg.Username, jcfg.Password)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusNoContent {
		log.Fatalf("Error:%s", checkResponseCode(resp))
	}
	// fmt.Println(string(body))

	err = json.Unmarshal(body, jsonResponse)
	if err != nil {
		log.Fatalf("Failed to parse json response: %s\n", err)
	}

	resp.Body.Close()
}

func exists(url string) bool {
	jcfg.DecryptPassword()

	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(jcfg.Username, jcfg.Password)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	if resp.StatusCode != http.StatusOK &&
		resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusNoContent {
		log.Fatalf("Error:%s", checkResponseCode(resp))
	}

	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func checkResponseCode(resp *http.Response) string {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return resp.Status + ". Please check your credentials"
	case http.StatusForbidden:
		return resp.Status + ". Please check that your account is not blocked by captcha."
	default:
		return resp.Status
	}
}
