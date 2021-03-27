/*
Copyright © 2020 Morten Hersson

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
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// createCmd represents the create command.
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new issue",
	Long: `Create new issue
This guides the user through as series of questions which
can be aborted at anytime.

The description input supports multiple  lines of text,
and must be terminated by Ctrl+D. Writing JIRA notation, with
{noformat} and {code}, is supported, but for easier writing
three backticks will be converted to {noformat}.

After all data is collected they must be verified and confirmed
by the user, and only then will the request be sent to JIRA.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToUpper(args[0])
		validProjects := getValidProjectsAndIssueType()
		project := validateProjectKey(key, validProjects)
		if project.ID == "" {
			fmt.Printf("%s is not a valid project key\n", key)
			os.Exit(1)
		}
		fmt.Printf("Creating new %s issue\n", project.Key)
		summary, rawSummary := getUserInputSummary()
		issueTypeID, issueTypeName := getUserInputIssueType(project)
		priorityID, priorityName := getUserInputPriority()
		desc, rawDesc := getUserInputDescription()

		getUserInputConfirmOk(project, issueTypeName, priorityName, rawSummary, rawDesc)

		newKey, err := createNewIssue(project, issueTypeID, priorityID, summary, desc)
		if err != nil {
			fmt.Printf("Failed to create issue - %s\n", err.Error())
			fmt.Println(newKey)
			os.Exit(1)
		}

		fmt.Printf("%sNew issue has got key %s%s\n", color.blue, newKey, color.nocolor)

		ans := getUserInput("Do you want to set the new issue active [y/N]: ", "[y|n]")
		if ans == "y" {
			setActiveIssue(newKey)
		}

		fmt.Printf("\n%sSuccessfully created new issue - run describe to see the details%s\n\n",
			color.green, color.nocolor)
		// printIssue(getIssue(newKey), IssueDescriptionResponse{})

	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func getValidProjectsAndIssueType() IssueCreateMeta {
	url := config.JiraURL + "/rest/api/2/issue/createmeta"

	jsonResponse := &IssueCreateMeta{}

	getJSONResponse("GET", url, nil, jsonResponse)

	return *jsonResponse
}

func validateProjectKey(key string, projects IssueCreateMeta) Project {
	// This validates the project key not the issue key
	// hvis is the project key + a number
	for _, v := range projects.Projects {
		if key == strings.ToUpper(v.Key) {
			return v
		}
	}

	return Project{}
}

func getPriorities() []Priority {
	url := config.JiraURL + "/rest/api/2/priority"

	jsonResponse := &[]Priority{}

	getJSONResponse("GET", url, nil, jsonResponse)

	return *jsonResponse
}

func getIssueTypes() *[]IssueType {
	url := config.JiraURL + "/rest/api/2/issuetype"

	jsonResponse := &[]IssueType{}

	getJSONResponse("GET", url, nil, jsonResponse)

	return jsonResponse
}

func getUserInputPriority() (string, string) {
	priorities := getPriorities()

	fmt.Println("Choose issue priority:")

	for i, v := range priorities {
		fmt.Printf("%d. %s\n", i, v.Name)
	}

	r := fmt.Sprintf("^([0-%d])$", len(priorities)-1)
	index := getUserInput("", r)

	x, _ := strconv.Atoi(index)

	for i, v := range priorities {
		if i == x {
			return v.ID, v.Name
		}
	}

	return "", ""
}

func getUserInputIssueType(project Project) (string, string) {
	fmt.Println("Choose issue type:")

	for i, v := range project.IssueTypes {
		fmt.Printf("%d. %s\n", i, v.Name)
	}

	// This is bad code - only supports range from 0-19
	r := fmt.Sprintf("^([0-%d])$", len(project.IssueTypes)-1)
	if len(project.IssueTypes) > 10 {
		r = fmt.Sprintf("^([0-9][0-%d]?)$", len(project.IssueTypes)-11)
	}

	index := getUserInput("", r)

	x, _ := strconv.Atoi(index)

	for i, v := range project.IssueTypes {
		if i == x {
			return v.ID, v.Name
		}
	}

	return "", ""
}

func getUserInputSummary() (string, string) {
	fmt.Print("Enter summary: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadBytes('\n')

	if input[0] == '\n' {
		os.Exit(0)
	}

	st := strings.TrimSpace(string(input))

	summary, err := json.Marshal(st)
	if err != nil {
		fmt.Println("Failed to parse comment")
		os.Exit(1)
	}

	// Remove the {} around the comment
	escaped := string(summary[1 : len(summary)-1])

	return escaped, st
}

func getUserInputDescription() (string, string) {
	desc, err := captureInputFromEditor("", "description*")
	if err != nil {
		fmt.Println("Failed to read user input")
		os.Exit(1)
	}

	escaped := makeStringJSONSafe(string(desc))

	return escaped, string(desc)
}

func getUserInputConfirmOk(project Project, issueType, pri, summary, description string) bool {
	fmt.Printf("%sPlease check you input:%s\n", color.blue, color.nocolor)
	fmt.Printf("Project %s, Type: %s, Priority: %s\n", project.Key, issueType, pri)
	fmt.Printf("Summary: %s\n", summary)
	fmt.Printf("Description:\n%s\n", description)

	ans := getUserInput("Is this correct [y/N]: ", "[y|n]")
	if ans == "y" {
		return true
	}

	fmt.Println("Cancelled by user")

	return false
}

func createNewIssue(project Project, issueTypeID,
	priorityID, summary, description string) (string, error) {
	url := config.JiraURL + "/rest/api/2/issue"
	method := "POST"

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
