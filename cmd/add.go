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
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const addCommentUsage string = `This command will add a new comment to an issue.
The input supports multiline text, and must be terminated with Ctrl+D.
Writing JIRA notation, with {noformat} and {code}, is supported, but for
easier writing three backticks will be converted to {noformat}.

By default the comment is added to the active issue,
but this can be changed by adding the issue key as argument.

Usage:
  gojira add comment [ISSUE KEY] [flags]

Aliases:
  comment, c

Flags:
  -h, --help                   help for comment
`

const addWorkUsage string = `This command will add work to an issue worklog.
The time must be specified in either hours OR minutes on this format: 2h OR 120m

By default the time is registered on the active issue with todays date,
but both the issue key and the date can be set explicitly. The issue key as argument,
and the date by using the date flag.

When specifying the issue key the argument order is important,
and the issue key must always come first.

Valid date format is yyyy-mm-dd

Usage:
  gojira add work [ISSUE KEY] <TIME> [flags]

Aliases:
  work, w

Flags:
  -c. --comment                add a comment with the worklog
  -d, --date                   set the date
  -h, --help                   help for work

Example:
# Add 2 hours of work to the active issue
  # gojira add work 2h

Example specifying the issue and the date:
  # gojira add work GOJIRA-1 2h --date 2020-04-12

Example specifying the issue and adding a comment:
  # gojira add work GOJIRA-1 2h --comment "Helping out customer X"
`

var addCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add a comment or register time",
	Args:    cobra.NoArgs,
	Aliases: []string{"a"},
}

var addWorkCmd = &cobra.Command{
	Use:     "work",
	Short:   "Add work (format 2h or 120m)",
	Aliases: []string{"w"},
	Args:    cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var work string

		if len(args) == 1 {
			work = args[0]
		} else {
			issueKey = strings.ToUpper(args[0])
			work = args[1]
		}

		validateIssueKey(&issueKey)
		if workDate != "" && !validateDate(workDate) {
			fmt.Println("Invalid date. Date must be on the format yyyy-mm-dd")
			os.Exit(1)
		}

		duration, err := validateWorkArgs(work)
		if err != nil {
			fmt.Printf("Failed to add worklog - %s", err.Error())
			os.Exit(1)
		}

		err = addWork(issueKey, duration, workComment)
		if err != nil {
			fmt.Printf("Failed to add worklog - %s", err.Error())
			os.Exit(1)
		}

		fmt.Printf("%sSuccessfully added new worklog.%s\n", color.green, color.nocolor)
		printTimeTracking(issueKey)
	},
}

var addCommentCmd = &cobra.Command{
	Use:     "comment",
	Short:   "Add new comment",
	Aliases: []string{"c"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			issueKey = strings.ToUpper(args[0])
		}

		validateIssueKey(&issueKey)

		comment, err := captureInputFromEditor("", "comment*")
		if err != nil {
			fmt.Println("Failed to add comment")
		}

		err = addComment(issueKey, comment)
		if err != nil {
			fmt.Printf("Failed to add comment - %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Println("Successfully added comment")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.AddCommand(addCommentCmd)
	addCmd.AddCommand(addWorkCmd)

	addCommentCmd.SetUsageTemplate(addCommentUsage)
	addWorkCmd.SetUsageTemplate(addWorkUsage)

	addWorkCmd.PersistentFlags().StringVarP(&workDate,
		"date", "d", "", "date, overrides the default date (today)")
	addWorkCmd.PersistentFlags().StringVarP(&workComment,
		"comment", "c", "", "add a comment to you worklog")
}

func validateWorkArgs(args string) (string, error) {
	re := regexp.MustCompile("^([0-9.]{1,})([m|h])$")
	m := re.FindStringSubmatch(args)

	var seconds float64

	if m != nil {
		num, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return "", fmt.Errorf("%w", err)
		}

		if m[2] == "h" {
			seconds = num * 3600
		} else {
			seconds = num * 60
		}

		return strconv.FormatFloat(seconds, 'f', 0, 64), nil
	}

	return "", &Error{"invalid duration format"}
}

func setWorkStarttime() string {
	now := time.Now().UTC()
	// jira time format - "started": "2017-12-07T09:23:19.552+0000"
	startTime := now.Format("2006-01-02T15:04:05.000+0000")

	if workDate == "" {
		return startTime
	}

	// Shouldn't get here unless the workDate is valid
	re := regexp.MustCompile("202[0-9]-((0[1-9])|(1[0-2]))-((0[1-9])|([1-2][0-9])|(3[0-1]))")
	startTime = re.ReplaceAllString(startTime, workDate)

	return startTime
}

func validateDate(date string) bool {
	re := regexp.MustCompile("202[0-9]-((0[1-9])|(1[0-2]))-((0[1-9])|([1-2][0-9])|(3[0-1]))")

	return re.MatchString(date)
}

func addWork(key string, seconds string, comment string) error {
	if comment == "" {
		comment = "Worklog updated by Gojira"
	}

	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/worklog"
	payload := []byte(`{
		"comment": "` + comment + `",
		"started": "` + setWorkStarttime() + `",
		"timeSpentSeconds": ` + seconds +
		`}`)

	resp, err := update("POST", url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func addComment(key string, comment []byte) error {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/comment"

	escaped := makeStringJSONSafe(string(comment))

	payload := []byte(`{
		"body": "` + escaped + `",
		"visibility": {
			"type": "group",
			"value": "Internal users"
		}
	}`)

	resp, err := update("POST", url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func makeStringJSONSafe(str string) string {
	strText := strings.ReplaceAll(str, "```", "{noformat}")
	// Convert the string into json to escape whatever
	// chars json needs to have escaped
	jsonStr, err := json.Marshal(strText)
	if err != nil {
		fmt.Println("Failed to parse comment")
		os.Exit(1)
	}

	// Remove the surrounding curly brackets
	escaped := string(jsonStr[1 : len(jsonStr)-1])

	return escaped
}
