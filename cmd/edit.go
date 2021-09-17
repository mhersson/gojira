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
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

const editDescriptionUsage string = `By default the active issue is edited,
but this can be changed by adding the issue key as argument.

Usage:
  gojira edit description [ISSUE KEY] [flags]

Aliases:
  description, d

Flags:
  -h, --help                   help for description
`

const editCommentUsage string = `By default the active issue is edited, but this can be
changed by adding the issue key as argument. When this is
the case the argument order is important, and the issue key
must always be the first argument.

The comment id can be found by running either "get comments" or "describe".
If not set the comment id of the most recent comment will be used.

Usage:
  gojira edit comment [ISSUE KEY] <COMMENT ID> [flags]

Aliases:
  comment, c

Flags:
  -h, --help                   help for comment
`

var editCmd = &cobra.Command{
	Use:     "edit",
	Short:   "Edit",
	Long:    "Edit comments, descriptions, worklog",
	Args:    cobra.NoArgs,
	Aliases: []string{"u"},
}

var editDescrptionCmd = &cobra.Command{
	Use:     "description",
	Short:   "Edit the description",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"d"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			issueKey = strings.ToUpper(args[0])
		}
		validateIssueKey(&issueKey)
		issue := getIssue(issueKey)

		desc, err := captureInputFromEditor(issue.Fields.Description, "description*")
		if err != nil {
			fmt.Println("Failed to read description")
			os.Exit(1)
		}

		err = updateDescription(issueKey, desc)
		if err != nil {
			fmt.Printf("Failed to update description, %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully saved new description")

	},
}

var editCommentCmd = &cobra.Command{
	Use:     "comment",
	Short:   "Edit comment",
	Aliases: []string{"c"},
	Args:    cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var commentID string

		switch len(args) {
		case 1:
			// First argument can be both comment id for the active issue
			// or another issue key, where the target comment is the latest
			if validateCommentID(args[0]) {
				// Comment id is valid, the issuekey will be set to the active issue
				commentID = args[0]
				validateIssueKey(&issueKey)
			} else {
				// The argument is not a valid comment id, check if it
				// is a valid issue key
				issueKey = strings.ToUpper(args[0])
				validateIssueKey(&issueKey)
			}

		case 2:
			// If two arguments are provided first must be the issueKey,
			// and second must be the comment id
			issueKey = strings.ToUpper(args[0])
			validateIssueKey(&issueKey)

			commentID = args[1]
			if !validateCommentID(commentID) {
				fmt.Println("Invalid comment id")
				os.Exit(1)
			}

		default:
			// If no argument is provided edit the last comment of the current active issue
			validateIssueKey(&issueKey)
		}

		// Get the existing comment
		ec := getComment(issueKey, commentID)
		if commentID == "" {
			commentID = ec.ID
		}

		if ec.ID == "" {
			if commentID != "" {
				fmt.Println("Comment id does not exist")
			} else {

				fmt.Println("Issue does not have any comments. Try add comment instead")
			}
			os.Exit(1)
		}

		comment, err := captureInputFromEditor(ec.Body, "comment*")
		if err != nil {
			fmt.Println("Failed to read comment")
		}

		err = updateComment(issueKey, comment, commentID)
		if err != nil {
			fmt.Printf("Failed to update comment - %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Println("Successfully saved new comment")
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	editCmd.AddCommand(editDescrptionCmd)
	editCmd.AddCommand(editCommentCmd)

	editDescrptionCmd.SetUsageTemplate(editDescriptionUsage)
	editCommentCmd.SetUsageTemplate(editCommentUsage)
}

func updateDescription(key string, desc []byte) error {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key)

	jsonDesc := makeStringJSONSafe(string(desc))

	payload := []byte(`{"fields":{"description":"` + jsonDesc + `"}}`)

	resp, err := update("PUT", url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func updateComment(key string, comment []byte, id string) error {
	url := config.JiraURL + "/rest/api/2/issue/" + strings.ToUpper(key) + "/comment/" + id

	escaped := makeStringJSONSafe(string(comment))

	payload := []byte(`{
		"body": "` + escaped + `",
		"visibility": {
			"type": "group",
			"value": "Internal users"
		}
	}`)

	resp, err := update("PUT", url, payload)
	if err != nil {
		fmt.Printf("%s\n", resp)

		return err
	}

	return nil
}

func validateCommentID(commentID string) bool {
	// This maybe wrong, but so far I have not
	// seen an id which is not 6 digits long
	re := regexp.MustCompile("^[0-9]{6}$")

	return re.MatchString(commentID)
}
