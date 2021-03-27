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
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const describeUsage string = `
By default the the active issue will be described,
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
			issueKey = strings.ToUpper(args[0])
		}
		validateIssueKey(&issueKey)
		issue := getIssue(issueKey)

		var epic IssueDescriptionResponse
		if issue.Fields.Epic != "" {
			epic = getIssue(issue.Fields.Epic)
		}

		printIssue(issue, epic)

	},
}

func init() {
	rootCmd.AddCommand(describeCmd)

	describeCmd.SetUsageTemplate(describeUsage)
}

func getIssue(key string) IssueDescriptionResponse {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key)

	jsonResponse := &IssueDescriptionResponse{}

	getJSONResponse("GET", url, nil, jsonResponse)

	return *jsonResponse
}

func printIssue(issue, epic IssueDescriptionResponse) {
	fmt.Println()
	fmt.Println(formatHeader(issue.Fields.Project.Name, issue.Key, issue.Fields.Summary))
	fmt.Printf("%sDetails:%s\n", color.ul, color.nocolor)
	fmt.Printf("Type:              %sStatus:      %s\n",
		formatIssueType(issue.Fields.IssueType.Name, false), formatStatus(issue.Fields.Status.Name, false))
	fmt.Printf("Priority:          %sResolution:  %s\n",
		formatPriority(issue.Fields.Priority.Name, false), issue.Fields.Resolution.Name)
	fmt.Printf("Labels:            %s\n", strings.Join(issue.Fields.Labels, ", "))
	fmt.Printf("Fixed Version/s:   %s\n", formatFixVersions(issue))
	fmt.Printf("Visibility:        %s\n", issue.Fields.ChangeVisibility.Value)

	if epic.Fields.Summary != "" {
		fmt.Printf("Epic:              %s\n", formatEpic(epic.Fields.Summary))
	}
	// ******************************************************************
	fmt.Printf("\n%sPeople:%s%-57s%sDates:%s\n",
		color.ul, color.nocolor, " ", color.ul, color.nocolor)
	fmt.Printf("Assignee:          %-45sCreated: %s\n",
		issue.Fields.Assignee.DisplayName+" ("+issue.Fields.Assignee.Name+")",
		issue.Fields.Created[:16]) // Truncated at minutes
	fmt.Printf("Reporter:          %-45sUpdated: %s\n",
		issue.Fields.Reporter.DisplayName+" ("+issue.Fields.Reporter.Name+")",
		issue.Fields.Updated[:16]) // Truncated at minutes

	// ******************************************************************
	fmt.Printf("\n%sTime Tracking:%s\n", color.ul, color.nocolor)
	fmt.Printf("Estimated: %-25sLogged: %-20sRemaining: %s\n",
		formatTimeEstimate(issue.Fields.TimeTracking.Estimate),
		issue.Fields.TimeTracking.TimeSpent, issue.Fields.TimeTracking.Remaining)

	// ******************************************************************
	fmt.Printf("\n%sDescription:%s\n%s\n", color.ul, color.nocolor, issue.Fields.Description)

	// ******************************************************************
	printIssueLinks(issue)

	// ******************************************************************
	if len(issue.Fields.Comment.Comments) > 0 {
		fmt.Printf("\n%sLatest comments:%s\n", color.ul, color.nocolor)
		printComments(issue.Fields.Comment, 3)
	}
}

func printIssueLinks(issue IssueDescriptionResponse) {
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
				formatIssueType(link.InwardIssue.Fields.IssueType.Name, true),
				link.InwardIssue.Key,
				summary,
				formatPriority(link.InwardIssue.Fields.Priority.Name, true),
				formatStatus(link.InwardIssue.Fields.Status.Name, true)))
		} else {
			summary = link.OutwardIssue.Fields.Summary
			if len(summary) > 42 {
				summary = summary[:42] + ".."
			}

			outward[link.Type.Outward] = append(outward[link.Type.Outward], fmt.Sprintf(
				"%s%-15s%-45s%s%s\n",
				formatIssueType(link.OutwardIssue.Fields.IssueType.Name, true),
				link.OutwardIssue.Key,
				summary,
				formatPriority(link.OutwardIssue.Fields.Priority.Name, true),
				formatStatus(link.OutwardIssue.Fields.Status.Name, true)))
		}
	}

	for k, v := range outward {
		fmt.Printf("\n%s%s:%s\n", color.ul, strings.Title(k), color.nocolor)

		for _, l := range v {
			fmt.Print(l)
		}
	}

	for k, v := range inward {
		fmt.Printf("\n%s%s:%s\n", color.ul, strings.Title(k), color.nocolor)

		for _, l := range v {
			fmt.Print(l)
		}
	}
}

func formatHeader(project, key, summary string) string {
	header := fmt.Sprintf("%s%s%s%s / %s - %s%s",
		color.bold, color.ul, color.blue, project, key, summary, color.nocolor)

	// If possible try and center the header within a page width of 100 char
	// The - 12 is the spacing of the invisible color chars added above
	l := len(summary)
	if ((100-l)/2 - 12) <= 0 {
		return header
	}

	s := strings.Repeat(" ", (100-l)/2-12)

	return s + header
}

func formatEpic(summary string) string {
	return fmt.Sprintf("%s%s%s", color.magenta, summary, color.nocolor)
}

func formatIssueType(issueType string, short bool) string {
	var col string

	switch issueType {
	case "Improvement":
		col = color.green
	case "Task":
		col = color.blue
	case "Bug":
		col = color.red
	case "Epic":
		col = color.magenta
	}

	if short {
		return fmt.Sprintf("%s%-12s%s", col, issueType, color.nocolor)
	}

	return fmt.Sprintf("%s%-45s%s", col, issueType, color.nocolor)
}

func formatStatus(status string, short bool) string {
	var col string

	switch status {
	case "Closed", "Resolved":
		col = color.green
	case "Ready for Test":
		col = color.cyan
	case "Peer Review", "To Be Fixed", "Programmed", "In Progress", "Accepted":
		col = color.blue
	case "Rejected":
		col = color.red
	}

	if short {
		return fmt.Sprintf("%s%-10s%s", col, status, color.nocolor)
	}

	return fmt.Sprintf("%s%-20s%s", col, status, color.nocolor)
}

func formatPriority(priority string, short bool) string {
	var col string

	switch priority {
	case "Low":
		col = color.green
	case "Normal":
		col = color.blue
	case "Critical":
		col = color.red
	case "High":
		col = color.red
	case "Blocker":
		col = color.red
	}

	if short {
		return fmt.Sprintf("%s%-10s%s", col, priority, color.nocolor)
	}

	return fmt.Sprintf("%s%-45s%s", col, priority, color.nocolor)
}

func formatTimeEstimate(estimate string) string {
	if estimate == "" {
		return "Not Specified"
	}

	return estimate
}

func formatFixVersions(issue IssueDescriptionResponse) string {
	fixVersions := ""
	for _, v := range issue.Fields.FixVersions {
		fixVersions += ", " + v.Name
	}

	return strings.Replace(fixVersions, ", ", "", 1)
}
