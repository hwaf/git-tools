package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	// "path/filepath"
	"regexp"
	// "strings"

	"github.com/mana-fwk/git-tools/utils"
)

var (
	g_exitcode    = flag.Bool("exit-code", false, "return an exit code instead of a message")
	g_unstaged    = flag.Bool("unstaged", true, "check for unstaged files")
	g_uncommitted = flag.Bool("uncommitted", true, "check for staged but uncommitted files")
	g_untracked   = flag.Bool("untracked", true, "check for untracked files")
	g_unmerged    = flag.Bool("unmerged", true, "check for unmerged files")
	g_ignoresub   = flag.String("ignore-submodules", "<nil>", "ignore submodules, optionally specifying defaults to \"all\"")
	g_warn        = flag.Bool("warn", false, "do not issue an error, but just a warning")
)

var g_output bool = false

func output(msg string) {
	g_output = true
	fmt.Fprint(os.Stderr, msg)
}

func main() {
	flag.Parse()
	var err error

	pwd, err := os.Getwd()
	utils.HandleErr(err)

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

	g_lines := utils.SplitLines(bout)
	//status := string(bout)

	if *g_unstaged {
		var matched bool
		for _, line := range g_lines {
			// echo "$status" | grep -E '^[ MARC][MD]' -q
			var i bool
			i, err = regexp.MatchString("^( |M|A|R|C)(M|D)", line)
			utils.HandleErr(err)
			if !i && line != "" {
				matched = true
				break
			}
		}

		if matched {
			if *g_exitcode {
				os.Exit(1)
			}
			output("There are unstaged changes. Use \"git add <file>\" to add.\n")
		}

		if *g_ignoresub != "<nil>" {
			// git ls-files --error-unmatch --stage | grep -E '^160000' | sed -e 's/^.* //' | tr '\n' ' '
			bout, err = exec.Command(
				"git", "ls-files", "--error-unmatch", "--stage",
			).Output()
			utils.HandleErr(err)
			lines := utils.SplitLines(bout)
			matched := false
			subm := []string{}
			for _, line := range lines {
				var i bool
				i, err = regexp.MatchString("^160000", line)
				utils.HandleErr(err)
				if !i {
					matched = true
					subm = append(subm, line)
				}
			}
			if matched {
				cmdargs := []string{"status", "--porcelain", "--"}
				cmdargs = append(cmdargs, subm...)
				bout, err = exec.Command(
					"git", cmdargs...,
				).Output()
				utils.HandleErr(err)
				lines = utils.SplitLines(bout)
				matched = false
				for _, line := range lines {
					var i bool
					i, err = regexp.MatchString("^( |M|A|R|C)(M|D)", line)
					if !i {
						matched = true
						output("There is modified content in submodules.\n")
						break
					}
				}
			}
		}
	}

	testfct := func(lines []string, pattern, message string) {
		var matched bool
		for _, line := range lines {
			var i bool
			i, err = regexp.MatchString(pattern, line)
			utils.HandleErr(err)
			if !i && line != "" {
				matched = true
				break
			}
		}

		if matched {
			if *g_exitcode {
				os.Exit(1)
			}
			output(message)
		}
	}

	if *g_unmerged {
		testfct(
			g_lines,
			"^(DD|AU|UD|UA|DU|AU|UU)",
			"There are unmerged files. Use \"git add <file>\" when merged.\n",
		)
	}

	if *g_uncommitted {
		testfct(
			g_lines,
			"^(M|A|D|R|C)",
			"There are uncommitted files. Use \"git commit\" to commit.\n",
		)
	}

	if *g_untracked {
		testfct(
			g_lines,
			"^?.*",
			"There are untracked files not in .gitignore. Try \"make clean\" to remove temporary files.\n",
		)
	}

	if g_output {
		if *g_warn {
			fmt.Fprintf(os.Stderr, "Warning in %s\n", pwd)
			os.Exit(0)
		} else {
			fmt.Fprintf(os.Stderr, "Error in %s\n", pwd)
			os.Exit(1)
		}
	}
}

// EOF
