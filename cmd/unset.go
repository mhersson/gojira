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
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// TODO: Split into subcommands for issue, sprint and kanban
var unsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Unset (clear) active issue and board",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "issue":
			unsetActive(IssueFile)
			fmt.Println("Active issue cleared")
		case "board":
			unsetActive(BoardFile)
			fmt.Println("Active board cleared")
		default:
			fmt.Println("First argument must be issue or board")
		}
	},
}

func init() {
	rootCmd.AddCommand(unsetCmd)
}

func unsetActive(file string) {
	err := os.Remove(file)
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("Failed to clear active issue")
		os.Exit(1)
	}
}
