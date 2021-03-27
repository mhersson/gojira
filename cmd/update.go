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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const updateStatusUsage string = `By default the active issue gets updated,
but this can be changed by adding the issue key as argument.

Usage:
  gojira update status [ISSUE KEY] [flags]

Aliases:
  status, s

Flags:
  -h, --help                   help for status
`

const updateDescriptionUsage string = `By default the active issue gets updated,
but this can be changed by adding the issue key as argument.

Usage:
  gojira update description [ISSUE KEY] [flags]

Aliases:
  description, d

Flags:
  -h, --help                   help for description
`

const updateCommentUsage string = `By default the active issue gets updated, but this can be
changed by adding the issue key as argument. When this is
the case the argument order is important, and the issue key
must always be the first argument.

The comment id can be found by running either "get comments" or "describe".
If not set the comment id of the most recent comment will be used.

Usage:
  gojira update comment [ISSUE KEY] <COMMENT ID> [flags]

Aliases:
  comment, c

Flags:
  -h, --help                   help for comment
`

const updateAssigneeUsage string = `By default the active issue gets updated,
but this can be changed by adding the issue key as argument.

Username can be set by adding the username flag.
If no username is given the issue is assigned to you

Usage:
  gojira update assignee [ISSUE KEY] [flags]

Aliases:
  assignee, a

Flags:
  -u, --username               username of the new assignee
  -h, --help                   help for assignee
`

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update issue",
	Long:    "Update comments, description, status or assignee",
	Args:    cobra.NoArgs,
	Aliases: []string{"u"},
}

var updateStatusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Update the status",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			issueKey = strings.ToUpper(args[0])
		}
		validateIssueKey(&issueKey)
		status := getStatus(issueKey)
		printStatus(status, false)
		tr := getTransistions(issueKey)
		printTransitions(tr)
		if len(tr.Transitions) >= 1 {
			err := updateStatus(issueKey, tr)
			if err != nil {
				fmt.Printf("Update failed: %s", err.Error())
				os.Exit(1)
			}
			status = getStatus(issueKey)
			printStatus(status, true)
		}
	},
}

var updateDescrptionCmd = &cobra.Command{
	Use:     "description",
	Short:   "Update the description",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"d"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			issueKey = strings.ToUpper(args[0])
		}
		validateIssueKey(&issueKey)
		issue := getIssue(issueKey)

		desc, err := captureInputFromEditor(issue.Fields.Description, "description*")
		if err != nil {
			fmt.Println("Failed to read updated description")
			os.Exit(1)
		}

		err = updateDescription(issueKey, desc)
		if err != nil {
			fmt.Printf("Failed to update description, %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Description updated")

	},
}

