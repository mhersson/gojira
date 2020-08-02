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

type IssueDescriptionResponse struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		FixVersions []struct {
			Name string `json:"name"`
		} `json:"fixVersions"`
		Summary    string `json:"summary"`
		Epic       string `json:"customfield_10500"`
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
		Worklog   WorklogsResponse `json:"worlog"`
		IssueType struct {
			Name string `json:"name"`
		} `json:"issueType"`
		Project struct {
			Name string `json:"name"`
		} `json:"project"`
		ChangeVisibility struct {
			Value string `json:"value"`
		} `json:"customfield_10707"`
		Created      string `json:"created"`
		Updated      string `json:"updated"`
		Description  string `json:"description"`
		TimeTracking struct {
			Estimate  string `json:"originalEstimate"`
			Remaining string `json:"remainingEstimate"`
			TimeSpent string `json:"timeSpent"`
		} `json:"timetracking"`
		Comment CommentsResponse `json:"comment"`
	} `json:"fields"`
}

type IssuesResponse struct {
	Issues []IssueResponse `json:"issues"`
}

type IssueResponse struct {
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

type CommentsResponse struct {
	Comments []CommentResponse `json:"comments"`
}

type CommentResponse struct {
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

type WorklogsResponse struct {
	Worklogs []WorklogResponse `json:"worklogs"`
}

type WorklogResponse struct {
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

type TransitionsResponse struct {
	Transitions []TransitionResponse `json:"transitions"`
}

type TransitionResponse struct {
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

type Priorities []struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Struct for representing the time a user
// has spent on an issue on a given date.
type TimeSpentUserIssue struct {
	Key              string
	Date             string
	User             string
	Summary          string
	TimeSpent        string
	TimeSpentSeconds int
}
