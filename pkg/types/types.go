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
package types

import (
	"time"
)

// var cfgFile string.
type Config struct {
	JiraURL             string  `yaml:"JiraURL"` //nolint:tagliatelle
	Username            string  `yaml:"username"`
	Password            string  `yaml:"password"`
	PasswordType        string  `yaml:"passwordtype"`
	UseTimesheetPlugin  bool    `yaml:"useTimesheetPlugin"`
	CheckForUpdates     bool    `yaml:"checkForUpdates"`
	NumWorkingDays      int     `yaml:"numberOfWorkingDays"`
	WorkingHoursPerDay  float64 `yaml:"numberOfWorkingHoursPerDay"`
	WorkingHoursPerWeek float64 `yaml:"numberOfWorkingHoursPerWeek"`
}

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// Color type.
type Color struct {
	Red     string
	Green   string
	Yellow  string
	Blue    string
	Magenta string
	Cyan    string
	Bold    string
	Ul      string
	Nocolor string
}

type IssueDescription struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		FixVersions []struct {
			Name string `json:"name"`
		} `json:"fixVersions"`
		Summary    string `json:"summary"`
		Epic       string `json:"customfield_10500"` //nolint:tagliatelle
		Resolution struct {
			Name string `json:"name"`
		} `json:"resolution"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
		Labels     []string `json:"labels"`
		IssueLinks []struct {
			Type struct {
				Name    string `json:"name"`
				Inward  string `json:"inward"`
				Outward string `json:"outward"`
			}
			OutwardIssue struct {
				Key    string `json:"key"`
				Fields struct {
					Summary string `json:"summary"`
					Status  struct {
						Name string `json:"name"`
					} `json:"status"`
					Priority struct {
						Name string `json:"name"`
					} `json:"priority"`
					IssueType struct {
						Name string `json:"name"`
					} `json:"issueType"`
				} `json:"fields"`
			} `json:"outwardIssue"`
			InwardIssue struct {
				Key    string `json:"key"`
				Fields struct {
					Summary string `json:"summary"`
					Status  struct {
						Name string `json:"name"`
					} `json:"status"`
					Priority struct {
						Name string `json:"name"`
					} `json:"priority"`
					IssueType struct {
						Name string `json:"name"`
					} `json:"issueType"`
				} `json:"fields"`
			} `json:"inwardIssue"`
		} `json:"issuelinks"`
		Assignee struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
		Reporter struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"reporter"`
		Worklog struct {
			Worklogs []Worklog `json:"worklogs"`
		} `json:"worklog"`
		IssueType struct {
			Name string `json:"name"`
		} `json:"issueType"`
		Project struct {
			Name string `json:"name"`
		} `json:"project"`
		ChangeVisibility struct {
			Value string `json:"value"`
		} `json:"customfield_10707"` //nolint:tagliatelle
		Created      string `json:"created"`
		Updated      string `json:"updated"`
		Description  string `json:"description"`
		TimeTracking struct {
			Estimate  string `json:"originalEstimate"`
			Remaining string `json:"remainingEstimate"`
			TimeSpent string `json:"timeSpent"`
		} `json:"timetracking"`
		Comment struct {
			Comments []Comment `json:"comments"`
		} `json:"comment"`
	} `json:"fields"`
}

type Issue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		Summary   string `json:"summary"`
		IssueType struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"issuetype"`
		Assignee struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
		Priority struct {
			Name string `json:"name"`
		} `json:"priority"`
		Updated string `json:"updated"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
	} `json:"fields"`
}

type Comment struct {
	ID     string `json:"id"`
	Author struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	} `json:"author"`
	Body       string `json:"body"`
	Created    string `json:"created"`
	Visibility struct {
		Value string `json:"value"`
	} `json:"visibility"`
}

