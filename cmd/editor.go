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
	"io/ioutil"
	"os"
	"os/exec"
)

const DefaultEditor = "vim"

func openFileInEditor(filename string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = DefaultEditor
	}

	executable, err := exec.LookPath(editor)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	cmd := exec.Command(executable, filename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run() //nolint:wrapcheck
}

func captureInputFromEditor(text, pattern string) ([]byte, error) {
	file, err := ioutil.TempFile(os.TempDir(), pattern)
	if err != nil {
		return []byte{}, fmt.Errorf("%w", err)
	}

	filename := file.Name()

	if text != "" {
		err := ioutil.WriteFile(filename, []byte(text), 0600)
		if err != nil {
			return []byte{}, fmt.Errorf("%w", err)
		}
	}

	defer os.Remove(filename)

	if err = file.Close(); err != nil {
		return []byte{}, fmt.Errorf("%w", err)
	}

	f, _ := os.Stat(filename)
	modtime := f.ModTime()

	if err = openFileInEditor(filename); err != nil {
		return []byte{}, err
	}

	f, _ = os.Stat(filename)
	if modtime.Equal(f.ModTime()) {
		return []byte{}, nil
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return []byte{}, fmt.Errorf("%w", err)
	}

	return bytes, nil
}
