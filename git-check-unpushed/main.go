package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	//"path/filepath"
	"regexp"
	//"strings"

	"github.com/mana-fwk/git-tools/utils"
)

var (
	g_exitcode = flag.Bool("exit-code", true, "return an exit code instead of a message")
	g_verbose  = flag.Bool("verbose", false, "")
)

func printf(format string, args ...interface{}) (n int, err error) {
	if *g_verbose {
		return fmt.Printf(format, args...)
	}
	return
}
func debug(cmd *exec.Cmd) {
	if *g_verbose {
		dir := cmd.Dir
		if dir == "" {
			dir, _ = os.Getwd()
		}
		printf(">>> %v (%s)\n", cmd.Args, dir)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
}

func main() {
	flag.Parse()
	var err error

	switch flag.NArg() {
	case 0:
		// ok
	default:
		err = fmt.Errorf("no argument expected")
		utils.HandleErr(err)
	}

	unpushed := make([]string, 0, 2)

	// cycle all branches
	cmd := exec.Command(
		"git",
		"branch", "-vv", "--no-color", //"--no-merged",
	)
	//debug(cmd)
	bout, err := cmd.Output()
	utils.HandleErr(err)

	lines := utils.SplitLines(bout)
	//printf(">>> lines:\n%v\n", lines)

	branch_re := regexp.MustCompile(
		`\* (?P<branch>.*?) (?P<hash>(\w|\d)*?) .*`,
	)

	remote_re := regexp.MustCompile(
		`^(\*| ) (?P<branch>.*?) (?P<hash>(\w|\d)*?) [[](?P<remote>.*?)[]].*?`,
	)

	ahead_re := regexp.MustCompile(
		`.*?: ahead \(.(?P<ahead>.*?)*\)`,
	)

	for _, line := range lines {
		//printf(">>> %q\n", line)
		sub := branch_re.FindStringSubmatch(line)
		if sub == nil {
			continue
		}
		// printf("   sub: %q\n", strings.Join(sub, ","))
		// printf("   branch=%q\n", sub[1])
		// printf("   hash=%q\n", sub[2])
		if remote_re.MatchString(line) {
			// printf("   ==> remote branch\n")
			continue
		}
		ahead_sub := ahead_re.FindStringSubmatch(line)
		if ahead_sub == nil {
			continue
		}
		unpushed = append(
			unpushed,
			fmt.Sprintf("branch [%v] is ahead [%s] commit(s)", sub[1], ahead_sub[1]),
		)
	}

	if len(unpushed) == 0 {
		return
	}

	for _, msg := range unpushed {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
	if *g_exitcode {
		os.Exit(1)
	}
}

// EOF
