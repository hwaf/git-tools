package main

import (
	"fmt"
	"flag"
	"os"
	"os/exec"
	// "path/filepath"
	"regexp"
	// "strings"

	"github.com/mana-fwk/git-tools/utils"
)

var (
	g_exitcode = flag.Bool("exit-code", false, "return an exit code instead of a message")
	g_unstaged = flag.Bool("unstaged", true, "check for unstaged files")
	g_uncommitted = flag.Bool("uncommitted", true, "check for staged but uncommitted files")
	g_untracked = flag.Bool("untracked", true, "check for untracked files")
	g_unmerged = flag.Bool("unmerged", true, "check for unmerged files")
	g_ignoresub = flag.String("ignore-submodules", "<nil>", "ignore submodules, optionally specifying defaults to \"all\"")
	g_warn = flag.Bool("warn", false, "do not issue an error, but just a warning")
	)

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

	var git *exec.Cmd
	if *g_ignoresub != "<nil>" {
		git = exec.Command(
			"git", "status", "--porcelain", "--ignore-submodules=",
			*g_ignoresub,
			)
	} else {
		git = exec.Command("git", "status", "--porcelain")
	}
	bout, err := git.Output()
	utils.HandleErr(err)

	//lines := utils.SplitLines(bout)
	//status := string(bout)
	
	if *g_unstaged {
		var matched bool
		matched, err = regexp.Match("^( |M|A|R|C)(M|D)", bout)
		utils.HandleErr(err)

		if matched {
			if *g_exitcode {
				os.Exit(1)
			}
			
		}
	}
}

// EOF
