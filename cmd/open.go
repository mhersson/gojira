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
package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"gitlab.com/mhersson/gojira/pkg/types"
)

const openUsage string = `This command will open the issue in your default browser

Usage:
  gojira open [ISSUE KEY] [flags]

Aliases:
  open, o

Flags:
  -h, --help                   help for open
`

// openCmd represents the open command.
var openCmd = &cobra.Command{
	Use:     "open",
	Short:   "Open in browser",
	Aliases: []string{"o"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			IssueKey = strings.ToUpper(args[0])
		}
		openbrowser(Cfg.JiraURL + "/browse/" + IssueKey)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
	openCmd.SetUsageTemplate(openUsage)
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = &types.Error{Message: "unsupported platform"}
	}

	if err != nil {
		fmt.Println(err)
	}
}
