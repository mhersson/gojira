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
package convert

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/mhersson/gojira/pkg/types"
)

func DurationStringToSeconds(duration string) (string, error) {
	// Format 0.5h OR 30m alone or 1h 30m combined
	re := regexp.MustCompile(`((([0-9.]{1,})(h))?\s?(([0-9]?[0-9])(m))?)`)
	m := re.FindStringSubmatch(duration)

	var seconds float64

	if m[0] != "" {
		if m[3] != "" {
			num, _ := strconv.ParseFloat(m[3], 64)
			seconds += num * 3600
		}

		if m[6] != "" {
			num, _ := strconv.ParseFloat(m[6], 64)
			seconds += num * 60
		}

		return strconv.FormatFloat(seconds, 'f', 0, 64), nil
	}

	return "", &types.Error{Message: "invalid duration format"}
}

func SecondsToHoursAndMinutes(seconds int, dropMinutes bool) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	if dropMinutes {
		return fmt.Sprintf("%dh", hours)
	}

	if minutes < 10 {
		return fmt.Sprintf("%dh 0%dm", hours, minutes)
	}

	return fmt.Sprintf("%dh %dm", hours, minutes)
}
