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
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/mhersson/gojira/pkg/jira"
	"gitlab.com/mhersson/gojira/pkg/types"
	"gitlab.com/mhersson/gojira/pkg/util"
	"gitlab.com/mhersson/gojira/pkg/util/convert"
	"gitlab.com/mhersson/gojira/pkg/util/validate"
)

const editDescriptionUsage string = `By default the active issue is edited,
but this can be changed by adding the issue key as argument.

Usage:
  gojira edit description [ISSUE KEY] [flags]

Aliases:
  description, d

Flags:
  -h, --help                   help for description
`

const editCommentUsage string = `By default the active issue is edited, but this can be
changed by adding the issue key as argument. When this is
the case the argument order is important, and the issue key
must always be the first argument.

The comment id can be found by running either "get comments" or "describe".
If not set the comment id of the most recent comment will be used.

Usage:
  gojira edit comment [ISSUE KEY] <COMMENT ID> [flags]

Aliases:
  comment, c

Flags:
  -h, --help                   help for comment
`

var editCmd = &cobra.Command{
	Use:     "edit",
	Short:   "Edit comments, descriptions and your worklog",
	Args:    cobra.NoArgs,
	Aliases: []string{"e"},
}

var editDescrptionCmd = &cobra.Command{
	Use:     "description",
	Short:   "Edit the description",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"d"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			IssueKey = strings.ToUpper(args[0])
		}
		jira.CheckIssueKey(&IssueKey, IssueFile)
		issue := jira.GetIssue(IssueKey)

		desc, err := captureInputFromEditor(issue.Fields.Description, "description*")
		if err != nil {
			fmt.Println("Failed to read description")
			os.Exit(1)
		}

		err = jira.UpdateDescription(IssueKey, desc)
		if err != nil {
			fmt.Printf("Failed to update description, %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully saved new description")

	},
}

var editCommentCmd = &cobra.Command{
	Use:     "comment",
	Short:   "Edit comment",
	Aliases: []string{"c"},
	Args:    cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var commentID string

		switch len(args) {
		case 1:
			// First argument can be both comment id for the active issue
			// or another issue key, where the target comment is the latest
			if validate.CommentID(args[0]) {
				// Comment id is valid, the issuekey will be set to the active issue
				commentID = args[0]
				jira.CheckIssueKey(&IssueKey, IssueFile)
			} else {
				// The argument is not a valid comment id, check if it
				// is a valid issue key
				IssueKey = strings.ToUpper(args[0])
				jira.CheckIssueKey(&IssueKey, IssueFile)
			}

		case 2:
			// If two arguments are provided first must be the issueKey,
			// and second must be the comment id
			IssueKey = strings.ToUpper(args[0])
			jira.CheckIssueKey(&IssueKey, IssueFile)

			commentID = args[1]
			if !validate.CommentID(commentID) {
				fmt.Println("Invalid comment id")
				os.Exit(1)
			}

		default:
			// If no argument is provided edit the last comment of the current active issue
			jira.CheckIssueKey(&IssueKey, IssueFile)
		}

		// Get the existing comment
		ec := getComment(IssueKey, commentID)
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
			fmt.Println("Failed to read comment")
		}

		err = jira.UpdateComment(IssueKey, comment, commentID)
		if err != nil {
			fmt.Printf("Failed to update comment - %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Println("Successfully saved new comment")
	},
}

