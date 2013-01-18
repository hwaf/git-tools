package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mana-fwk/git-tools/utils"
)

var (
	g_no_commit = flag.Bool("no-commit", false, "do not commit the result")
	g_verbose   = flag.Bool("verbose", false, "")
)

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
		utils.HandleErr(err)
	case 1:
		dir = flag.Args()[0]
	default:
		// TODO: allow to remove n submodules in one go...
		err = fmt.Errorf("you need to give a submodule directory to remove")
		utils.HandleErr(err)
	}

	// make sure we get the correct collation stuff
	err = os.Setenv("LC_MESSAGES", "C")
	utils.HandleErr(err)

	// check 'dir' is a valid submodule
	if !utils.PathExists(dir) {
		err = fmt.Errorf("no such directory [%s]", dir)
		utils.HandleErr(err)
	}

	git := exec.Command(
		"git", "ls-files",
		"--error-unmatch", "--stage", "--", dir,
	)
	bout, err := git.Output()
	utils.HandleErr(err)
	out := strings.Trim(string(bout), " \r\n")
	if out == "" {
		err = fmt.Errorf("no such submodule [%s]", dir)
		utils.HandleErr(err)
	}

	// get the full path of the submodule
	git = exec.Command("git", "ls-files", "--full-name", dir)
	bout, err = git.Output()
	utils.HandleErr(err)

	lines := utils.SplitLines(bout)
	dir = strings.Trim(lines[0], " \r\n")

	if *g_verbose {
		fmt.Printf("found submodule [%s]\n", dir)
	}

	// ensure we are in the toplevel directory
	git = exec.Command("git", "rev-parse", "--show-toplevel")
	bout, err = git.Output()
	utils.HandleErr(err)
	top := strings.Trim(string(bout), " \r\n")
	top, err = filepath.Abs(top)
	utils.HandleErr(err)

	pwd, err := os.Getwd()
	utils.HandleErr(err)
	defer os.Chdir(pwd)

	err = os.Chdir(top)
	utils.HandleErr(err)

	if *g_verbose {
		fmt.Printf("root [%s]\n", top)
	}

	// check if submodule is clean
	git = exec.Command("git", "check-clean")
	debug(git)
	utils.HandleErr(git.Run())

	// check for unpushed changes
	git = exec.Command("git", "check-unpushed")
	debug(git)
	utils.HandleErr(git.Run())

	// check for local non-tracking-branches
	git = exec.Command("git", "check-non-tracking")
	debug(git)
	utils.HandleErr(git.Run())

	// find the real git-dir
	git = exec.Command("git", "rev-parse", "--git-dir")
	bout, err = git.Output()
	utils.HandleErr(err)

	gitdir := strings.Trim(string(bout), " \r\n")
	gitdir, err = filepath.Abs(gitdir)
	utils.HandleErr(err)

	if *g_verbose {
		fmt.Printf("gitdir [%s]\n", gitdir)
	}

	// get submodule url
	url := "unknown"
	git = exec.Command(
		"git", "config", "--get",
		fmt.Sprintf("submodule.%s.url", dir),
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
	utils.HandleErr(err)

	git = exec.Command(
		"git", "config", "--remove-section",
		fmt.Sprintf("submodule.%s", dir),
	)
	debug(git)
	err = git.Run()
	utils.HandleErr(err)

	git = exec.Command("git", "rm", "--cached", dir)
	debug(git)
	err = git.Run()
	utils.HandleErr(err)

	err = os.RemoveAll(dir)
	utils.HandleErr(err)

	// remove git dir as well.
	if utils.PathExists(filepath.Join(gitdir, "modules", dir)) {
		err = os.RemoveAll(filepath.Join(gitdir, "modules", dir))
		utils.HandleErr(err)
	}

	if *g_no_commit {
		return
	}

	// commit changes
	git = exec.Command("git", "add", ".gitmodules")
	debug(git)
	err = git.Run()
	utils.HandleErr(err)

	git = exec.Command(
		"git", "commit", "-m",
		fmt.Sprintf("removed submodule [%s] (url=%s)", dir, url),
	)
	debug(git)
	err = git.Run()
	utils.HandleErr(err)

	// TODO: commit in super repositories of this one as well
}

// EOF
