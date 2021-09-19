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
	"gitlab.com/mhersson/gojira/pkg/jira"
	"gitlab.com/mhersson/gojira/pkg/types"
	"gitlab.com/mhersson/gojira/pkg/util/validate"
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
			IssueKey = strings.ToUpper(args[0])
		}
		validate.IssueKey(Cfg, &IssueKey, IssueFile)
		status := getStatus(IssueKey)
		printStatus(status, false)
		tr := jira.GetTransistions(Cfg, IssueKey)
		printTransitions(tr)
		if len(tr) >= 1 {
			err := updateStatus(IssueKey, tr)
			if err != nil {
				fmt.Printf("Update failed: %s", err.Error())
				os.Exit(1)
			}
			status = getStatus(IssueKey)
			printStatus(status, true)
		}
	},
}

var updateAssigneeCmd = &cobra.Command{
	Use:     "assignee",
	Short:   "Assign issue to user",
	Aliases: []string{"a"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			IssueKey = strings.ToUpper(args[0])
		}
		validate.IssueKey(Cfg, &IssueKey, IssueFile)

		if Assignee == "" {
			Assignee = Cfg.Username
		}

		err := updateAssignee(IssueKey, Assignee)
		if err != nil {
			fmt.Printf("Failed to update assignee - %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Printf("%s is assigned to %s\n", IssueKey, Assignee)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	updateCmd.AddCommand(updateStatusCmd)
	updateCmd.AddCommand(updateAssigneeCmd)

	updateStatusCmd.SetUsageTemplate(updateStatusUsage)
	updateAssigneeCmd.SetUsageTemplate(updateAssigneeUsage)

	updateAssigneeCmd.PersistentFlags().StringVarP(&Assignee,
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

func updateStatus(key string, transitions []types.Transition) error {
	r := fmt.Sprintf("^([0-%d])$", len(transitions)-1)
	index := getUserInput("", r)

	i, err := strconv.Atoi(index)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	url := Cfg.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/transitions"
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

	resp, err := update("POST", url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func updateAssignee(key string, user string) error {
	url := Cfg.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/assignee"
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
	req.SetBasicAuth(Cfg.Username, Cfg.Password)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 204 {
		return body, &types.Error{Message: resp.Status}
	}

	return body, nil
}
