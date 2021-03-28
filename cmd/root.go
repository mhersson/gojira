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
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var color = Color{"\033[31m", "\033[32m", "\033[33m", "\033[34m", "\033[35m", "\033[36m",
	"\033[1m", "\033[4m", "\033[0m"}
var issueKey string
var workDate string    // Used by `add work` to specify date
var workComment string // Used by `add work` to add a custom comment to the log
var jqlFilter string   // Used by `get all` to create customer queries
var assignee string    // Used by `update assignee`
var cacheFolder = path.Join(getHomeFolder(), ".gojira")
var issueFile = path.Join(cacheFolder, "issue")
var issueTypeFile = path.Join(cacheFolder, "issuetype")
var boardFile = path.Join(cacheFolder, "board")

// Color type.
type Color struct {
	red     string
	green   string
	yellow  string
	blue    string
	magenta string
	cyan    string
	bold    string
	ul      string
	nocolor string
}

// var cfgFile string.
type Config struct {
	JiraURL      string `yaml:"JiraURL"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	PasswordType string `yaml:"passwordtype"`
}

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

var config Config

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use: "gojira",
	Long: `The Gojira JIRA client

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

Example:
  To show the current active issue (get active) using aliases
  # gojira g a`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	home := getHomeFolder()

	ex, err := os.Executable()
	if err != nil {
		fmt.Println(err.Error())
	}

	exedir := path.Dir(ex)

	// Search config in home directory with name "config" (without extension).
	viper.AddConfigPath(path.Join(home, ".config/gojira"))
	viper.AddConfigPath(exedir)
	viper.SetConfigName("config")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
		config.JiraURL = viper.GetString("JiraURL")
		config.Username = viper.GetString("username")
		config.Password = viper.GetString("password")
		config.PasswordType = viper.GetString("passwordtype")

		if config.JiraURL[len(config.JiraURL)-1:] == "/" {
			config.JiraURL = config.JiraURL[:len(config.JiraURL)-1]
		}

		err := getPassword(&config)
		if err != nil {
			fmt.Println("Failed to get password")
			os.Exit(1)
		}
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

func getPassword(config *Config) error {
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