type Worklog struct {
	Author struct {
		DisplayName string `json:"displayName"`
		Name        string `json:"name"`
	} `json:"author"`
	Comment          string `json:"comment"`
	Created          string `json:"created"`
	Started          string `json:"started"`
	TimeSpent        string `json:"timeSpent"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
}

type Transition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   struct {
		Description    string `json:"description"`
		Name           string `json:"name"`
		ID             string `json:"id"`
		StatusCategory struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"statusCategory"`
	} `json:"to"`
}

type IssueCreateMeta struct {
	Projects []Project `json:"projects"`
}

type Project struct {
	ID         string      `json:"id"`
	Key        string      `json:"key"`
	Name       string      `json:"name"`
	IssueTypes []IssueType `json:"issuetypes"`
}

type IssueType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Priority struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Struct for representing the time a user
// has spent on an issue on a given date.
type TimeSpentUserIssue struct {
	ID               string
	Key              string
	Date             string
	User             string
	Summary          string
	TimeSpent        string
	TimeSpentSeconds int
}

// Used by getmyworklog command
// when timesheet plugin is enabled.
type Timesheet struct {
	Key     string `json:"key"`
	Summary string `json:"summary"`
	Entries []struct {
		ID             int    `json:"id"`
		Author         string `json:"author"`
		AuthorFullName string `json:"authorFullName"`
		StartDate      int    `json:"startDate"`
		TimeSpent      int    `json:"timeSpent"`
		Comment        string `json:"comment"`
	}
}

// Used by worklog command to be able to
// sort and edit worklogs.
type SimplifiedTimesheet struct {
	ID        int
	Date      string
	StartDate string
	Key       string
	Summary   string
	Comment   string
	TimeSpent int
}

type RapidView struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	SprintSupportEnabled bool   `json:"sprintSupportEnabled"`
}

type Sprint struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	IssuesIDs []int  `json:"issuesIds"`
}

type SprintIssue struct {
	ID                int      `json:"id"`
	Key               string   `json:"key"`
	TypeID            string   `json:"typeId"`
	Summary           string   `json:"summary"`
	PriorityID        string   `json:"priorityId"`
	Assignee          string   `json:"assignee"`
	AssigneeName      string   `json:"assigneeName"`
	Epic              string   `json:"epic"`
	EstimateStatistic TimeStat `json:"estimateStatistic"`
	TrackingStatistic TimeStat `json:"trackingStatistic"`
	Hidden            bool     `json:"hidden"`
	Done              bool     `json:"done"`
}

type SprintContent struct {
	CompletedIssues                   []SprintIssue `json:"completedIssues"`
	IssuesNotCompletedInCurrentSprint []SprintIssue `json:"issuesNotCompletedInCurrentSprint"`
	IssuesCompletedInAnotherSprint    []SprintIssue `json:"issuesCompletedInAnotherSprint"`
}

type TimeStat struct {
	StatFieldID    string `json:"statFieldId"`
	StatFieldValue struct {
		Value float64 `json:"value"`
		Text  string  `json:"text"`
	} `json:"statFieldValue"`
}

type Week struct {
	StartDate time.Time
	EndDate   time.Time
	Worklogs  []SimplifiedTimesheet
}

func inSlice(slice []string, s string) bool {
	for _, a := range slice {
		if a == s {
			return true
		}
	}

	return false
}

func (w *Week) Number() int {
	_, i := w.StartDate.ISOWeek()

	return i
}

func (w *Week) WorkDays() int {
	days := []string{}
	for _, d := range w.Worklogs {
		if !inSlice(days, d.Date) {
			days = append(days, d.Date)
		}
	}

	return len(days)
}

func (w *Week) TotalTime() float64 {
	total := float64(0)
	for _, d := range w.Worklogs {
		total += float64(d.TimeSpent)
	}

	return total / 3600
}

func (w *Week) Average() float64 {
	if w.WorkDays() == 0 {
		return 0
	}

	return w.TotalTime() / float64(w.WorkDays())
}
