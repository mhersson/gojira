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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const getAllIssuesUsage string = `This command will by default display all unresolved
issues assinged to you, but by using the --filter flag
you can compose your own jql filter. All query results,
default as well as custom ones, will be sorted by priority
and their latest update time.

Usage:
  gojira get all [flags]

Aliases:
  all, l

Flags:
  -f, --filter [JQL FILTER]    write your own jql filter
  -h, --help                   help for all

Examples:
  # Display all issues assigned to you (default)
  gojira get all

  # Display all issues with label = toolsmith, status != Ready for review
  gojira get all -f "labels = toolsmith and status != 'Ready for review'"

  # All open issues on project OSE
  gojira get all -f "project = OSE and resolution = unresolved"

`
const getCommentsUsage string = `
By default the comments from the active issue is displayed,
but this can be changed by adding the issue key as argument.

Usage:
  gojira get comment [ISSUE KEY] [flags]

Aliases:
  comment, c

Flags:
  -h, --help                   help for comment
`

const getWorklogUsage string = `
By default the worklog from the active issue is displayed,
but this can be changed by adding the issue key as argument.

Usage:
  gojira get worklog [ISSUE KEY] [flags]

Aliases:
  worklog, w

Flags:
  -h, --help                   help for worklog
`

const myWorklogUsage string = `This command will show the issues you have worked on
and the hours you have logged on a given date.

Usage:
  gojira myworklog [yyyy-mm-dd] [flags]

Aliases:
  myworklog

Flags:
  -h, --help                   help for myworklog
`

// getCmd represents the get command.
var getCmd = &cobra.Command{
	Use:     "get",
	Short:   "Display one or many resources",
	Aliases: []string{"g"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("You must specify the type of resource to get")
	},
}

var getAllIssuesCmd = &cobra.Command{
	Use:     "all",
	Short:   "Display all issues assigned to you",
	Args:    cobra.NoArgs,
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {
		myIssues := getIssues(jqlFilter)
		printIssues(myIssues)
	},
}

var getActiveCmd = &cobra.Command{
	Use:     "active",
	Short:   "Display the active issue",
	Args:    cobra.NoArgs,
	Aliases: []string{"a"},
	Run: func(cmd *cobra.Command, args []string) {
		key := getActiveIssue()
		summary := getSummary(key)
		fmt.Printf("Active Issue: %s %s\n", key, summary)
	},
}

var getStatusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Display the current status",
	Args:    cobra.NoArgs,
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		validateIssueKey(&issueKey)
		status := getStatus(issueKey)
		printStatus(status, false)
	},
}

var getTransistionsCmd = &cobra.Command{
	Use:     "transitions",
	Short:   "Display available transistions",
	Args:    cobra.NoArgs,
	Aliases: []string{"t"},
	Run: func(cmd *cobra.Command, args []string) {
		validateIssueKey(&issueKey)
		status := getStatus(issueKey)
		printStatus(status, false)
		tr := getTransistions(issueKey)
		printTransitions(tr)
	},
}

var getCommentsCmd = &cobra.Command{
	Use:     "comments",
	Short:   "Display all comments",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			issueKey = strings.ToUpper(args[0])
		}
		validateIssueKey(&issueKey)
		comments := getComments(issueKey)
		printComments(comments, 0)
	},
}

var getWorklogCmd = &cobra.Command{
	Use:     "worklog",
	Short:   "Display the worklog",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"wl", "w"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			issueKey = strings.ToUpper(args[0])
		}
		validateIssueKey(&issueKey)
		worklogs := getWorklogs(issueKey)
		printWorklogs(issueKey, worklogs)
	},
}

var getMyWorklogCmd = &cobra.Command{
	Use:   "myworklog",
	Short: "Display your worklog for a given date",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		date := getCurrentDate()
		if len(args) == 1 {
			date = args[0]
		}
		if validateDate(date) {
			issues := getIssues("worklogDate = " + date +
				" AND worklogAuthor = currentUser()")

			myIssues := getUserTimeOnIssueAtDate(config.Username, date, issues)

			printMyWorklog(myIssues)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getAllIssuesCmd)
	getCmd.AddCommand(getActiveCmd)
	getCmd.AddCommand(getStatusCmd)
	getCmd.AddCommand(getTransistionsCmd)
	getCmd.AddCommand(getCommentsCmd)
	getCmd.AddCommand(getWorklogCmd)
	getCmd.AddCommand(getMyWorklogCmd)

	getAllIssuesCmd.Flags().StringVarP(&jqlFilter,
		"filter", "f", "", "write your own jql filter")

	getAllIssuesCmd.SetUsageTemplate(getAllIssuesUsage)
	getCommentsCmd.SetUsageTemplate(getCommentsUsage)
	getWorklogCmd.SetUsageTemplate(getWorklogUsage)

	getMyWorklogCmd.SetUsageTemplate(myWorklogUsage)
}

