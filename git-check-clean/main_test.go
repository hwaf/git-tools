package main

import (
	//"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	//"github.com/mana-fwk/git-tools/utils"
)

func get_gitroot() (string, error) {
	// create temporary root tempdir
	g_gitroot, err := ioutil.TempDir("", "git-check-clean-")
	if err != nil {
		return "", err
	}

	// prepare a git repository
	cmd := exec.Command("git", "init", "repo")
	cmd.Dir = g_gitroot
	err = cmd.Run()
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	gitroot := filepath.Join(g_gitroot, "repo")
	err = ioutil.WriteFile(
		filepath.Join(gitroot, "file.txt"),
		[]byte("this is a file in git-repo\n"),
		0700,
	)
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	cmd = exec.Command("git", "add", "file.txt")
	cmd.Dir = gitroot
	err = cmd.Run()
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	cmd = exec.Command("git", "commit", "-m", "adding file.txt")
	cmd.Dir = gitroot
	err = cmd.Run()
	if err != nil {
		_ = os.RemoveAll(g_gitroot)
		return "", err
	}

	return g_gitroot, nil
}

func TestGitRepo(t *testing.T) {
	g_gitroot, err := get_gitroot()
	if err != nil {
		t.Error(err)
	}
	//fmt.Printf(">>> %s\n", g_gitroot)
	defer os.RemoveAll(g_gitroot)

	gitroot := filepath.Join(g_gitroot, "repo")

	// check the repository is clean
	cmd := exec.Command(
		"git",
		"check-clean", "-exit-code",
	)
	cmd.Dir = gitroot
	err = cmd.Run()
	if err != nil {
		t.Error(err)
	}

	// add a new file to make it non-clean
	err = ioutil.WriteFile(
		filepath.Join(gitroot, "non-clean-file.txt"),
		[]byte("non-clean file in repo\n"),
		0700,
		)
	if err != nil {
		t.Error(err)
	}

	// check the repository is non-clean
	cmd = exec.Command(
	"git",
	"check-clean", "-exit-code",
	)
	cmd.Dir = gitroot
	err = cmd.Run()
	if err == nil {
		t.Errorf("repository [%s] is detected as clean, but it should NOT", gitroot)
	}

	// remove file
	err = os.Remove(filepath.Join(gitroot, "non-clean-file.txt"))
	if err != nil {
		t.Error(err)
	}

	// check the repository is clean
	cmd = exec.Command(
		"git",
		"check-clean", "-exit-code",
	)
	cmd.Dir = gitroot
	err = cmd.Run()
	if err != nil {
		t.Error(err)
	}

	// modify a file to make the repo non-clean
	err = ioutil.WriteFile(
		filepath.Join(gitroot, "file.txt"),
		[]byte("this is a file in git-repo. a modified one\n"),
		0700,
	)
	if err != nil {
		t.Error(err)
	}

	cmd = exec.Command(
	"git",
	"check-clean", "-exit-code",
	)
	cmd.Dir = gitroot
	err = cmd.Run()
	if err == nil {
		t.Errorf("repository [%s] is detected as clean, but it should NOT (modified file)", gitroot)
	}

	// commit the change to make the repo clean again
	cmd = exec.Command(
		"git", "add", "file.txt",
	)
	cmd.Dir = gitroot
	err = cmd.Run()
	if err != nil {
		t.Error(err)
	}
	cmd = exec.Command(
		"git", "commit", "-m", "modified file.txt",
	)
	cmd.Dir = gitroot
	err = cmd.Run()
	if err != nil {
		t.Error(err)
	}

	cmd = exec.Command(
		"git",
		"check-clean", "-exit-code",
	)
	cmd.Dir = gitroot
	err = cmd.Run()
	if err != nil {
		t.Errorf("repository [%s] is detected as non-clean, but it should (modified file has been committed)", gitroot)
	}

}

// EOF
