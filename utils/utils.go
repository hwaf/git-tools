package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/gonuts/iochan"
)

func SplitLines(b []byte) []string {
	lines := []string{}
	r := bufio.NewReader(bytes.NewBuffer(b))
	for line := range iochan.ReaderChan(r, "\n") {
		lines = append(lines, line)
	}
	return lines
}

func PathExists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func HandleErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error**: %v\n", err.Error())
		os.Exit(1)
	}
}

func init() {
	// make sure we get the correct collation stuff
	err := os.Setenv("LC_MESSAGES", "C")
	HandleErr(err)
}

// EOF
