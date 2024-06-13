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
package convert_test

import (
	"testing"

	"github.com/mhersson/gojira/pkg/types"
	"github.com/mhersson/gojira/pkg/util/convert"
	"github.com/stretchr/testify/assert"
)

func TestDurationStringToSeconds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
		err      error
	}{
		{"1h 5m", "3900", nil},
		{"1h 0m", "3600", nil},
		{"0.5h", "1800", nil},
		{"05m", "300", nil},
		{"5m", "300", nil},
		{"85m", "5100", nil},
		{"5x", "", &types.Error{}},
		{"Wrong", "", &types.Error{}},
	}

	for _, v := range tests {
		ans, err := convert.DurationStringToSeconds(v.input)
		if v.err != nil {
			assert.Error(t, err)
		}

		if ans != v.expected {
			t.Errorf("Input: %s, got: %s, want: %s", v.input, ans, v.expected)
		}
	}
}

func TestSecondsToHoursAndMinutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		secs        int
		dropMinutes bool
		expected    string
	}{
		{3900, false, "1h 05m"},
		{3600, false, "1h 00m"},
		{1800, false, "0h 30m"},
		{300, false, "0h 05m"},
		{3900, true, "1h"},
		{3600, true, "1h"},
		{1800, true, "0h"},
		{300, true, "0h"},
	}

	for _, v := range tests {
		ans := convert.SecondsToHoursAndMinutes(v.secs, v.dropMinutes)
		if ans != v.expected {
			t.Errorf("Input: %d, got: %s, want: %s", v.secs, ans, v.expected)
		}
	}
}
