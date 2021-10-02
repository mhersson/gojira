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

package util

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gitlab.com/mhersson/gojira/pkg/types"
	"gitlab.com/mhersson/gojira/pkg/util/convert"
)

//go:embed tpl/*.tmpl
var tplFS embed.FS

func WeekStartEndDate(year, week int) (string, string) {
	t := time.Date(year, 7, 1, 0, 0, 0, 0, time.UTC)

	if wd := t.Weekday(); wd == time.Sunday {
		t = t.AddDate(0, 0, -6)
	} else {
		t = t.AddDate(0, 0, -int(wd)+1)
	}

	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (week-w)*7)
	e := t.AddDate(0, 0, 6)

	return t.Format("2006-01-02"), e.Format("2006-01-02")
}

func GetCurrentDate() string {
	now := time.Now().UTC()
	// jira date format - "2017-12-07"
	return now.Format("2006-01-02")
}

func DateIsToday(date string) bool {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		fmt.Printf("Failed to parse date: %v\n", err)
		os.Exit(1)
	}

	t := time.Now()

	if d.Year() == t.Year() && d.Month() == t.Month() && d.Day() == t.Day() {
		return true
	}

	return false
}

// Returns to today's date on format 2006-01-02.
func Today() string {
	t := time.Now()

	return t.Format("2006-01-02")
}

func GetActiveIssue(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Active issue is not set")
		os.Exit(1)
	}

	out, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Failed to get active issue")
		os.Exit(1)
	}

	return string(out)
}

func GetActiveBoard(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Active board is not set")
		os.Exit(0)
	}

	out, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Failed to get active board")
		os.Exit(1)
	}

	return string(out)
}

func GetWorklogsSorted(worklogs []types.Timesheet, truncate bool) []types.SimplifiedTimesheet {
	week := []types.SimplifiedTimesheet{}

	for _, wl := range worklogs {
		if len(wl.Summary) > 40 && truncate {
			wl.Summary = wl.Summary[:40] + ".."
		}

		for _, entry := range wl.Entries {
			if len(entry.Comment) > 31 && truncate {
				entry.Comment = entry.Comment[:31] + ".."
			}

			date := time.Unix(0, int64(entry.StartDate*int(time.Millisecond))).Format("2006-01-02")
			startdate := time.Unix(0, int64(entry.StartDate*int(time.Millisecond))).Format("2006-01-02 15:04")
			ts := types.SimplifiedTimesheet{
				ID:        entry.ID,
				Date:      date,
				StartDate: startdate,
				Key:       wl.Key,
				Summary:   wl.Summary,
				Comment:   entry.Comment,
				TimeSpent: entry.TimeSpent}
			week = append(week, ts)
		}
	}

	sort.Slice(week, func(i, j int) bool {
		return week[i].StartDate < week[j].StartDate
	})

	return week
}

func GroupWorklogsByWeek(
	fromDate, toDate string, worklogs []types.SimplifiedTimesheet, holidays []string) []types.Week {
	t1, _ := time.Parse("2006-01-02", fromDate)
	t2, _ := time.Parse("2006-01-02", toDate)

	weeks := []types.Week{}

	for t1.Before(t2) || t1.Equal(t2) {
		week := types.Week{}

		for _, w := range worklogs {
			d, _ := time.Parse("2006-01-02", w.Date)

			if (t1.Before(d) || t1.Equal(d)) && (t1.Equal(d) || t1.Add(7*24*time.Hour).After(d)) {
				week.Worklogs = append(week.Worklogs, w)
			} else if d.After(t1.Add(6 * 24 * time.Hour)) {
				break
			}
		}

		for _, h := range holidays {
			d, _ := time.Parse("2006-01-02", h)
			if (t1.Before(d) || t1.Equal(d)) && (t1.Equal(d) || t1.Add(5*24*time.Hour).After(d)) {
				week.PublicHolidays++
			}
		}

		week.StartDate = t1
		week.EndDate = t1.Add(6 * 24 * time.Hour)
		weeks = append(weeks, week)
		t1 = t1.Add(7 * 24 * time.Hour)
	}

	return weeks
}

func GetUserInput(prompt string, regRange string) string {
	if prompt == "" {
		fmt.Print("\nPlease enter value (press enter to quit): ")
	} else {
		fmt.Print(prompt)
	}

	reader := bufio.NewReader(os.Stdin)

	var answer string

	for {
		input, _ := reader.ReadBytes('\n')
		if input[0] == '\n' {
			fmt.Println("Cancelled by user")
			os.Exit(0)
		}

		re := regexp.MustCompile(regRange)
		m := re.Find(bytes.TrimSpace(input))

		if m == nil {
			fmt.Println("Invalid choice")
			fmt.Print("Please try again: ")

			continue
		}

		answer = string(m)

		break
	}

	return answer
}

func MakeStringJSONSafe(str string) string {
	strText := strings.ReplaceAll(str, "```", "{noformat}")
	// Convert the string into json to escape whatever
	// chars json needs to have escaped
	jsonStr, err := json.Marshal(strText)
	if err != nil {
		fmt.Println("Failed to parse comment")
		os.Exit(1)
	}

	// Remove the surrounding curly brackets
	escaped := string(jsonStr[1 : len(jsonStr)-1])

	return escaped
}

func ExecuteTemplate(filename string, content interface{}) []byte {
	temp, err := tplFS.ReadFile(filepath.Join("tpl", filename))
	if err != nil {
		panic(err)
	}

	t := template.Must(template.New(filepath.Base(filename)).Funcs(templateFuncMap()).Parse(string(temp)))

	var buffer bytes.Buffer

	err = t.Execute(&buffer, content)
	if err != nil {
		panic(err)
	}

	return buffer.Bytes()
}

func templateFuncMap() template.FuncMap {
	var fns = template.FuncMap{
		"getTime": func(date string) string {
			return strings.Split(date, " ")[1]
		},
		"convertTimeSpent": convert.SecondsToHoursAndMinutes,
		"new": func(id int) string {
			if id == 666 {
				return "new"
			}

			return strconv.Itoa(id)
		},
	}

	return fns
}

func LoadPublicHolidays(filename, year, countryCode string) []types.PublicHoliday {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		data := HTTPGet("https://date.nager.at/api/v3/publicholidays/" + year + "/" + strings.ToUpper(countryCode))

		err := os.WriteFile(filename, data, 0600)
		if err != nil {
			fmt.Printf("Failed to write public holidays to cache - %v\n", err)
		}
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Failed to load public holidays - %v\n", err)
	}

	publicHolidays := []types.PublicHoliday{}

	err = json.Unmarshal(data, &publicHolidays)
	if err != nil {
		fmt.Printf("Failed to parse public holidays - %v\n", err)
	}

	return publicHolidays
}

func GetPublicHolidayDates(publicHolidays []types.PublicHoliday) []string {
	dates := []string{}
	for _, h := range publicHolidays {
		dates = append(dates, h.Date)
	}

	return dates
}

func HTTPGet(url string) []byte {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)

		return []byte{}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	return body
}