func issueExists(issueKey *string) bool {
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

func validateIssueKey(key *string) {
	if *key != "" {
		re := regexp.MustCompile("[A-Z]{2,9}-[0-9]{1,4}")

		m := re.MatchString(*key)
		if !m {
			fmt.Println("Invalid key")
			os.Exit(1)
		}

		if !issueExists(key) {
			fmt.Printf("%s does not exist\n", *key)
			os.Exit(1)
		}
	} else {
		*key = getActiveIssue()
	}
}

func getActiveIssue() string {
	_, err := os.Stat(issueFile)
	if os.IsNotExist(err) {
		fmt.Println("Active issue is not set")
		os.Exit(1)
	}

	out, err := ioutil.ReadFile(issueFile)
	if err != nil {
		fmt.Println("Failed to get active issue")
		os.Exit(1)
	}

	return string(out)
}

func getIssues(filter string) IssuesResponse {
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

	jsonResponse := &IssuesResponse{}

	getJSONResponse("POST", url, payload, &jsonResponse)

	return *jsonResponse
}

func getStatus(key string) string {
	jsonResponse := getIssues("key = " + key)
	if len(jsonResponse.Issues) != 1 {
		fmt.Printf("Issue %s does not exist\n", key)
		os.Exit(1)
	}

	return jsonResponse.Issues[0].Fields.Status.Name
}

func getSummary(key string) string {
	jsonResponse := getIssues("key = " + key)
	if len(jsonResponse.Issues) != 1 {
		fmt.Printf("Issue %s does not exist\n", key)
		os.Exit(1)
	}

	return jsonResponse.Issues[0].Fields.Summary
}

func getTransistions(key string) TransitionsResponse {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/transitions"

	jsonResponse := &TransitionsResponse{}

	getJSONResponse("GET", url, nil, jsonResponse)

	return *jsonResponse
}

func getComments(key string) CommentsResponse {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/comment"

	jsonResponse := &CommentsResponse{}

	getJSONResponse("GET", url, nil, jsonResponse)

	return *jsonResponse
}

func getComment(key, commentID string) CommentResponse {
	comments := getComments(key)

	if commentID == "" && len(comments.Comments) >= 1 {
		return comments.Comments[len(comments.Comments)-1]
	}

	for _, c := range comments.Comments {
		if c.ID == commentID {
			return c
		}
	}

	return CommentResponse{}
}

func getWorklogs(key string) WorklogsResponse {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/worklog"

	jsonResponse := &WorklogsResponse{}

	getJSONResponse("GET", url, nil, jsonResponse)

	return *jsonResponse
}

func getUserTimeOnIssueAtDate(user, date string, issues IssuesResponse) []TimeSpentUserIssue {
	userIssues := []TimeSpentUserIssue{}

	for _, v := range issues.Issues {
		t := getTimeSpentOnIssue(user, date, v.Key)

		i := &TimeSpentUserIssue{}
		i.Key = v.Key
		i.Date = date
		i.Summary = v.Fields.Summary
		i.TimeSpent = convertSecondsToHoursAndMinutes(t)
		i.TimeSpentSeconds = t
		userIssues = append(userIssues, *i)
	}

	return userIssues
}

func getTimeSpentOnIssue(user, date string, key string) int {
	// Returns the number of hours and minutes a user
	// has logged on an issue on the given date as total
	// number of seconds
	wl := getWorklogs(key)

	timeSpent := 0

	for _, l := range wl.Worklogs {
		if l.Author.Name == user && strings.HasPrefix(l.Started, date) {
			timeSpent += l.TimeSpentSeconds
		}
	}

	return timeSpent
}

func convertSecondsToHoursAndMinutes(seconds int) string {
	//  Converts number of seconds to a string on format '2h 0m'
	dur := time.Duration(seconds) * time.Second
	hm := strings.Split(strconv.FormatFloat(dur.Hours(), 'f', 2, 64), ".")
	rest, _ := strconv.ParseFloat(hm[1], 64)
	minutes := (rest / 100) * 60

	return fmt.Sprintf("%sh %0.fm", hm[0], minutes)
}

func getCurrentDate() string {
	now := time.Now().UTC()
	// jira date format - "2017-12-07"
	return now.Format("2006-01-02")
}

func getJSONResponse(method string, url string, payload []byte, jsonResponse interface{}) {
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
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println(string(body))

	err = json.Unmarshal(body, jsonResponse)
	if err != nil {
		log.Fatalf("Failed to parse json response: %s\n", err)
	}
}

func printIssues(jsonResponse IssuesResponse) {
	fmt.Printf("%s%s\n%-15s%-12s%-10s%-64s%-20s%-15s%s\n", color.ul, color.yellow,
		"Key", "Type", "Priority", "Summary", "Status", "Assignee", color.nocolor)

	for _, v := range jsonResponse.Issues {
		if len(v.Fields.Summary) >= 60 {
			v.Fields.Summary = v.Fields.Summary[:60] + ".."
		}

		fmt.Printf("%-15s%s%s%-64s%s%s\n",
			v.Key,
			formatIssueType(v.Fields.IssueType.Name, true),
			formatPriority(v.Fields.Priority.Name, true),
			v.Fields.Summary,
			formatStatus(v.Fields.Status.Name, false),
			v.Fields.Assignee.DisplayName)
	}
}

func printStatus(status string, hasBeenUpdated bool) {
	if hasBeenUpdated {
		fmt.Printf("\n%s%sNew status:%s %s%s\n",
			color.yellow, color.bold, color.green, status, color.nocolor)
	} else {
		fmt.Printf("\n%s%sCurrent status:%s %s%s\n",
			color.yellow, color.bold, color.green, status, color.nocolor)
	}
}

func printTransitions(jsonResponse TransitionsResponse) {
	fmt.Println("The following transitions are available:")

	for i, v := range jsonResponse.Transitions {
		fmt.Printf("%s%s%d.%s %s\n", color.bold, color.yellow, i, color.nocolor, v.Name)
	}
}

func printComments(jsonResponse CommentsResponse, maxNumber int) {
	comments := jsonResponse.Comments
	if len(jsonResponse.Comments) >= maxNumber && maxNumber != 0 {
		comments = jsonResponse.Comments[len(jsonResponse.Comments)-maxNumber:]
	}

	for _, v := range comments {
		fmt.Printf("%sComment:    %s%-45sCreated: %s\n", color.yellow, color.nocolor, v.ID, v.Created[:16])
		fmt.Printf("Visibility: %-45sAuthor: %s (%s)\n", v.Visibility.Value, v.Author.DisplayName, v.Author.Name)
		fmt.Printf("\n%s", strings.ReplaceAll(v.Body, "{noformat}", "```"))
		fmt.Println("\n" + color.ul + strings.Repeat(" ", 100) + color.nocolor)
	}
}

func printWorklogs(issueKey string, jsonResponse WorklogsResponse) {
	totalTimeSpent := 0

	for _, v := range jsonResponse.Worklogs {
		totalTimeSpent += v.TimeSpentSeconds

		fmt.Printf("%s %s%-30s%sTime Spent: %s%-8s%s%s\n",
			v.Started[:16],
			color.cyan, v.Author.DisplayName, color.nocolor,
			color.yellow, v.TimeSpent, color.nocolor, v.Comment)
	}

	if totalTimeSpent == 0 {
		fmt.Println("No work has been logged on this issue")
	} else {
		printTimeTracking(issueKey)
	}
}

func printTimeTracking(key string) {
	issue := getIssue(key)

	colorRemaining := color.yellow
	if issue.Fields.TimeTracking.Remaining == "0h" && issue.Fields.TimeTracking.Estimate != "" {
		colorRemaining = color.red
	}

	fmt.Printf("%sTotal time spent:%s %-9s%sEstimated:%s %-9s%sRemaining:%s %s\n",
		color.green, color.nocolor, formatTimeEstimate(issue.Fields.TimeTracking.TimeSpent),
		color.blue, color.nocolor, issue.Fields.TimeTracking.Estimate,
		colorRemaining, color.nocolor, issue.Fields.TimeTracking.Remaining)
}

func printMyWorklog(ti []TimeSpentUserIssue) {
	if len(ti) >= 1 {
		fmt.Printf("%s%s\n%-12s%-15s%-64s%s%s\n", color.ul, color.yellow,
			"Date", "Key", "Summary", "Time Spent", color.nocolor)

		total := 0

		for _, v := range ti {
			if len(v.Summary) > 60 {
				v.Summary = v.Summary[:60] + ".."
			}

			fmt.Printf("%-12s%-15s%-64s%s\n", v.Date, v.Key, v.Summary, v.TimeSpent)
			total += v.TimeSpentSeconds
		}

		fmt.Printf("%s%sTotal time spent:%s %s%s\n",
			strings.Repeat(" ", 73), color.ul, color.nocolor,
			convertSecondsToHoursAndMinutes(total), color.nocolor)
	} else {
		fmt.Println("You have not logged any hours on this date")
	}
}
