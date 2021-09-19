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
	"path"

	"gitlab.com/mhersson/gojira/pkg/types"
)

// GojiraVersion GojiraGitRevision and GojiraRepository
// are all inserted at build time from the Makefile.
var GojiraVersion string
var GojiraGitRevision string
var GojiraRepository string

var IssueKey string
var WorkDate string    // Used by `add work` to specify date
var WorkTime string    // Used by `add work` to specify at what time the work was done
var WorkComment string // Used by `add work` to add a custom comment to the log
var JQLFilter string   // Used by `get all` to create customer queries
var Assignee string    // Used by `update assignee`
var VersionFlag bool
var ShowEntireWeek = false // Used by `get myworklog`
var MergeToday = false     // Used by `edit myworklog`
var CacheFolder = path.Join(getHomeFolder(), ".gojira")
var IssueFile = path.Join(CacheFolder, "issue")
var IssueTypeFile = path.Join(CacheFolder, "issuetype")
var BoardFile = path.Join(CacheFolder, "board")

var Cfg types.Config
