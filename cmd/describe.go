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
	"strings"

	"github.com/spf13/cobra"

	"github.com/mhersson/gojira/pkg/jira"
	"github.com/mhersson/gojira/pkg/types"
	"github.com/mhersson/gojira/pkg/util/format"
)

const describeUsage string = `
By default the active issue will be described,
but this can be changed by adding the issue key as argument.

Usage:
  gojira describe [ISSUE KEY] [flags]

Aliases:
  describe, d

Flags:
  -h, --help                   help for describe
`

// describeCmd represents the describe command.
var describeCmd = &cobra.Command{
	Use:     "describe",
	Short:   "Display issue with all its gory details",
	Aliases: []string{"d"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			IssueKey = strings.ToUpper(args[0])
		}
		jira.CheckIssueKey(&IssueKey, IssueFile)
		issue := jira.GetIssue(IssueKey)

		var epic types.IssueDescription
		if issue.Fields.Epic != "" {
			epic = jira.GetIssue(issue.Fields.Epic)
		}

		var issues []types.Issue
		if issue.Fields.IssueType.Name == "Epic" {
			issues = jira.GetIssuesInEpic(issue.Key)
		}

		printIssue(issue, epic)

		if len(issues) > 0 {
			fmt.Printf("\n%sIssues in Epic:%s\n", format.Color.Ul, format.Color.Nocolor)
			printIssues(issues, false, true)
		}
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)

	describeCmd.SetUsageTemplate(describeUsage)
}

func printIssue(issue, epic types.IssueDescription) {
	fmt.Println()
	fmt.Println(format.Header(issue.Fields.Project.Name, issue.Key, issue.Fields.Summary))
	fmt.Printf("%sDetails:%s\n", format.Color.Ul, format.Color.Nocolor)
	fmt.Printf("Type:              %sStatus:      %s\n",
		format.IssueType(issue.Fields.IssueType.Name, false), format.Status(issue.Fields.Status.Name, false))
	fmt.Printf("Priority:          %sResolution:  %s\n",
		format.Priority(issue.Fields.Priority.Name, false), issue.Fields.Resolution.Name)
	fmt.Printf("Labels:            %s\n", strings.Join(issue.Fields.Labels, ", "))
	fmt.Printf("Fixed Version/s:   %s\n", format.FixVersions(issue))
	fmt.Printf("Visibility:        %s\n", issue.Fields.ChangeVisibility.Value)

	if epic.Fields.Summary != "" {
		fmt.Printf("Epic:              %s\n", format.Epic(epic.Fields.Summary))
	}
	// ******************************************************************
	fmt.Printf("\n%sPeople:%s%-57s%sDates:%s\n",
		format.Color.Ul, format.Color.Nocolor, " ", format.Color.Ul, format.Color.Nocolor)
	fmt.Printf("Assignee:          %-45sCreated: %s\n",
		issue.Fields.Assignee.DisplayName+" ("+issue.Fields.Assignee.Name+")",
		issue.Fields.Created[:16]) // Truncated at minutes
	fmt.Printf("Reporter:          %-45sUpdated: %s\n",
		issue.Fields.Reporter.DisplayName+" ("+issue.Fields.Reporter.Name+")",
		issue.Fields.Updated[:16]) // Truncated at minutes

	// ******************************************************************
	fmt.Printf("\n%sTime Tracking:%s\n", format.Color.Ul, format.Color.Nocolor)
	fmt.Printf("Estimated: %-25sLogged: %-20sRemaining: %s\n",
		format.TimeEstimate(issue.Fields.TimeTracking.Estimate),
		issue.Fields.TimeTracking.TimeSpent, issue.Fields.TimeTracking.Remaining)

	// ******************************************************************
	fmt.Printf("\n%sDescription:%s\n%s\n", format.Color.Ul, format.Color.Nocolor, issue.Fields.Description)

	// ******************************************************************
	printIssueLinks(issue)

	// ******************************************************************
	if len(issue.Fields.Comment.Comments) > 0 {
		fmt.Printf("\n%sLatest comments:%s\n", format.Color.Ul, format.Color.Nocolor)
		printComments(issue.Fields.Comment.Comments, 3)
	}
}

func printIssueLinks(issue types.IssueDescription) {
	outward := make(map[string][]string)
	inward := make(map[string][]string)

	for _, link := range issue.Fields.IssueLinks {
		var summary string
		if link.OutwardIssue.Key == "" {
			summary = link.InwardIssue.Fields.Summary
			if len(summary) > 42 {
				summary = summary[:42] + ".."
			}

			inward[link.Type.Inward] = append(inward[link.Type.Inward], fmt.Sprintf(
				"%s%-15s%-45s%s%s\n",
				format.IssueType(link.InwardIssue.Fields.IssueType.Name, true),
				link.InwardIssue.Key,
				summary,
				format.Priority(link.InwardIssue.Fields.Priority.Name, true),
				format.Status(link.InwardIssue.Fields.Status.Name, true)))
		} else {
			summary = link.OutwardIssue.Fields.Summary
			if len(summary) > 42 {
				summary = summary[:42] + ".."
			}

			outward[link.Type.Outward] = append(outward[link.Type.Outward], fmt.Sprintf(
				"%s%-15s%-45s%s%s\n",
				format.IssueType(link.OutwardIssue.Fields.IssueType.Name, true),
				link.OutwardIssue.Key,
				summary,
				format.Priority(link.OutwardIssue.Fields.Priority.Name, true),
				format.Status(link.OutwardIssue.Fields.Status.Name, true)))
		}
	}

	for k, v := range outward {
		fmt.Printf("\n%s%s:%s\n", format.Color.Ul, strings.ToTitle(k), format.Color.Nocolor)

		for _, l := range v {
			fmt.Print(l)
		}
	}

	for k, v := range inward {
		fmt.Printf("\n%s%s:%s\n", format.Color.Ul, strings.ToTitle(k), format.Color.Nocolor)

		for _, l := range v {
			fmt.Print(l)
		}
	}
}