var updateCommentCmd = &cobra.Command{
	Use:     "comment",
	Short:   "Update (edit) comment",
	Aliases: []string{"c"},
	Args:    cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var commentID string

		switch len(args) {
		case 1:
			// First argument can be both comment id for the active issue
			// or another issue key, where the target comment is the latest
			if validateCommentID(args[0]) {
				// Comment id is valid, the issuekey will be set to the active issue
				commentID = args[0]
				validateIssueKey(&issueKey)
			} else {
				// The argument is not a valid comment id, check if it
				// is a valid issue key
				issueKey = strings.ToUpper(args[0])
				validateIssueKey(&issueKey)
			}

		case 2:
			// If two arguments are provided first must be the issueKey,
			// and second must be the comment id
			issueKey = strings.ToUpper(args[0])
			validateIssueKey(&issueKey)

			commentID = args[1]
			if !validateCommentID(commentID) {
				fmt.Println("Invalid comment id")
				os.Exit(1)
			}

		default:
			// If no argument is provided edit the last comment of the current active issue
			validateIssueKey(&issueKey)
		}

		// Get the existing comment
		ec := getComment(issueKey, commentID)
		if commentID == "" {
			commentID = ec.ID
		}

		if ec.ID == "" {
			if commentID != "" {
				fmt.Println("Comment id does not exist")
			} else {

				fmt.Println("Issue does not have any comments. Try add comment instead")
			}
			os.Exit(1)
		}

		comment, err := captureInputFromEditor(ec.Body, "comment*")
		if err != nil {
			fmt.Println("Failed to update comment")
		}

		err = updateComment(issueKey, comment, commentID)
		if err != nil {
			fmt.Printf("Failed to update comment - %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Println("Comment updated")
	},
}

var updateAssigneeCmd = &cobra.Command{
	Use:     "assignee",
	Short:   "Assign issue to user",
	Aliases: []string{"a"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			issueKey = strings.ToUpper(args[0])
		}
		validateIssueKey(&issueKey)

		if assignee == "" {
			assignee = config.Username
		}

		err := updateAssignee(issueKey, assignee)
		if err != nil {
			fmt.Printf("Failed to update assignee - %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Printf("%s is assigned to %s\n", issueKey, assignee)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.AddCommand(updateStatusCmd)
	updateCmd.AddCommand(updateDescrptionCmd)
	updateCmd.AddCommand(updateCommentCmd)
	updateCmd.AddCommand(updateAssigneeCmd)

	updateStatusCmd.SetUsageTemplate(updateStatusUsage)
	updateDescrptionCmd.SetUsageTemplate(updateDescriptionUsage)
	updateAssigneeCmd.SetUsageTemplate(updateAssigneeUsage)
	updateCommentCmd.SetUsageTemplate(updateCommentUsage)

	updateAssigneeCmd.PersistentFlags().StringVarP(&assignee,
		"username", "u", "", "username of the new assignee")
}

func getUserInput(prompt string, regRange string) string {
	if prompt == "" {
		fmt.Print("\nPlease enter value (press enter to quit): ")
	} else {
		fmt.Print(prompt)
	}

	reader := bufio.NewReader(os.Stdin)

	var answer string

	for {
		input, _ := reader.ReadBytes('\n')
		if input[0] == '\n' {
			fmt.Println("Cancelled by user")
			os.Exit(0)
		}

		re := regexp.MustCompile(regRange)
		m := re.Find(bytes.TrimSpace(input))

		if m == nil {
			fmt.Println("Invalid choice")
			fmt.Print("Please try again: ")

			continue
		}

		answer = string(m)

		break
	}

	return answer
}

func updateStatus(key string, tr TransitionsResponse) error {
	r := fmt.Sprintf("^([0-%d])$", len(tr.Transitions)-1)
	index := getUserInput("", r)

	i, err := strconv.Atoi(index)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/transitions"
	id := tr.Transitions[i].ID

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

	resp, err := update("POST", url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func updateDescription(key string, desc []byte) error {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key)

	jsonDesc := makeStringJSONSafe(string(desc))

	payload := []byte(`{"fields":{"description":"` + jsonDesc + `"}}`)

	resp, err := update("PUT", url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func updateComment(key string, comment []byte, id string) error {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/comment/" + id

	escaped := makeStringJSONSafe(string(comment))

	payload := []byte(`{
		"body": "` + escaped + `",
		"visibility": {
			"type": "group",
			"value": "Internal users"
		}
	}`)

	resp, err := update("PUT", url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func updateAssignee(key string, user string) error {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/assignee"
	payload := []byte(`{"name":"` + user + `"}`)

	resp, err := update("PUT", url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func update(method, url string, payload []byte) ([]byte, error) {
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.SetBasicAuth(config.Username, config.Password)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 204 {
		return body, &Error{resp.Status}
	}

	return body, nil
}

func validateCommentID(commentID string) bool {
	// This maybe wrong, but so far I have not
	// seen an id which is not 6 digits long
	re := regexp.MustCompile("^[0-9]{6}$")

	return re.MatchString(commentID)
}
