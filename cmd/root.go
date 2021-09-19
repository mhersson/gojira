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
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"gitlab.com/mhersson/gojira/pkg/types"
)

var rootCmdLong = `The Gojira JIRA client

This project is a product of me being bored out of my mind
because of Corona virus quarantine combined with Easter holidays.

Gojira is the Japanese word for Godzilla.

Features:
  - Create new issues
  - Add and update comments
  - Use your favorite editor set by $EDITOR, defaults to vim
  - Change status
  - Assign to yourself or to others
  - Report time spent
  - Show comments, current status and the entire worklog
  - One view to show it all with the describe command
  - Display all unresolved issues assigned to you
  - Display the current sprint with all issues and statuses
  - Mark issue and/or board as active for less typing
  - Open issue in default browser

Gojira integrates with passwordstore and gpg to keep your password safe.

All commands have a short help text you can access by passing -h or --help. Most
commands, but not all, have assigned aliases to their first letter for less
typing.
`

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:  "gojira",
	Long: rootCmdLong,
	Run: func(cmd *cobra.Command, args []string) {
		if VersionFlag {
			fmt.Printf("Gojira version: %s,  git rev: %s\n", GojiraVersion, GojiraGitRevision)

			os.Exit(0)
		}
		fmt.Println(rootCmdLong)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().BoolVar(&VersionFlag, "version", false, "Print version information")
}

func initConfig() {
	home := getHomeFolder()

	ex, err := os.Executable()
	if err != nil {
		fmt.Println(err.Error())
	}

	exedir := path.Dir(ex)

	viper.AddConfigPath(path.Join(home, ".config/gojira"))
	viper.AddConfigPath(exedir)
	viper.SetConfigName("config")

	if err := viper.ReadInConfig(); err == nil {
		Cfg.JiraURL = viper.GetString("JiraURL")
		Cfg.Username = viper.GetString("username")
		Cfg.Password = viper.GetString("password")
		Cfg.PasswordType = viper.GetString("passwordtype")
		Cfg.UseTimesheetPlugin = viper.GetBool("useTimesheetPlugin")
		Cfg.CheckForUpdates = viper.GetBool("checkForUpdates")

		if Cfg.JiraURL[len(Cfg.JiraURL)-1:] == "/" {
			Cfg.JiraURL = Cfg.JiraURL[:len(Cfg.JiraURL)-1]
		}

		err := getPassword(&Cfg)
		if err != nil {
			fmt.Println("Failed to get password")
			os.Exit(1)
		}
	}

	if GojiraGitRevision != "" && Cfg.CheckForUpdates {
		revs := runGit([]string{"ls-remote", GojiraRepository})
		getLatestRevision(revs)
	}
}

func getHomeFolder() string {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return home
}

func getPassword(config *types.Config) error {
	switch config.PasswordType {
	case "pass":
		pw, err := runPass([]string{config.Password})
		if err != nil {
			return err
		}

		config.Password = pw
	case "gpg":
		pw, err := decodeGPG(config.Password)
		if err != nil {
			return err
		}

		config.Password = pw
	default:
		fmt.Println("You should encrypt your password!!")
		fmt.Println("Start using your gpg key by running the following command")
		fmt.Println("echo \"yourpassword\" | gpg -r yourgpgkey -e --armor | base64 --wrap 0")
		fmt.Println("Copy the output and paste it into the config.yaml password field, all on one line")
		fmt.Println("Then set passwordtype = gpg in your config file")
	}

	return nil
}

func runPass(args []string) (string, error) {
	output, err := exec.Command("pass", args...).Output()
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func decodeGPG(b64Armored string) (string, error) {
	cmd := exec.Command("gpg", "--decrypt")
	armored, _ := base64.StdEncoding.DecodeString(b64Armored)
	cmd.Stdin = bytes.NewReader(armored)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func getLatestRevision(revs string) {
	re := regexp.MustCompile(`([a-z0-9]{40})\s{1,}refs/heads/main`)
	m := re.FindStringSubmatch(revs)

	if len(m) == 2 {
		if !strings.HasPrefix(m[1], GojiraGitRevision) {
			fmt.Println("A new version of Gojira is available")
		}
	}
}

func runGit(args []string) string {
	out, err := exec.Command("git", args...).CombinedOutput()
	cobra.CheckErr(err)

	return strings.TrimSpace(string(out))
}
