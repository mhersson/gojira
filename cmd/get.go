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

const getSprintUsage string = `
Get the current active sprint and all it's issues.

Usage:
  gojira get sprint [NAME OF BOARD]

Aliases:
  sprint

Flags:
  -h, --help                   help for sprint
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
		printIssues(myIssues, true)
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

var getActiveBoardCmd = &cobra.Command{
	Use:     "board",
	Short:   "Display the active board",
	Args:    cobra.NoArgs,
	Aliases: []string{"b"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Active Board: %s\n", getActiveBoard())
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
			if config.UseTimesheetPlugin {
				worklogs := getTimesheet(date)

				printTimesheet(date, worklogs)
			} else {
				issues := getIssues("worklogDate = " + date +
					" AND worklogAuthor = currentUser()")

				myIssues := getUserTimeOnIssueAtDate(config.Username, date, issues)

				printMyWorklog(myIssues)

			}
		}
	},
}

var getSprintCMD = &cobra.Command{
	Use:   "sprint",
	Short: "Display current sprint",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var board string
		if len(args) >= 1 {
			board = args[0]
		} else {
			board = getActiveBoard()
		}
		rapidView := getRapidViewID(board)
		if rapidView != nil && rapidView.SprintSupportEnabled {
			sprints := getSprints(rapidView.ID)
			for _, sprint := range getActiveOrLatestSprint(sprints) {
				contents := getSprintIssues(rapidView.ID, sprint.ID)

				issueTypes := getIssueTypes()

				priorities := getPriorities()

				fmt.Println(formatSprintHeader(*sprint))

				printSprintIssues("Not completed", contents.IssuesNotCompletedInCurrentSprint, *issueTypes, priorities)
				printSprintIssues("Completed", contents.CompletedIssues, *issueTypes, priorities)
				printSprintIssues("Completed in another sprint", contents.IssuesCompletedInAnotherSprint, *issueTypes, priorities)

			}
		} else {
			fmt.Printf("%s does not exist or sprint support is not enabled\n", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getAllIssuesCmd)
	getCmd.AddCommand(getActiveCmd)
	getCmd.AddCommand(getActiveBoardCmd)
	getCmd.AddCommand(getStatusCmd)
	getCmd.AddCommand(getTransistionsCmd)
	getCmd.AddCommand(getCommentsCmd)
	getCmd.AddCommand(getWorklogCmd)
	getCmd.AddCommand(getMyWorklogCmd)
	getCmd.AddCommand(getSprintCMD)

	getAllIssuesCmd.Flags().StringVarP(&jqlFilter,
		"filter", "f", "", "write your own jql filter")

	getAllIssuesCmd.SetUsageTemplate(getAllIssuesUsage)
	getCommentsCmd.SetUsageTemplate(getCommentsUsage)
	getWorklogCmd.SetUsageTemplate(getWorklogUsage)

	getMyWorklogCmd.SetUsageTemplate(myWorklogUsage)

	getSprintCMD.SetUsageTemplate(getSprintUsage)
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
	if _, err := os.Stat(issueFile); os.IsNotExist(err) {
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

func getActiveBoard() string {
	if _, err := os.Stat(boardFile); os.IsNotExist(err) {
		fmt.Println("Active board is not set")
		os.Exit(0)
	}

	out, err := ioutil.ReadFile(boardFile)
	if err != nil {
		fmt.Println("Failed to get active board")
		os.Exit(1)
	}

	return string(out)
}

func getIssues(filter string) []Issue {
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

	jsonResponse := new(struct {
		Issues []Issue `json:"issues"`
	})

	getJSONResponse("POST", url, payload, jsonResponse)

	return jsonResponse.Issues
}

func getTimesheet(date string) []Timesheet {
	url := config.JiraURL + "/rest/timesheet-gadget/1.0/raw-timesheet.json?startDate=" + date + "&endDate=" + date

	jsonResponse := new(struct {
		Worklog []Timesheet `json:"worklog"`
	})

	getJSONResponse(http.MethodGet, url, nil, jsonResponse)

	return jsonResponse.Worklog
}

func getStatus(key string) string {
	jsonResponse := getIssues("key = " + key)
	if len(jsonResponse) != 1 {
		fmt.Printf("Issue %s does not exist\n", key)
		os.Exit(1)
	}

	return jsonResponse[0].Fields.Status.Name
}

func getSummary(key string) string {
	issues := getIssues("key = " + key)
	if len(issues) != 1 {
		fmt.Printf("Issue %s does not exist\n", key)
		os.Exit(1)
	}

	return issues[0].Fields.Summary
}

func getTransistions(key string) []Transition {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/transitions"

	jsonResponse := new(struct {
		Transitions []Transition `json:"transitions"`
	})

	getJSONResponse("GET", url, nil, jsonResponse)

	return jsonResponse.Transitions
}

func getComments(key string) []Comment {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/comment"

	jsonResponse := new(struct {
		Comments []Comment `json:"comments"`
	})

	getJSONResponse("GET", url, nil, jsonResponse)

	return jsonResponse.Comments
}

func getComment(key, commentID string) Comment {
	comments := getComments(key)

	if commentID == "" && len(comments) >= 1 {
		return comments[len(comments)-1]
	}

	for _, c := range comments {
		if c.ID == commentID {
			return c
		}
	}

	return Comment{}
}

func getWorklogs(key string) []Worklog {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/worklog"

	jsonResponse := new(struct {
		Worklogs []Worklog `json:"worklogs"`
	})

	getJSONResponse("GET", url, nil, jsonResponse)

	return jsonResponse.Worklogs
}

func getRapidViewID(board string) *RapidView {
	url := config.JiraURL + "/rest/greenhopper/1.0/rapidview"

	resp := new(struct {
		Views []RapidView `json:"views"`
	})

	getJSONResponse(http.MethodGet, url, nil, resp)

	for _, x := range resp.Views {
		if strings.EqualFold(board, x.Name) {
			return &x
		}
	}

	return nil
}

func getSprints(rapidViewID int) []Sprint {
	url := fmt.Sprintf(
		"%s/rest/greenhopper/latest/sprintquery/%d?includeHistoricSprints=true&includeFutureSprints=true",
		config.JiraURL, rapidViewID)

	resp := new(struct {
		Sprints []Sprint `json:"sprints"`
	})

	getJSONResponse(http.MethodGet, url, nil, resp)

	return resp.Sprints
}

func getActiveOrLatestSprint(sprints []Sprint) []*Sprint {
	active := []*Sprint{}

	for x := range sprints {
		if sprints[x].State == "ACTIVE" {
			active = append(active, &sprints[x])
		}
	}

	if len(active) > 0 {
		return active
	}

	// If none of the sprints are active return the most recent
	if len(sprints) > 0 {
		active = append(active, &sprints[len(sprints)-1])

		return active
	}

	return active
}

func getSprintIssues(rapidViewID, sprintID int) *SprintContent {
	url := fmt.Sprintf("%s/rest/greenhopper/latest/rapid/charts/sprintreport?rapidViewId=%d&sprintId=%d",
		config.JiraURL, rapidViewID, sprintID)

	resp := new(struct {
		Contents SprintContent `json:"contents"`
	})

	getJSONResponse(http.MethodGet, url, nil, resp)

	return &resp.Contents
}

func getUserTimeOnIssueAtDate(user, date string, issues []Issue) []TimeSpentUserIssue {
	userIssues := []TimeSpentUserIssue{}

	for _, v := range issues {
		t := getTimeSpentOnIssue(user, date, v.Key)

		i := &TimeSpentUserIssue{}
		i.Key = v.Key
		i.Date = date
		i.Summary = v.Fields.Summary
		i.TimeSpent = convertSecondsToHoursAndMinutes(t, false)
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

	for _, l := range wl {
		if l.Author.Name == user && strings.HasPrefix(l.Started, date) {
			timeSpent += l.TimeSpentSeconds
		}
	}

	return timeSpent
}

func convertSecondsToHoursAndMinutes(seconds int, dropMinutes bool) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	if dropMinutes {
		return fmt.Sprintf("%dh", hours)
	}

	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func getCurrentDate() string {
	now := time.Now().UTC()
	// jira date format - "2017-12-07"
	return now.Format("2006-01-02")
}

func getIssueTypeNameByID(issueTypes []IssueType, id string) string {
	for _, x := range issueTypes {
		if x.ID == id {
			return x.Name
		}
	}

	return "Unknown"
}

func getPriorityNameByID(priorities []Priority, id string) string {
	for _, x := range priorities {
		if x.ID == id {
			return x.Name
		}
	}

	return "Unknown"
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(string(body))

	err = json.Unmarshal(body, jsonResponse)
	if err != nil {
		log.Fatalf("Failed to parse json response: %s\n", err)
	}

	defer resp.Body.Close()
}

func printIssues(jsonResponse []Issue, header bool) {
	if header {
		fmt.Printf("%s%s\n%-15s%-12s%-10s%-64s%-20s%-15s%s\n", color.ul, color.yellow,
			"Key", "Type", "Priority", "Summary", "Status", "Assignee", color.nocolor)
	}

	for _, v := range jsonResponse {
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

func printTransitions(transitions []Transition) {
	fmt.Println("The following transitions are available:")

	for i, v := range transitions {
		fmt.Printf("%s%s%d.%s %s\n", color.bold, color.yellow, i, color.nocolor, v.Name)
	}
}

func printComments(comments []Comment, maxNumber int) {
	c := comments
	if len(comments) >= maxNumber && maxNumber != 0 {
		c = comments[len(comments)-maxNumber:]
	}

	for _, v := range c {
		fmt.Printf("%sComment:    %s%-45sCreated: %s\n", color.yellow, color.nocolor, v.ID, v.Created[:16])
		fmt.Printf("Visibility: %-45sAuthor: %s (%s)\n", v.Visibility.Value, v.Author.DisplayName, v.Author.Name)
		fmt.Printf("\n%s", strings.ReplaceAll(v.Body, "{noformat}", "```"))
		fmt.Println("\n" + color.ul + strings.Repeat(" ", 100) + color.nocolor)
	}
}

