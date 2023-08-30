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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/mhersson/gojira/pkg/jira"
	"gitlab.com/mhersson/gojira/pkg/util"
)

const setActiveUsage string = `
Setting an issue as active does not change the status of the issue in JIRA, it
just tells Gojira that this is your active issue, the one you are currently
working on. Setting an issue as active removes the need of specifying an
issueKey with (almost) every command

The same goes for setting a board as active. It marks the given board as your
board of interest, and will be used by the get sprint command when no other board
name is specified
Usage:
  gojira set active [issue|board] [ISSUE KEY|BOARD NAME] [flags]

Aliases:
  active, a

Flags:
  -h, --help                   help for comment
`

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set issue or board active",
}

var setActiveCmd = &cobra.Command{
	Use:     "active",
	Short:   "Set issue or board active",
	Aliases: []string{"a"},
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "issue":
			IssueKey = strings.ToUpper(args[1])
			setActiveIssue(IssueKey)
			key := util.GetActiveIssue(IssueFile)
			fmt.Printf("Issue %s is active\n", key)
		case "board":
			board := strings.ToLower(args[1])
			setActiveBoard(board)
			fmt.Printf("Board '%s' is active\n", board)
		default:
			fmt.Println("First argument must be issue or board")
		}
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.AddCommand(setActiveCmd)

	setActiveCmd.SetUsageTemplate(setActiveUsage)
}

func setActiveIssue(key string) {
	issues := jira.GetIssues("key = " + key)
	if len(issues) != 1 {
		fmt.Printf("Issue %s does not exist, and can not be set active\n", key)
		os.Exit(1)
	}

	createConfigFolder()

	err := os.WriteFile(IssueFile, []byte(key), 0o600)
	if err != nil {
		fmt.Printf("Failed to set %s active\n", key)
		os.Exit(1)
	}

	err = os.WriteFile(IssueTypeFile,
		[]byte(issues[0].Fields.IssueType.ID), 0o600)
	if err != nil {
		fmt.Printf("Failed to set %s active\n", key)
		os.Exit(1)
	}
}

func setActiveBoard(board string) {
	if id := jira.GetRapidViewID(board); id == nil {
		fmt.Printf("Board %s does not exist, and can not be set active\n", board)
		os.Exit(1)
	}

	createConfigFolder()

	err := os.WriteFile(BoardFile, []byte(board), 0o600)
	if err != nil {
		fmt.Printf("Failed to set %s active\n", board)
		os.Exit(1)
	}
}

func createConfigFolder() {
	_, err := os.Stat(ConfigFolder)

	if os.IsNotExist(err) {
		err := os.Mkdir(ConfigFolder, 0o755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
