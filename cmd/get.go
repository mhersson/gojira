/*
Copyright © 2020-2024 Morten Hersson

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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/mhersson/gojira/pkg/jira"
	"github.com/mhersson/gojira/pkg/types"
	"github.com/mhersson/gojira/pkg/util"
	"github.com/mhersson/gojira/pkg/util/convert"
	"github.com/mhersson/gojira/pkg/util/format"
	"github.com/mhersson/gojira/pkg/util/validate"
)

var GetAllSprints bool

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
  gojira get myworklog [yyyy-mm-dd] [flags]
  gojira get myworklog stats [yyyy-mm-dd] [yyyy-mm-dd]

Available Commands:
  stats       Display you worklog statistics

Aliases:
  myworklog, m

Flags:
  -h, --help                   help for myworklog
  -w, --week                   current week (only with timesheet plugin)
`

const myWorklogStatisticsUsage string = `Shows per week worklog statistics for a given period.
Aligns the week numbers to the dates entered,
and calculates the average and total amount of hours per week.


Usage:
  gojira get myworklog stats [yyyy-mm-dd] [yyyy-mm-dd]

Aliases:
  stats, s

Flags:
  -h, --help                   help for myworklog
`

const getSprintUsage string = `
Usage:
  gojira get sprint [NAME OF BOARD]

Aliases:
  sprint

Flags:
  -h, --help                   help for sprint
  -a, --all                    get all sprints (future and  active)
`

const getKanbanBoardUsage string = `
Usage:
  gojira get kanban [NAME OF BOARD]

Aliases:
  kanban

Flags:
  -h, --help                   help for kanban
  -c, --closed                 show closed issues
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
		myIssues := jira.GetIssues(JQLFilter)
		printIssues(myIssues, true, false)
	},
}

var getActiveCmd = &cobra.Command{
	Use:     "active",
	Short:   "Display the active issue, sprint or kanban board",
	Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Aliases: []string{"a"},
}

var getActiveIssueCmd = &cobra.Command{
	Use:     "issue",
	Short:   "Display the active issue",
	Args:    cobra.NoArgs,
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		key := util.GetActiveIssue(IssueFile)
		summary := getSummary(key)
		fmt.Printf("Active issue: %s %s\n", key, summary)
	},
}

var getActiveSprintCmd = &cobra.Command{
	Use:     "sprint",
	Short:   "Display the active sprint",
	Aliases: []string{"s"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Active sprint board: %s\n", util.GetActiveSprintOrKanban(BoardFile, "sprint"))
	},
}

var getActiveKanbanCmd = &cobra.Command{
	Use:     "kanban",
	Short:   "Display the active kanban board",
	Aliases: []string{"k"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Active kanban board: %s\n", util.GetActiveSprintOrKanban(BoardFile, "kanban"))
	},
}

var getStatusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Display the current status",
	Args:    cobra.NoArgs,
	Aliases: []string{"st"},
	Run: func(cmd *cobra.Command, args []string) {
		jira.CheckIssueKey(&IssueKey, IssueFile)
		status := getStatus(IssueKey)
		printStatus(status, false)
	},
}

var getTransistionsCmd = &cobra.Command{
	Use:     "transitions",
	Short:   "Display available transistions",
	Args:    cobra.NoArgs,
	Aliases: []string{"t"},
	Run: func(cmd *cobra.Command, args []string) {
		jira.CheckIssueKey(&IssueKey, IssueFile)
		status := getStatus(IssueKey)
		printStatus(status, false)
		tr := jira.GetTransistions(IssueKey)
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
			IssueKey = strings.ToUpper(args[0])
		}
		jira.CheckIssueKey(&IssueKey, IssueFile)
		comments := jira.GetComments(IssueKey)
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
			IssueKey = strings.ToUpper(args[0])
		}
		jira.CheckIssueKey(&IssueKey, IssueFile)
		worklogs := jira.GetWorklogs(IssueKey)
		printWorklogs(IssueKey, worklogs)
	},
}

var getMyWorklogCmd = &cobra.Command{
	Use:     "myworklog",
	Short:   "Display your worklog for a given date",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"m"},
	Run: func(cmd *cobra.Command, args []string) {
		date := util.GetCurrentDate()
		if len(args) == 1 {
			date = args[0]
		}
		if validate.Date(date) {
			if Cfg.UseTimesheetPlugin {
				ts := jira.GetTimesheet(date, date, ShowEntireWeek)
				if len(ts) == 0 && util.DateIsToday(date) {
					fmt.Println("You havn't logged any hours today.")
					os.Exit(0)
				}

				worklogs := util.GetWorklogsSorted(ts, true)
				printTimesheet(worklogs)
			} else {
				issues := jira.GetIssues("worklogDate = " + date +
					" AND worklogAuthor = currentUser()")

				if len(issues) == 0 && util.DateIsToday(date) {
					fmt.Println("You havn't logged any hours today.")
					os.Exit(0)
				}

				myIssues := getUserTimeOnIssueAtDate(Cfg.Username, date, issues)
				printMyWorklog(myIssues)
			}
		}
	},
}

var getMyWorklogStatistics = &cobra.Command{
	Use:     "stats",
	Short:   "Display your worklog statistics",
	Aliases: []string{"s"},
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if !Cfg.UseTimesheetPlugin {
			fmt.Println("This command is only available with the timesheet plugin")
			os.Exit(1)
		}
		if validate.Date(args[0]) && validate.Date(args[1]) {
			t1, _ := time.Parse("2006-01-02", args[0])
			t2, _ := time.Parse("2006-01-02", args[1])
			if t2.Before(t1) {
				t2, t1 = t1, t2
			}

			fromDate, _ := util.WeekStartEndDate(t1.ISOWeek())
			_, toDate := util.WeekStartEndDate(t2.ISOWeek())

			if t2.Sub(t1).Hours() > (24 * 365) {
				fmt.Println("1 year is the max time period.")
				os.Exit(1)
			}

			ts := jira.GetTimesheet(fromDate, toDate, false)
			if len(ts) == 0 {
				fmt.Printf("You havn't logged any hours between %s - %s\n", args[0], args[1])
				os.Exit(0)
			}

			worklogs := util.GetWorklogsSorted(ts, true)

			if _, err := os.Stat(ConfigFolder); errors.Is(err, os.ErrNotExist) {
				_ = os.Mkdir(ConfigFolder, 0o755)
			}

			publicHolidays := util.LoadPublicHolidays(
				filepath.Join(ConfigFolder, "public-holidays-"+t1.Format("2006")+"-"+Cfg.CountryCode+".json"),
				t1.Format("2006"),
				Cfg.CountryCode)

			weeks := util.GroupWorklogsByWeek(fromDate, toDate, worklogs, util.GetPublicHolidayDates(publicHolidays))

			printStatistics(weeks)
		} else {
			fmt.Println("Invalid date.")
		}
	},
}

var getSprintCmd = &cobra.Command{
	Use:     "sprint",
	Short:   "Display sprint board",
	Aliases: []string{"s"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var board string
		if len(args) >= 1 {
			board = args[0]
		} else {
			board = util.GetActiveSprintOrKanban(BoardFile, "sprint")
		}
		rapidView := jira.GetRapidViewID(board)
		if rapidView != nil && rapidView.SprintSupportEnabled {
			issueTypes := jira.GetIssueTypes()
			priorities := jira.GetPriorities()
			sprints, issues := jira.GetSprints(rapidView.ID)
			for i := range sprints {
				sprint := sprints[i]
				if !sprint.MatchesFilter(Cfg.SprintFilter) {
					continue
				}
				if sprint.State != "ACTIVE" && !GetAllSprints {
					continue
				}
				fmt.Println(format.SprintHeader(sprint))
				printSprintIssues(&sprint, issues, *issueTypes, priorities)
			}
		} else {
			fmt.Printf("%s does not exist or sprint support is not enabled\n", board)
		}
	},
}

var getKanbanBoardCmd = &cobra.Command{
	Use:     "kanban",
	Short:   "Display kanban board",
	Aliases: []string{"k"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var board string
		if len(args) >= 1 {
			board = args[0]
		} else {
			board = util.GetActiveSprintOrKanban(BoardFile, "kanban")
		}

		rapidView := jira.GetRapidViewID(board)
		if rapidView == nil {
			fmt.Printf("Board %s does not exist\n", board)
			os.Exit(1)
		}

		issues := jira.GetKanbanIssues(rapidView.ID)

		fmt.Println(format.KanbanBoardHeader(board))
		if cmd.Flag("closed").Changed {
			printIssues(issues, true, true)
		} else {
			printIssues(issues, true, false)
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
	getCmd.AddCommand(getSprintCmd)
	getCmd.AddCommand(getKanbanBoardCmd)

	getAllIssuesCmd.Flags().StringVarP(&JQLFilter,
		"filter", "f", "", "write your own jql filter")

	getAllIssuesCmd.SetUsageTemplate(getAllIssuesUsage)
	getCommentsCmd.SetUsageTemplate(getCommentsUsage)
	getWorklogCmd.SetUsageTemplate(getWorklogUsage)

	getActiveCmd.AddCommand(getActiveIssueCmd)
	getActiveCmd.AddCommand(getActiveSprintCmd)
	getActiveCmd.AddCommand(getActiveKanbanCmd)

	getMyWorklogCmd.SetUsageTemplate(myWorklogUsage)
	getMyWorklogCmd.Flags().BoolVarP(&ShowEntireWeek, "week", "w", false, "view current week (only with timesheet plugin)")
	getMyWorklogCmd.AddCommand(getMyWorklogStatistics)

	getMyWorklogStatistics.SetUsageTemplate(myWorklogStatisticsUsage)

	getSprintCmd.SetUsageTemplate(getSprintUsage)
	getSprintCmd.Flags().BoolVarP(&GetAllSprints, "all", "a", false, "get all sprints")

	getKanbanBoardCmd.SetUsageTemplate(getKanbanBoardUsage)
	getKanbanBoardCmd.Flags().BoolP("closed", "c", false, "Show closed issues")
}

func getStatus(key string) string {
	jsonResponse := jira.GetIssues("key = " + key)
	if len(jsonResponse) != 1 {
		fmt.Printf("Issue %s does not exist\n", key)
		os.Exit(1)
	}

	return jsonResponse[0].Fields.Status.Name
}

func getSummary(key string) string {
	issues := jira.GetIssues("key = " + key)
	if len(issues) != 1 {
		fmt.Printf("Issue %s does not exist\n", key)
		os.Exit(1)
	}

	return issues[0].Fields.Summary
}

func getUserTimeOnIssueAtDate(user, date string, issues []types.Issue) []types.TimeSpentUserIssue {
	userIssues := []types.TimeSpentUserIssue{}

	for _, v := range issues {
		t := getTimeSpentOnIssue(user, date, v.Key)

		i := &types.TimeSpentUserIssue{}
		i.ID = v.ID
		i.Key = v.Key
		i.Date = date
		i.Summary = v.Fields.Summary
		i.TimeSpent = convert.SecondsToHoursAndMinutes(t, false)
		i.TimeSpentSeconds = t
		userIssues = append(userIssues, *i)
	}

	return userIssues
}

func getTimeSpentOnIssue(user, date string, key string) int {
	// Returns the number of hours and minutes a user
	// has logged on an issue on the given date as total
	// number of seconds
	wl := jira.GetWorklogs(key)

	timeSpent := 0

	for _, l := range wl {
		if l.Author.Name == user && strings.HasPrefix(l.Started, date) {
			timeSpent += l.TimeSpentSeconds
		}
	}

	return timeSpent
}

func getIssueTypeNameByID(issueTypes []types.IssueType, id string) string {
	for _, x := range issueTypes {
		if x.ID == id {
			return x.Name
		}
	}

	return "Unknown"
}

func getPriorityNameByID(priorities []types.Priority, id string) string {
	for _, x := range priorities {
		if x.ID == id {
			return x.Name
		}
	}

	return "Unknown"
}

func printIssues(issues []types.Issue, header bool, printClosed bool) {
	if header {
		fmt.Printf("%s%s\n%-15s%-12s%-10s%-64s%-20s%-15s%s\n", format.Color.Ul, format.Color.Yellow,
			"Key", "Type", "Priority", "Summary", "Status", "Assignee", format.Color.Nocolor)
	}

	for _, v := range issues {
		if len(v.Fields.Summary) >= 60 {
			v.Fields.Summary = v.Fields.Summary[:60] + ".."
		}

		if !printClosed && slices.Contains([]string{"Closed", "Resolved", "Verified"}, v.Fields.Status.Name) {
			continue
		}

		fmt.Printf("%-15s%s%s%-64s%s%s\n",
			v.Key,
			format.IssueType(v.Fields.IssueType.Name, true),
			format.Priority(v.Fields.Priority.Name, true),
			v.Fields.Summary,
			format.Status(v.Fields.Status.Name, false),
			v.Fields.Assignee.DisplayName)
	}
}

func printStatus(status string, hasBeenUpdated bool) {
	if hasBeenUpdated {
		fmt.Printf("\n%s%sNew status:%s %s%s\n",
			format.Color.Yellow, format.Color.Bold, format.Color.Green, status, format.Color.Nocolor)
	} else {
		fmt.Printf("\n%s%sCurrent status:%s %s%s\n",
			format.Color.Yellow, format.Color.Bold, format.Color.Green, status, format.Color.Nocolor)
	}
}

func printTransitions(transitions []types.Transition) {
	fmt.Println("The following transitions are available:")

	for i, v := range transitions {
		fmt.Printf("%s%s%d.%s %s\n", format.Color.Bold, format.Color.Yellow, i, format.Color.Nocolor, v.Name)
	}
}

func printComments(comments []types.Comment, maxNumber int) {
	c := comments
	if len(comments) >= maxNumber && maxNumber != 0 {
		c = comments[len(comments)-maxNumber:]
	}

	for _, v := range c {
		fmt.Printf("%sComment:    %s%-45sCreated: %s\n", format.Color.Yellow, format.Color.Nocolor, v.ID, v.Created[:16])
		fmt.Printf("Visibility: %-45sAuthor: %s (%s)\n", v.Visibility.Value, v.Author.DisplayName, v.Author.Name)
		fmt.Printf("\n%s", strings.ReplaceAll(v.Body, "{noformat}", "```"))
		fmt.Println("\n" + format.Color.Ul + strings.Repeat(" ", 100) + format.Color.Nocolor)
	}
}

func printWorklogs(issueKey string, worklogs []types.Worklog) {
	totalTimeSpent := 0

	for _, v := range worklogs {
		totalTimeSpent += v.TimeSpentSeconds

		fmt.Printf("%s %s%-30s%sTime Spent: %s%-8s%s%s\n",
			v.Started[:16],
			format.Color.Cyan, v.Author.DisplayName, format.Color.Nocolor,
			format.Color.Yellow, v.TimeSpent, format.Color.Nocolor, v.Comment)
	}

	if totalTimeSpent == 0 {
		fmt.Println("No work has been logged on this issue")
	} else {
		printTimeTracking(issueKey)
	}
}

func printTimeTracking(key string) {
	issue := jira.GetIssue(key)

	colorRemaining := format.Color.Yellow
	if issue.Fields.TimeTracking.Remaining == "0h" && issue.Fields.TimeTracking.Estimate != "" {
		colorRemaining = format.Color.Red
	}

	fmt.Printf("%sTotal time spent:%s %-9s%sEstimated:%s %-9s%sRemaining:%s %s\n",
		format.Color.Green, format.Color.Nocolor, format.TimeEstimate(issue.Fields.TimeTracking.TimeSpent),
		format.Color.Blue, format.Color.Nocolor, issue.Fields.TimeTracking.Estimate,
		colorRemaining, format.Color.Nocolor, issue.Fields.TimeTracking.Remaining)
}

func printMyWorklog(ti []types.TimeSpentUserIssue) {
	if len(ti) >= 1 {
		fmt.Printf("%s%s\n%-12s%-15s%-64s%s%s\n", format.Color.Ul, format.Color.Yellow,
			"Date", "Key", "Summary", "Time Spent", format.Color.Nocolor)

		total := 0

		for _, v := range ti {
			if len(v.Summary) > 60 {
				v.Summary = v.Summary[:60] + ".."
			}

			fmt.Printf("%-12s%-15s%-64s%s\n", v.Date, v.Key, v.Summary, v.TimeSpent)
			total += v.TimeSpentSeconds
		}

		fmt.Printf("%s%sTotal time spent:%s %s%s\n",
			strings.Repeat(" ", 73), format.Color.Ul, format.Color.Nocolor,
			convert.SecondsToHoursAndMinutes(total, false), format.Color.Nocolor)
	} else {
		fmt.Println("You have not logged any hours on this date")
	}
}

func printTimesheet(worklogs []types.SimplifiedTimesheet) {
	if len(worklogs) >= 1 {
		fmt.Printf("%s%s\n%-11s%-7s%-15s%-44s%-33s%9s%s\n", format.Color.Ul, format.Color.Yellow,
			"Date", "Time", "Key", "Summary", "Comment", "Time Spent", format.Color.Nocolor)

		total := 0
		for _, w := range worklogs {
			total += w.TimeSpent
			fmt.Printf("%-18s%-15s%-44s%-33s%9s\n",
				w.StartDate, w.Key, w.Summary, w.Comment, convert.SecondsToHoursAndMinutes(w.TimeSpent, false))
		}

		fmt.Printf("%s%sTotal time spent: %11s%s\n",
			strings.Repeat(" ", 90), format.Color.Ul,
			convert.SecondsToHoursAndMinutes(total, false), format.Color.Nocolor)
	} else {
		fmt.Println("You have not logged any hours on this date")
	}
}

func printStatistics(weeks []types.Week) {
	if len(weeks) > 0 {
		fmt.Printf("%s%s\n%-9s%-11s%-12s%-12s%-12s%-5s%10s%s\n", format.Color.Ul, format.Color.Yellow,
			"Week#", "Start", "End", "Workdays", "Holidays", "Average", "Total", format.Color.Nocolor)

		var weeksTotal float64

		for _, week := range weeks {
			avg := format.StatsAverage(week.Average(), Cfg.WorkingHoursPerDay)
			tot := format.StatsTotal(week.TotalTime(), Cfg.WorkingHoursPerWeek, Cfg.WorkingHoursPerDay, week.PublicHolidays)
			days := format.StatsWorkdays(week.WorkDays(), Cfg.NumWorkingDays, week.PublicHolidays)
			holidays := format.StatsHolidays(week.PublicHolidays)

			fmt.Printf(" %-8d%-10s%-16s%-21s%-20s%-15s%18s\n",
				week.Number(), week.StartDate.Format("01/02"), week.EndDate.Format("01/02"),
				days, holidays, avg, tot)

			weeksTotal += week.TotalTime()
		}

		printStatisticsSummary(len(weeks), weeksTotal)
	} else {
		fmt.Println("There are no hours registered for this period")
	}
}

func printStatisticsSummary(numWeeks int, weeksTotal float64) {
	expectedTotal := Cfg.WorkingHoursPerWeek * float64(numWeeks)
	totalSummary := weeksTotal - expectedTotal

	if totalSummary >= 0 {
		fmt.Printf("\nYou are %s hours ahead of the expected %s hours total for this period\n",
			format.StatsSummary(totalSummary), format.StatsSummary(expectedTotal))
	} else {
		fmt.Printf("\nYou are %s hours short of the expected %s hours total for this period\n",
			format.StatsSummary(totalSummary), format.StatsSummary(expectedTotal))
	}
}

func printSprintIssues(
	sprint *types.Sprint, issues []types.SprintIssue, issueTypes []types.IssueType, priorites []types.Priority,
) {
	if len(issues) > 0 {
		fmt.Printf("%s%s\n%-15s%-12s%-10s%-64s%-10s%-10s%-6s%-20s%s\n", format.Color.Ul, format.Color.Yellow,
			"Key", "Type", "Priority", "Summary", "Est.", "Epic", "Done", "Assignee", format.Color.Nocolor)

		for _, i := range sprint.IssuesIDs {
			for _, v := range issues {
				if v.ID == i {
					if len(v.Summary) >= 60 {
						v.Summary = v.Summary[:60] + ".."
					}

					fmt.Printf("%-15s%s%s%-64s%-10s%-10s%-15s%-20s\n",
						v.Key,
						format.IssueType(getIssueTypeNameByID(issueTypes, v.TypeID), true),
						format.Priority(getPriorityNameByID(priorites, v.PriorityID), true),
						v.Summary,
						convert.SecondsToHoursAndMinutes(int(v.EstimateStatistic.StatFieldValue.Value), true),
						v.Epic,
						format.SprintStatus(v.Done),
						v.AssigneeName,
					)

					break
				}
			}
		}
	}
}
