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
package validate

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gitlab.com/mhersson/gojira/pkg/jira"
	"gitlab.com/mhersson/gojira/pkg/types"
	"gitlab.com/mhersson/gojira/pkg/util"
)

func Date(date string) bool {
	re := regexp.MustCompile("^202[0-9]-((0[1-9])|(1[0-2]))-((0[1-9])|([1-2][0-9])|(3[0-1]))$")

	return re.MatchString(date)
}

func Time(time string) bool {
	re := regexp.MustCompile("^([0-1][0-9]|2[0-3]):[0-5][0-9]$")

	return re.MatchString(time)
}

func IssueKey(key *string, issueFile string) {
	if *key != "" {
		re := regexp.MustCompile("[A-Z]{2,9}-[0-9]{1,4}")

		m := re.MatchString(*key)
		if !m {
			fmt.Println("Invalid key")
			os.Exit(1)
		}

		if !jira.IssueExists(key) {
			fmt.Printf("%s does not exist\n", *key)
			os.Exit(1)
		}
	} else {
		*key = util.GetActiveIssue(issueFile)
	}
}

func ProjectKey(key string, projects types.IssueCreateMeta) types.Project {
	// This validates the project key not the issue key
	// hvis is the project key + a number
	for _, v := range projects.Projects {
		if key == strings.ToUpper(v.Key) {
			return v
		}
	}

	return types.Project{}
}

func CommentID(commentID string) bool {
	// This maybe wrong, but so far I have not
	// seen an id which is not 6 digits long
	re := regexp.MustCompile("^[0-9]{6}$")

	return re.MatchString(commentID)
}
