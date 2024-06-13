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
	"regexp"
	"strings"

	"github.com/mhersson/gojira/pkg/jira"
	"github.com/mhersson/gojira/pkg/util"
	"github.com/spf13/cobra"
)

const setActiveUsage string = `
Setting an issue as active does not change the status of the issue in JIRA, it
just tells Gojira that this is your active issue, the one you are currently
working on. Setting an issue as active removes the need of specifying an
issueKey with (almost) every command

The same goes for setting a board as active. It marks the given board as your
board of interest, and will be used by the get sprint or kanban commands when
no other board name is specified

Usage:
  gojira set active [issue|sprint|kanban] [ISSUE KEY|BOARD NAME] [flags]

Aliases:
  active, a

Available Commands:
  issue       Set the active issue
  kanban      Set the active kanban board
  sprint      Set the active sprint

Flags:
  -h, --help                   help for comment
`

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set issue or board active",
}

var setActiveCmd = &cobra.Command{
	Use:     "active",
	Short:   "Set issue, sprint or kanban board active",
	Aliases: []string{"a"},
}

var setActiveIssueCmd = &cobra.Command{
	Use:     "issue",
	Short:   "Set the active issue",
	Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Aliases: []string{"i"},
	Run: func(cmd *cobra.Command, args []string) {
		IssueKey = strings.ToUpper(args[0])
		setActiveIssue(IssueKey)
		key := util.GetActiveIssue(IssueFile)
		fmt.Printf("Issue %s is active\n", key)
	},
}

var setActiveSprintCmd = &cobra.Command{
	Use:     "sprint",
	Short:   "Set the active sprint",
	Aliases: []string{"s"},
	Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		board := strings.ToLower(args[0])
		setActiveBoard(board, "sprint")
		fmt.Printf("Sprint '%s' is active\n", board)
	},
}

var setActiveKanbanCmd = &cobra.Command{
	Use:     "kanban",
	Short:   "Set the active kanban board",
	Aliases: []string{"k"},
	Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		board := strings.ToLower(args[0])
		setActiveBoard(board, "kanban")
		fmt.Printf("Kanban board '%s' is active\n", board)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.AddCommand(setActiveCmd)

	setActiveCmd.SetUsageTemplate(setActiveUsage)
	setActiveCmd.AddCommand(setActiveIssueCmd)
	setActiveCmd.AddCommand(setActiveSprintCmd)
	setActiveCmd.AddCommand(setActiveKanbanCmd)
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

func setActiveBoard(board, boardType string) {
	if id := jira.GetRapidViewID(board); id == nil {
		fmt.Printf("Board %s does not exist, and can not be set active\n", board)
		os.Exit(1)
	}

	var content []byte

	if _, err := os.Stat(BoardFile); err == nil {
		content, err = os.ReadFile(BoardFile)
		if err != nil {
			fmt.Println("Failed to read existing board config")
			os.Exit(1)
		}

		p := regexp.MustCompile(boardType + `=(.*)`)
		repl := p.ReplaceAllString(string(content), boardType+"="+board)
		if repl == string(content) {
			content = append(content, []byte(boardType+"="+board)...)
		} else {
			content = []byte(repl)
		}

	} else {
		createConfigFolder()
		content = []byte(boardType + "=" + board + "\n")
	}

	err := os.WriteFile(BoardFile, content, 0o600)
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
