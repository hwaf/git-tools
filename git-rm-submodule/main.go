package main

import (
	"bytes"
	"bufio"
	"fmt"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gonuts/iochan"
)

var (
	g_no_commit = flag.Bool("no-commit", false, "do not commit the result")
	g_verbose = flag.Bool("verbose", false, "")
)

func split_lines(b []byte) []string {
	lines := []string{}
	r := bufio.NewReader(bytes.NewBuffer(b))
	for line := range iochan.ReaderChan(r, "\n") {
		lines = append(lines, line)
	}
	return lines
}

func path_exists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func handle_err(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "**error**: %v\n", err.Error())
		os.Exit(1)
	}
}

func debug(cmd *exec.Cmd) {
	if *g_verbose {
		dir := cmd.Dir
		if dir == "" {
			dir, _ = os.Getwd()
		}
		fmt.Printf(">>> %v (%s)\n", cmd.Args, dir)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
}

func main() {
	flag.Parse()
	var err error

	dir := ""
	switch flag.NArg() {
	case 0:
		err = fmt.Errorf("you need to give a submodule directory to remove")
		handle_err(err)
	case 1:
		dir = flag.Args()[0]
	default:
		// TODO: allow to remove n submodules in one go...
		err = fmt.Errorf("you need to give a submodule directory to remove")
		handle_err(err)
	}

	// make sure we get the correct collation stuff
	err = os.Setenv("LC_MESSAGES", "C")
	handle_err(err)

	// check 'dir' is a valid submodule
	if !path_exists(dir) {
		err = fmt.Errorf("no such directory [%s]", dir)
		handle_err(err)
	}

	git := exec.Command(
		"git", "ls-files", 
		"--error-unmatch", "--stage", "--", dir,
		)
	bout, err := git.Output()
	handle_err(err)
	out := strings.Trim(string(bout), " \r\n")
	if out == "" {
		err = fmt.Errorf("no such submodule [%s]", dir)
		handle_err(err)
	}

	// get the full path of the submodule
	git = exec.Command("git", "ls-files", "--full-name", dir)
	bout, err = git.Output()
	handle_err(err)

	lines := split_lines(bout)
	dir = strings.Trim(lines[0], " \r\n")

	if *g_verbose {
		fmt.Printf("found submodule [%s]\n", dir)
	}
	
	// ensure we are in the toplevel directory
	git = exec.Command("git", "rev-parse", "--show-toplevel")
	bout, err = git.Output()
	handle_err(err)
	top := strings.Trim(string(bout), " \r\n")
	top, err = filepath.Abs(top)
	handle_err(err)

	pwd, err := os.Getwd()
	handle_err(err)
	defer os.Chdir(pwd)

	err = os.Chdir(top)
	handle_err(err)

	if *g_verbose {
		fmt.Printf("root [%s]\n", top)
	}

	// check if submodule is clean
	git = exec.Command("git", "check-clean")
	debug(git)
	handle_err(git.Run())
	
	// check for unpushed changes
	git = exec.Command("git", "check-unpushed")
	debug(git)
	handle_err(git.Run())

	// check for local non-tracking-branches
	git = exec.Command("git", "check-non-tracking")
	debug(git)
	handle_err(git.Run())

	// find the real git-dir
	git = exec.Command("git", "rev-parse", "--git-dir")
	bout, err = git.Output()
	handle_err(err)
	
	gitdir := strings.Trim(string(bout), " \r\n")
	gitdir, err = filepath.Abs(gitdir)
	handle_err(err)
	
	if *g_verbose {
		fmt.Printf("gitdir [%s]\n", gitdir)
	}

	// get submodule url
	url := "unknown"
	git = exec.Command(
		"git", "config", "--get", 
		fmt.Sprintf("submodule.\"%s\".url", dir),
		)
	bout, err = git.Output()
	if err == nil {
		url = strings.Trim(string(bout), " \r\n")
	}
	
	// remove config entries
	git = exec.Command(
		"git", "config", "-f", ".gitmodules", "--remove-section",
		fmt.Sprintf("submodule.%s", dir),
		)
	debug(git)
	err = git.Run()
	handle_err(err)

	git = exec.Command(
		"git", "config", "--remove-section",
		fmt.Sprintf("submodule.%s", dir),
		)
	debug(git)
	err = git.Run()
	handle_err(err)

	git = exec.Command("git", "rm", "--cached", dir)
	debug(git)
	err = git.Run()
	handle_err(err)

	err = os.RemoveAll(dir)
	handle_err(err)

	// remove git dir as well.
	//err = os.RemoveAll(gitdir)
	//handle_err(err)

	if *g_no_commit {
		return
	}

	// commit changes
	git = exec.Command("git", "add", ".gitmodules")
	debug(git)
	err = git.Run()
	handle_err(err)

	git = exec.Command(
		"git", "commit", "-m",
		fmt.Sprintf("removed submodule [%s] (url: %q)", dir, url),
		)
	debug(git)
	err = git.Run()
	handle_err(err)
	
	// TODO: commit in super repositories of this one as well
}


// EOF