var editMyWorklogCmd = &cobra.Command{
	Use:     "myworklog",
	Short:   "Edit your worklog for a given date",
	Aliases: []string{"m"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		date := util.GetCurrentDate()
		if len(args) == 1 {
			date = args[0]
		}
		if Cfg.UseTimesheetPlugin {
			if validate.Date(date) {
				ts := jira.GetTimesheet(date, date, ShowEntireWeek)
				if len(ts) == 0 {
					fmt.Println("There is nothing to edit.")
					os.Exit(0)
				}
				worklogs := util.GetWorklogsSorted(ts, false)

				// If mergetoday is set
				if !util.DateIsToday(date) && MergeToday && !ShowEntireWeek {
					date = util.Today() // Set the date today
					ts = jira.GetTimesheet(date, date, ShowEntireWeek)
					wlToday := util.GetWorklogsSorted(ts, false)

					// Reset the ID and the date, and append the logs on today
					for _, w := range worklogs {
						wlToday = append(wlToday, types.SimplifiedTimesheet{
							ID:        666,
							Date:      date,
							StartDate: w.StartDate,
							Key:       w.Key,
							Summary:   w.Summary,
							Comment:   w.Comment,
							TimeSpent: w.TimeSpent,
						})
					}

					// Set the complete list as the worklog
					worklogs = wlToday
				}

				out := util.ExecuteTemplate("edit-worklog.tmpl", worklogs)
				edited, err := captureInputFromEditor(string(out), "edit-worklog-*")
				cobra.CheckErr(err)

				editedWorklogs := parseEditedWorklog(date, edited)
				updateChangedWorklogs(worklogs, editedWorklogs)
				addNewWorklogs(editedWorklogs)
			}
		} else {
			fmt.Println("This command is currently only supported with the timesheet plugin enabled")
		}
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.AddCommand(editDescrptionCmd)
	editCmd.AddCommand(editCommentCmd)
	editCmd.AddCommand(editMyWorklogCmd)

	editDescrptionCmd.SetUsageTemplate(editDescriptionUsage)
	editCommentCmd.SetUsageTemplate(editCommentUsage)
	editMyWorklogCmd.Flags().BoolVarP(&MergeToday, "merge-today", "", false, "merge/append the records from that date")
}

func updateChangedWorklogs(worklogs, editedWorklogs []types.SimplifiedTimesheet) {
	success := 0

	for _, e := range editedWorklogs {
		for _, w := range worklogs {
			if e.ID == w.ID && e.ID != 666 &&
				(e.StartDate != w.StartDate || e.TimeSpent != w.TimeSpent || e.Comment != w.Comment) {
				err := jira.UpdateWorklog(e)
				if err != nil {
					fmt.Printf("Failed to update worklog id: %d, key; %s\n", e.ID, e.Key)
					fmt.Printf("%v\n", err)
					os.Exit(1)
				}
				success++

				break
			}
		}
	}

	if success >= 1 {
		fmt.Printf("Successfully updated %d worklog entries\n", success)
	}
}

func addNewWorklogs(editedWorklogs []types.SimplifiedTimesheet) {
	success := 0

	for _, e := range editedWorklogs {
		dateAndTime := strings.Split(e.StartDate, " ")

		if e.ID == 666 {
			err := jira.AddWorklog(dateAndTime[0], dateAndTime[1], e.Key, strconv.Itoa(e.TimeSpent), e.Comment)
			if err != nil {
				fmt.Printf("Failed to add new worklog key; %s\n", e.Key)
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
			success++
		}
	}

	if success >= 1 {
		fmt.Printf("Successfully added %d worklog entries\n", success)
	}
}

func getComment(key, commentID string) types.Comment {
	comments := jira.GetComments(key)

	if commentID == "" && len(comments) >= 1 {
		return comments[len(comments)-1]
	}

	for _, c := range comments {
		if c.ID == commentID {
			return c
		}
	}

	return types.Comment{}
}

func parseEditedWorklog(date string, logs []byte) []types.SimplifiedTimesheet {
	// (#123456)    ISSUE-1       14:30    0h 30m    Some comment
	re := regexp.MustCompile(
		`\(#?([0-9]{6}|new)\)\s{1,}` + // ID
			`([A-Z]{2,9}-[0-9]{1,4})\s{1,}` + // Key
			`(([0-1][0-9]|2[0-3]):[0-5][0-9])\s{1,}` + // Time
			`(([0-9.]{1,}h)?\s?([0-6]?[0-9]m)?)\s*` + // Duration
			`([A-Za-z0-9_\-,\.\s]+)`) // Comment

	m := re.FindAllStringSubmatch(string(logs), -1)

	worklogs := []types.SimplifiedTimesheet{}

	for _, match := range m {
		ts := new(types.SimplifiedTimesheet)
		if match[1] == "new" {
			ts.ID = 666
		} else {
			ts.ID, _ = strconv.Atoi(match[1])
		}

		ts.Key = match[2]
		ts.StartDate = date + " " + match[3]

		d, err := convert.DurationStringToSeconds(match[5])
		cobra.CheckErr(err)

		ts.TimeSpent, _ = strconv.Atoi(d)
		ts.Comment = strings.TrimSpace(match[8])

		worklogs = append(worklogs, *ts)
	}

	return worklogs
}
