package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mana-fwk/git-tools/utils"
)

func get_gitroot() (string, error) {
	// create temporary root tempdir
	g_gitroot, err := ioutil.TempDir("", "git-rm-submodule-test-")
	if err != nil {
		return "", err
	}

	// prepare a few dummy git repositories
	cmd := exec.Command("git", "init", "meta-repo")
	cmd.Dir = g_gitroot
	err = cmd.Run()
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	// sub-repo0
	cmd = exec.Command("git", "init", "sub-repo-0")
	cmd.Dir = g_gitroot
	err = cmd.Run()
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	err = ioutil.WriteFile(
		filepath.Join(g_gitroot, "sub-repo-0", "file.txt"),
		[]byte("this is a file in sub-repo-0\n"),
		0700,
	)
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	cmd = exec.Command("git", "add", "file.txt")
	cmd.Dir = filepath.Join(g_gitroot, "sub-repo-0")
	err = cmd.Run()
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	cmd = exec.Command("git", "commit", "-m", "adding file.txt")
	cmd.Dir = filepath.Join(g_gitroot, "sub-repo-0")
	err = cmd.Run()
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	// sub-repo1
	cmd = exec.Command("git", "init", "sub-repo-1")
	cmd.Dir = g_gitroot
	err = cmd.Run()
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	err = ioutil.WriteFile(
		filepath.Join(g_gitroot, "sub-repo-1", "file.txt"),
		[]byte("this is a file in sub-repo-1\n"),
		0700,
	)
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	cmd = exec.Command("git", "add", "file.txt")
	cmd.Dir = filepath.Join(g_gitroot, "sub-repo-1")
	err = cmd.Run()
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	cmd = exec.Command("git", "commit", "-m", "adding file.txt")
	cmd.Dir = filepath.Join(g_gitroot, "sub-repo-1")
	err = cmd.Run()
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	gitroot := filepath.Join(g_gitroot, "work")
	cmd = exec.Command("git", "init", "work")
	cmd.Dir = g_gitroot
	err = cmd.Run()
	if err != nil {
		return "", err
	}

	// add 2 submodules
	for _, sub := range []string{"sub-repo-0", "sub-repo-1"} {

		cmd = exec.Command(
			"git",
			"submodule", "add",
			filepath.Join(g_gitroot, sub),
			filepath.Join("src", sub),
		)
		cmd.Dir = gitroot
		err = cmd.Run()
		if err != nil {
			return "", err
		}
		cmd = exec.Command(
			"git", "commit", "-m",
			fmt.Sprintf("adding subrepo [%s]", sub),
		)
		cmd.Dir = gitroot
		err = cmd.Run()
		if err != nil {
			return "", err
		}
	}

	return g_gitroot, nil
}

func TestNormalSubmodule(t *testing.T) {

	g_gitroot, err := get_gitroot()
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(g_gitroot)

	gitroot := filepath.Join(g_gitroot, "work")

	cmd := exec.Command("git", "submodule", "foreach", "")
	cmd.Dir = gitroot
	bout, err := cmd.Output()
	if err != nil {
		t.Error(err)
	}
	out := string(bout)
	{
		const expected = "Entering 'src/sub-repo-0'\nEntering 'src/sub-repo-1'\n"
		if out != expected {
			t.Errorf("got [%v]. expected [%v]\n", out, expected)
		}
	}

	// remove one submodule
	cmd = exec.Command("git", "rm-submodule", "src/sub-repo-0")
	cmd.Dir = gitroot
	bout, err = cmd.CombinedOutput()
	if err != nil {
		t.Errorf("error: %v\noutput: %v\n", err, string(bout))
	}

	// check the submodule has been removed
	cmd = exec.Command("git", "submodule", "foreach", "")
	cmd.Dir = gitroot
	bout, err = cmd.Output()
	if err != nil {
		t.Error(err)
	}
	out = string(bout)
	{
		const expected = "Entering 'src/sub-repo-1'\n"
		if out != expected {
			t.Errorf("got [%v]. expected [%v]\n", out, expected)
		}
	}

	// check the submodule has been also physically removed
	if utils.PathExists(filepath.Join("src", "sub-repo-0")) {
		t.Errorf("directory [%s] still exists", filepath.Join("src", "sub-repo-0"))
	}

}

func TestUncleanSubmodule(t *testing.T) {
	g_gitroot, err := get_gitroot()
	if err != nil {
		t.Error(err)
	}
	//fmt.Printf(">>> %s\n", g_gitroot)
	defer os.RemoveAll(g_gitroot)

	gitroot := filepath.Join(g_gitroot, "work")

	// add a new file under sup-repo-0 to make it non-clean
	err = ioutil.WriteFile(
		filepath.Join(gitroot, "src", "sub-repo-0", "non-clean-file.txt"),
		[]byte("non-clean file in sub-repo-0\n"),
		0700,
	)
	if err != nil {
		t.Error(err)
	}

	// removing any submodule should fail
	for _, sub := range []string{"sub-repo-0", "sub-repo-1"} {
		cmd := exec.Command(
			"git",
			"rm-submodule", filepath.Join("src", sub),
		)
		cmd.Dir = gitroot
		bout, err := cmd.CombinedOutput()
		if err == nil {
			t.Errorf("rm-submodule %s should have failed.\noutput: %v\n",
				filepath.Join("src", sub), string(bout))
		}
	}

}

// EOF
