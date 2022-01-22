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

	"github.com/spf13/cobra"
)

const completionUsage string = `
This command prints shell code which must be evaluated
to provide interactive completion of gojira commands.

Usage:
  gojira completion SHELL [flags]

Examples:
  # Generate the gojira completion code for bash
  gojira completion bash > bash_completion.sh
  source bash_completion.sh
  
  # The above example depends on the bash-completion framework.
  # It must be sourced before sourcing the gojira completion,
  # i.e. on the Mac:
  
  brew install bash-completion
  source $(brew --prefix)/etc/bash_completion
  gojira completion bash > bash_completion.sh
  source bash_completion.sh
  
  # In zsh*, the following will load gojira cli zsh completion:
  source <(gojira completion zsh)
  
  * zsh completions are only supported in versions of zsh >= 5.2

`

// completionCmd represents the completion command.
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Output shell completion code for the specified shell",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "zsh":
			_ = rootCmd.GenZshCompletion(os.Stdout)
			fmt.Println("compdef _gojira gojira")
		case "bash":
			_ = rootCmd.GenBashCompletion(os.Stdout)
		default:
			fmt.Println("Completions are only supported for bash and zsh")
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
	completionCmd.SetUsageTemplate(completionUsage)
}
