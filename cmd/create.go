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
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mhersson/gojira/pkg/jira"
	"github.com/mhersson/gojira/pkg/types"
	"github.com/mhersson/gojira/pkg/util"
	"github.com/mhersson/gojira/pkg/util/format"
	"github.com/mhersson/gojira/pkg/util/validate"
)

const createUsage = `Create new issue
This guides the user through as series of questions which
can be aborted at anytime.

The description input supports multiple  lines of text,
and will open in $EDITOR, or vim by default. Writing JIRA notation,
with {noformat} and {code}, is supported, but for easier writing
three backticks will be converted to {noformat}.

After all data is collected they must be verified and confirmed
by the user, and only then will the request be sent to JIRA.

Usage:
  gojira create [PROJECT_KEY] [flags]

Flags:
  -h, --help   help for create
`

// createCmd represents the create command.
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new issue",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.ArbitraryArgs),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToUpper(args[0])
		validProjects := jira.GetValidProjects()
		project := validate.ProjectKey(key, validProjects)
		if project.ID == "" {
			fmt.Printf("%s is not a valid project key\n", key)
			os.Exit(1)
		}
		fmt.Printf("Creating new %s issue\n", project.Key)
		summary, rawSummary := getUserInputSummary()
		issueTypeID, issueTypeName := getUserInputIssueType(project)
		priorityID, priorityName := getUserInputPriority()
		desc, rawDesc := getUserInputDescription()

		getUserInputConfirmOk(project, issueTypeName, priorityName, rawSummary, rawDesc)

		newKey, err := jira.CreateNewIssue(project, issueTypeID, priorityID, summary, desc)
		if err != nil {
			fmt.Printf("Failed to create issue - %s\n", err.Error())
			fmt.Println(newKey)
			os.Exit(1)
		}

		fmt.Printf("%sNew issue has got key %s%s\n", format.Color.Blue, newKey, format.Color.Nocolor)

		ans := util.GetUserInput("Do you want to set the new issue active [y/N]: ", "[y|n]")
		if ans == "y" {
			setActiveIssue(newKey)
		}

		fmt.Printf("\n%sSuccessfully created new issue - run describe to see the details%s\n\n",
			format.Color.Green, format.Color.Nocolor)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.SetUsageTemplate(createUsage)
}

func getUserInputPriority() (string, string) {
	priorities := jira.GetPriorities()

	fmt.Println("Choose issue priority:")

	for i, v := range priorities {
		fmt.Printf("%d. %s\n", i, v.Name)
	}

	r := fmt.Sprintf("^([0-%d])$", len(priorities)-1)
	index := util.GetUserInput("", r)

	x, _ := strconv.Atoi(index)

	for i, v := range priorities {
		if i == x {
			return v.ID, v.Name
		}
	}

	return "", ""
}

func getUserInputIssueType(project types.Project) (string, string) {
	issueTypes := jira.GetProjectIssueTypes(project.Key)

	fmt.Println("Choose issue type:")

	for i, v := range issueTypes {
		fmt.Printf("%d. %s\n", i, v.Name)
	}

	// This is bad code - only supports range from 0-19
	r := fmt.Sprintf("^([0-%d])$", len(issueTypes)-1)
	if len(issueTypes) > 10 {
		r = fmt.Sprintf("^([0-9][0-%d]?)$", len(issueTypes)-11)
	}

	index := util.GetUserInput("", r)

	x, _ := strconv.Atoi(index)

	for i, v := range issueTypes {
		if i == x {
			return v.ID, v.Name
		}
	}

	return "", ""
}

func getUserInputSummary() (string, string) {
	fmt.Print("Enter summary: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadBytes('\n')

	if input[0] == '\n' {
		os.Exit(0)
	}

	st := strings.TrimSpace(string(input))

	summary, err := json.Marshal(st)
	if err != nil {
		fmt.Println("Failed to parse comment")
		os.Exit(1)
	}

	// Remove the {} around the comment
	escaped := string(summary[1 : len(summary)-1])

	return escaped, st
}

func getUserInputDescription() (string, string) {
	desc, err := captureInputFromEditor("", "description*")
	if err != nil {
		fmt.Println("Failed to read user input")
		os.Exit(1)
	}

	escaped := util.MakeStringJSONSafe(string(desc))

	return escaped, string(desc)
}

func getUserInputConfirmOk(project types.Project, issueType, pri, summary, description string) bool {
	fmt.Printf("%sPlease check you input:%s\n", format.Color.Blue, format.Color.Nocolor)
	fmt.Printf("Project %s, Type: %s, Priority: %s\n", project.Key, issueType, pri)
	fmt.Printf("Summary: %s\n", summary)
	fmt.Printf("Description:\n%s\n", description)

	ans := util.GetUserInput("Is this correct [y/N]: ", "[y|n]")
	if ans == "y" {
		return true
	}

	fmt.Println("Cancelled by user")

	return false
}
