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
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/mhersson/gojira/pkg/jira"
	"gitlab.com/mhersson/gojira/pkg/util/convert"
	"gitlab.com/mhersson/gojira/pkg/util/format"
	"gitlab.com/mhersson/gojira/pkg/util/validate"
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

By default the time is registered on the active issue with todays date and time,
but all of them can be set explicitly. The issue key as argument, and the date
and time by using the date and time flags.

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
  -t, --time                   set the time

Example:
# Add 2 hours of work to the active issue
  # gojira add work 2h

Example specifying the issue and the date:
  # gojira add work GOJIRA-1 2h --date 2020-04-12

Example specifying the issue and the date and time:
  # gojira add work GOJIRA-1 2h --date 2020-04-12 --time 20:30

Example specifying the issue and adding a comment:
  # gojira add work GOJIRA-1 2h --comment "Helping out customer X"

Example same as above but using alias (requires g1 set to GOJIRA-1 in config)
  # gojira add work g1 2h --comment "Helping out customer X"
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
			IssueKey = strings.ToUpper(args[0])
			work = args[1]
		}

		aliasValue := Cfg.Aliases[strings.ToLower(args[0])]
		if aliasValue != "" {
			IssueKey = strings.ToUpper(aliasValue)
		}

		jira.CheckIssueKey(&IssueKey, IssueFile)
		if WorkDate != "" && !validate.Date(WorkDate) {
			fmt.Println("Invalid date. Date must be on the format yyyy-mm-dd")
			os.Exit(1)
		}

		if WorkTime != "" && !validate.Time(WorkTime) {
			fmt.Println("Invalid time. Tate must be on the format hh:mm")
			os.Exit(1)
		}

		duration, err := convert.DurationStringToSeconds(work)
		if err != nil {
			fmt.Printf("Failed to add worklog - %s", err.Error())
			os.Exit(1)
		}

		err = jira.AddWorklog(WorkDate, WorkTime, IssueKey, duration, WorkComment)
		if err != nil {
			fmt.Printf("Failed to add worklog - %s", err.Error())
			os.Exit(1)
		}

		fmt.Printf("%sSuccessfully added new worklog.%s\n", format.Color.Green, format.Color.Nocolor)
	},
}

var addCommentCmd = &cobra.Command{
	Use:     "comment",
	Short:   "Add new comment",
	Aliases: []string{"c"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			IssueKey = strings.ToUpper(args[0])
		}

		jira.CheckIssueKey(&IssueKey, IssueFile)

		comment, err := captureInputFromEditor("", "comment*")
		if err != nil {
			fmt.Println("Failed to add comment")
		}

		err = jira.AddComment(IssueKey, comment)
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

	addWorkCmd.PersistentFlags().StringVarP(&WorkDate,
		"date", "d", "", "date, overrides the default date (today)")
	addWorkCmd.PersistentFlags().StringVarP(&WorkTime,
		"time", "t", "", "time, overrides the default time (now)")
	addWorkCmd.PersistentFlags().StringVarP(&WorkComment,
		"comment", "c", "", "add a comment to you worklog")
}