func printWorklogs(issueKey string, worklogs []Worklog) {
	totalTimeSpent := 0

	for _, v := range worklogs {
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
			convertSecondsToHoursAndMinutes(total, false), color.nocolor)
	} else {
		fmt.Println("You have not logged any hours on this date")
	}
}

func printTimesheet(date string, worklogs []Timesheet) {
	if len(worklogs) >= 1 {
		fmt.Printf("%s%s\n%-12s%-15s%-64s%s%s\n", color.ul, color.yellow,
			"Date", "Key", "Summary", "Time Spent", color.nocolor)

		total := 0

		for _, wl := range worklogs {
			if len(wl.Summary) > 60 {
				wl.Summary = wl.Summary[:60] + ".."
			}

			secs := 0
			for _, entry := range wl.Entries {
				secs += entry.TimeSpent
			}

			fmt.Printf("%-12s%-15s%-64s%s\n", date, wl.Key, wl.Summary, convertSecondsToHoursAndMinutes(secs, false))
			total += secs
		}

		fmt.Printf("%s%sTotal time spent:%s %s%s\n",
			strings.Repeat(" ", 73), color.ul, color.nocolor,
			convertSecondsToHoursAndMinutes(total, false), color.nocolor)
	} else {
		fmt.Println("You have not logged any hours on this date")
	}
}

