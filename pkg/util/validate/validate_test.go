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
package validate_test

import (
	"testing"

	"gitlab.com/mhersson/gojira/pkg/util/validate"
)

func TestDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"2006-01-02", false}, // Correct format, but way too old
		{"2020-01-02", true},
		{"2020-13-02", false},
		{"2020-12-32", false},
		{"2021-01-12", true},
		{"2021-1-12", false},
		{"2021-01-2", false},
		{"2020-01-02 15:04", false},
	}

	for _, v := range tests {
		ans := validate.Date(v.input)

		if ans != v.expected {
			t.Errorf("Input: %s, got: %v, want: %v", v.input, ans, v.expected)
		}
	}
}

func TestTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"15:04", true},
		{"05:04", true},
		{"5:04", false},
		{"05:4", false},
	}

	for _, v := range tests {
		ans := validate.Time(v.input)

		if ans != v.expected {
			t.Errorf("Input: %s, got: %v, want: %v", v.input, ans, v.expected)
		}
	}
}

func TestIssueKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"GOJIRA-1", true},
		{"GOJIRA-1910", true},
		{"GOJIRA-19101", true},
		{"gojira-1910", false},
		{"GOJIRA-1910342", false},
		{"GOJIRA1-1", false},
		{"gojira-1-1", false},
	}

	for _, v := range tests {
		ans := validate.IssueKey(&v.input)

		if ans != v.expected {
			t.Errorf("Input: %s, got: %v, want: %v", v.input, ans, v.expected)
		}
	}
}

func TestCommentID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"123234", true},
		{"892390", true},
		{"x23234", false},
		{"23234", false},
		{"23234983745", false},
	}

	for _, v := range tests {
		ans := validate.CommentID(v.input)

		if ans != v.expected {
			t.Errorf("Input: %s, got: %v, want: %v", v.input, ans, v.expected)
		}
	}
}
