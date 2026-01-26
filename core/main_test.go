package core

import (
	"os"
	"strings"
	"testing"
)

type testcase struct {
	filePath string
	expected string
}

const TESTCASES_DIR string = "../testcases"

func readFile(filePath string) string {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

func TestExamples(t *testing.T) {
	files, err := os.ReadDir(TESTCASES_DIR)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".bs") {
			continue
		}
		t.Run(file.Name(), func(t *testing.T) {
			expected := readFile(TESTCASES_DIR + "/" + file.Name() + ".stdout")

			buf := new(strings.Builder)
			opts := new(Opts)
			opts.ostr = buf
			opts.estr = buf

			rc := RunFile(opts, TESTCASES_DIR+"/"+file.Name())

			if rc > 0 {
				t.Errorf("program '%s' executed with nonzero exit ckde: %d", file.Name(), rc)
			}

			actual := buf.String()

			if expected != actual {
				t.Errorf("ran '%s' expected '%s', got '%s'", file.Name(), expected, actual)
			}

		})
	}
}
