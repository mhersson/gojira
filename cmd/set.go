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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set issue or board active",
	Long: `Set marks and issue or board as active

Active issues and boards will be used as default argument if no other arguments
are given.`,
}

var setActiveCmd = &cobra.Command{
	Use:   "active",
	Short: "Set issue active",
	Long: `Set issue active.

This does not change the status of the issue in JIRA, it just tells Gojira that
this is your active issue, the one you are currently working on. Setting an
issue as active removes the need of specifying an issueKey with (almost) every
command`,
	Aliases: []string{"a"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		issueKey = strings.ToUpper(args[0])
		setActiveIssue(issueKey)
		key := getActiveIssue()
		fmt.Printf("Issue %s is active\n", key)
	},
}

var setActiveBoardCmd = &cobra.Command{
	Use:   "board",
	Short: "Set board as active",
	Long: `Set the board active.

This marks the given board as active board or your board of interest, and will
be used by the get sprint command if no other board name is specified.
`,
	Aliases: []string{"b"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		board := strings.ToLower(args[0])
		setActiveBoard(board)
	},
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.AddCommand(setActiveCmd)
	setCmd.AddCommand(setActiveBoardCmd)
}

func setActiveIssue(key string) {
	issues := getIssues("key = " + key)
	if len(issues) != 1 {
		fmt.Printf("Issue %s does not exist, and can not be set active\n", key)
		os.Exit(1)
	}

	_, err := os.Stat(issueFile)
	if os.IsNotExist(err) {
		err := os.Mkdir(path.Join(getHomeFolder(), ".gojira"), 0755)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	err = ioutil.WriteFile(issueFile, []byte(key), 0600)
	if err != nil {
		fmt.Printf("Failed to set %s active\n", key)
		os.Exit(1)
	}

	err = ioutil.WriteFile(issueTypeFile,
		[]byte(issues[0].Fields.IssueType.ID), 0600)
	if err != nil {
		fmt.Printf("Failed to set %s active\n", key)
		os.Exit(1)
	}
}

func setActiveBoard(board string) {
	if id := getRapidViewID(board); id == nil {
		fmt.Printf("Board %s does not exist, and can not be set active\n", board)
		os.Exit(1)
	}

	_, err := os.Stat(boardFile)
	if os.IsNotExist(err) {
		if err := os.Mkdir(path.Join(getHomeFolder(), ".gojira"), 0755); !os.IsExist(err) {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	err = ioutil.WriteFile(boardFile, []byte(board), 0600)
	if err != nil {
		fmt.Printf("Failed to set %s active\n", board)
		os.Exit(1)
	}
}