func formatSprintHeader(sprint Sprint) string {
	var statusColor string

	switch sprint.State {
	case "ACTIVE":
		statusColor = color.green
	case "CLOSED":
		statusColor = color.red
	default:
		statusColor = color.blue
	}

	status := fmt.Sprintf("%s(%s%s%s)%s",
		color.cyan, statusColor, sprint.State, color.cyan, color.nocolor)

	return fmt.Sprintf("%s%s%70s %10s", color.bold, color.yellow, sprint.Name, status)
}

func printSprintIssues(header string, issues []SprintIssue, issueTypes []IssueType, priorites []Priority) {
	if len(issues) > 0 {
		fmt.Printf("\n%s%s:%s", color.red, header, color.nocolor)

		fmt.Printf("%s%s\n%-15s%-12s%-10s%-64s%-10s%-10s%-20s%s\n", color.ul, color.yellow,
			"Key", "Type", "Priority", "Summary", "ETA", "Epic", "Assignee", color.nocolor)

		for _, v := range issues {
			if len(v.Summary) >= 60 {
				v.Summary = v.Summary[:60] + ".."
			}

			fmt.Printf("%-15s%s%s%-64s%-10s%-10s%-15s\n",
				v.Key,
				formatIssueType(getIssueTypeNameByID(issueTypes, v.TypeID), true),
				formatPriority(getPriorityNameByID(priorites, v.PriorityID), true),
				v.Summary,
				convertSecondsToHoursAndMinutes(int(v.CurrentEstimateStatistic.StatFieldValue.Value), true),
				v.Epic,
				v.AssigneeName)
		}
	}
}