package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	_testdata       = "testdata"
	_expectedOutput = "expected.output"
)

func TestRealCases(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	require.NoError(t, err, "unable to determine cwd")
	testdata := filepath.Join(cwd, _testdata)
	filepath.Walk(testdata, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err, "Unexpected error walking testdata")
		if strings.HasSuffix(path, _testdata) {
			// skip the TLD
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		cleanPath := strings.Replace(path, testdata, "", 1)[1:]
		t.Run(cleanPath, func(t *testing.T) {
			out := &bytes.Buffer{}
			result := extract(path)
			// TODO check error codes
			result.summarize(out)
			expOut := filepath.Join(path, _expectedOutput)

			if bs, err := ioutil.ReadFile(expOut); err != nil {
				assert.Fail(t, "Unable to read expected output: %v", err)
				require.NoError(t, err, "Unable to read expected error file")
			} else {
				outScrubbed := strings.Replace(out.String(), path, "", -1)
				lines := bufio.NewScanner(bytes.NewBuffer(bs))
				for lines.Scan() {
					line := lines.Text()
					assert.Contains(t, outScrubbed, line)
				}
				require.NoError(t, lines.Err(), "got error scanning output")
			}
		})
		return nil
	})
}

func TestDirOrHere_Variations(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	cases := []struct {
		input    []string
		expected string
	}{
		{nil, cwd},
		{[]string{"."}, cwd},
		{[]string{"testdata"}, filepath.Join(cwd, "testdata")},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%#v", c.input), func(t *testing.T) {
			defer withArgs(c.input...)()
			assert.Equal(t, c.expected, dirOrHere())
		})
	}
}

func withArgs(args ...string) func() {
	old := os.Args
	new := append([]string{os.Args[0]}, args...)
	os.Args = new
	return func() {
		os.Args = old
	}
}
