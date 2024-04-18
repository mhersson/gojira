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
package format

import (
	"fmt"
	"strings"

	"github.com/mhersson/gojira/pkg/types"
)

var Color = types.Color{
	Red:     "\033[31m",
	Green:   "\033[32m",
	Yellow:  "\033[33m",
	Blue:    "\033[34m",
	Magenta: "\033[35m",
	Cyan:    "\033[36m",
	Bold:    "\033[1m",
	Ul:      "\033[4m",
	Nocolor: "\033[0m",
}

func Header(project, key, summary string) string {
	header := fmt.Sprintf("%s%s%s%s / %s - %s%s",
		Color.Bold, Color.Ul, Color.Blue, project, key, summary, Color.Nocolor)

	// If possible try and center the header within a page width of 100 char
	// The - 12 is the spacing of the invisible color chars added above
	l := len(summary)
	if ((100-l)/2 - 12) <= 0 {
		return header
	}

	s := strings.Repeat(" ", (100-l)/2-12)

	return s + header
}

func Epic(summary string) string {
	return fmt.Sprintf("%s%s%s", Color.Magenta, summary, Color.Nocolor)
}

func IssueType(issueType string, short bool) string {
	var col string

	switch issueType {
	case "Improvement":
		col = Color.Green
	case "Task":
		col = Color.Blue
	case "Bug":
		col = Color.Red
	case "Epic", "Story":
		col = Color.Magenta
	case "Setup":
		col = Color.Cyan
	}

	if short {
		return fmt.Sprintf("%s%-12s%s", col, issueType, Color.Nocolor)
	}

	return fmt.Sprintf("%s%-45s%s", col, issueType, Color.Nocolor)
}

func Status(status string, short bool) string {
	var col string

	switch status {
	case "Closed", "Resolved", "Verified":
		col = Color.Green
	case "Programmed", "Peer Review", "Ready for Test", "Ready for review":
		col = Color.Cyan
	case "To Be Fixed", "In Progress", "Accepted", "Awaiting info":
		col = Color.Blue
	case "New", "Open":
		col = Color.Bold
	case "Rejected":
		col = Color.Red
	}

	if short {
		return fmt.Sprintf("%s%-10s%s", col, status, Color.Nocolor)
	}

	return fmt.Sprintf("%s%-20s%s", col, status, Color.Nocolor)
}

func Priority(priority string, short bool) string {
	var col string

	switch priority {
	case "Low":
		col = Color.Green
	case "Normal":
		col = Color.Blue
	case "Critical", "High", "Blocker":
		col = Color.Red
	}

	if short {
		return fmt.Sprintf("%s%-10s%s", col, priority, Color.Nocolor)
	}

	return fmt.Sprintf("%s%-45s%s", col, priority, Color.Nocolor)
}

func TimeEstimate(estimate string) string {
	if estimate == "" {
		return "Not Specified"
	}

	return estimate
}

func FixVersions(issue types.IssueDescription) string {
	fixVersions := ""
	for _, v := range issue.Fields.FixVersions {
		fixVersions += ", " + v.Name
	}

	return strings.Replace(fixVersions, ", ", "", 1)
}

func SprintHeader(sprint types.Sprint) string {
	var statusColor string

	switch sprint.State {
	case "ACTIVE":
		statusColor = Color.Green
	case "CLOSED":
		statusColor = Color.Red
	default:
		statusColor = Color.Blue
	}

	status := fmt.Sprintf("%s(%s%s%s)%s",
		Color.Cyan, statusColor, sprint.State, Color.Cyan, Color.Nocolor)

	return fmt.Sprintf("%s%s%70s %10s", Color.Bold, Color.Yellow, sprint.Name, status)
}

func SprintStatus(done bool) string {
	if done {
		return fmt.Sprintf("%sYes%s", Color.Green, Color.Nocolor)
	}

	return fmt.Sprintf("%sNo%s", Color.Blue, Color.Nocolor)
}

func StatsTotal(tot, goal, hoursPrDay float64, holidays int) string {
	switch {
	case tot >= goal-(hoursPrDay*float64(holidays)):
		return fmt.Sprintf("%s%.2f%s", Color.Green, tot, Color.Nocolor)
	case tot == 0:
		return fmt.Sprintf("%s%.2f%s", Color.Blue, tot, Color.Nocolor)
	default:
		return fmt.Sprintf("%s%.2f%s", Color.Red, tot, Color.Nocolor)
	}
}

func StatsAverage(avg, goal float64) string {
	switch {
	case avg >= goal:
		return fmt.Sprintf("%s%.2f%s", Color.Green, avg, Color.Nocolor)
	case avg == 0:
		return fmt.Sprintf("%s%.2f%s", Color.Blue, avg, Color.Nocolor)
	default:
		return fmt.Sprintf("%s%.2f%s", Color.Red, avg, Color.Nocolor)
	}
}

func StatsWorkdays(days, goal, holidays int) string {
	switch {
	case days >= goal-holidays:
		return fmt.Sprintf("%s%d%s", Color.Green, days, Color.Nocolor)
	case days == 0:
		return fmt.Sprintf("%s%d%s", Color.Blue, days, Color.Nocolor)
	default:
		return fmt.Sprintf("%s%d%s", Color.Red, days, Color.Nocolor)
	}
}

func StatsHolidays(holidays int) string {
	if holidays >= 1 {
		return fmt.Sprintf("%s%d%s", Color.Red, holidays, Color.Nocolor)
	}

	return fmt.Sprintf("%s%d%s", Color.Blue, holidays, Color.Nocolor)
}

func StatsSummary(num float64) string {
	if num >= 0 {
		return fmt.Sprintf("%s%.2f%s", Color.Green, num, Color.Nocolor)
	}

	return fmt.Sprintf("%s%.2f%s", Color.Red, num*-1, Color.Nocolor)
}
