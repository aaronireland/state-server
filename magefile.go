//go:build mage
// +build mage

package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	slash       = string(filepath.Separator)
	bin         = "state-server"
	BinDir      = "bin" + slash
	CoverageDir = "cov" + slash
	AssetsDir   = "assets" + slash
	coverFile   = "coverage.txt"
)

func init() {
	rootDir := "."
	gitRepoRoot, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		rootDir = strings.TrimSpace(string(gitRepoRoot))
	}

	BinDir = filepath.Join(rootDir, BinDir)
	CoverageDir = filepath.Join(rootDir, CoverageDir)
	AssetsDir = filepath.Join(rootDir, AssetsDir)
	bin = filepath.Join(BinDir, bin)
	coverFile = filepath.Join(CoverageDir, coverFile)
}

func Clean() error {

	if locked() {
		if _, err := sh.Output(bin, "stop"); err != nil {
			return err
		}
	}
	_ = sh.Rm(BinDir)
	_ = sh.Rm(CoverageDir)

	return nil
}

func Build() error {
	mg.Deps(Clean)

	if err := os.Mkdir(BinDir, 0700); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to build bin directory %s: %w", BinDir, err)
	}

	return sh.RunV("go", "build", "-o", BinDir, "")
}

func Test() error {
	mg.Deps(Build)

	if err := os.Mkdir(CoverageDir, 0700); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to build coverage directory %s: %w", CoverageDir, err)
	}

	coverProfile := fmt.Sprintf("-coverprofile=%s", coverFile)

	return sh.RunV("go", "test", "-v", "-short", "-timeout", "60s", "-count=1", coverProfile, "./...")
}

func Coverage() error {
	mg.Deps(Test)

	_ = sh.RunV("go", "tool", "cover", fmt.Sprintf("-html=%s", coverFile))
	return nil
}

type Server mg.Namespace

func (Server) Start() error {
	out, err := sh.Output(bin, "start")
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func (Server) Stop() error {
	out, err := sh.Output(bin, "stop")
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func (Server) Seed(asset string) error {
	file, err := os.Open(filepath.Join(AssetsDir, asset))
	if err != nil {
		return err
	}
	defer file.Close()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cmd := fmt.Sprintf("curl --header \"Content-Type: application/json\" --request POST --data '%s' http://localhost:8080/api/v1/state", scanner.Text())
		commands = append(commands, cmd)
	}

	for _, cmd := range commands {
		err := sh.RunV("bash", "-c", cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func locked() bool {
	var lockFiles int
	filepath.WalkDir(BinDir, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ".lock" {
			lockFiles++
		}
		return nil
	})
	return lockFiles > 0
}
