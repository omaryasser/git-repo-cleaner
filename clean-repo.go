package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	ConsoleRedColorCode   = "\033[31m"
	ConsoleGreenColorCode = "\033[32m"
	ConsoleResetColorCode = "\033[0m"
)

var repoAbsolutePath = flag.String("repo-absolute-path", "", "The absolute path to the repo to be cleaned")
var mainBranchName = flag.String("main-branch-name", "", "The name of the main branch (e.g. master)")

func checkoutToMainBranch() {
	cmd := exec.Command("git", "checkout", *mainBranchName)
	cmd.Dir = *repoAbsolutePath

	if err := cmd.Run(); err != nil {
		log.Fatalf("Could not checkout to branch %s: %v", *mainBranchName, err)
	}
}

func getAllFilesSavedInGit() []string {
	cmd := exec.Command("git", "rev-list", "--objects", "--all")
	cmd.Dir = *repoAbsolutePath
	lines, err := runCmdAndGetOutputLines(cmd)
	if err != nil {
		log.Fatalf("Could not find all objects: %v", err)
	}
	files := make([]string, 0)
	for _, line := range lines {
		splitted := strings.Split(strings.TrimSpace((line)), " ")
		if len(splitted) == 2 {
			files = append(files, splitted[1])
		}
	}

	return files
}

func runCmdAndGetOutputLines(cmd *exec.Cmd) ([]string, error) {
	cmd.Dir = *repoAbsolutePath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(string(out), "\n"), nil
}

func isFileInRepoDir(file string) bool {
	if _, err := os.Stat(filepath.Join(*repoAbsolutePath, file)); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func isFileGitIgnored(file string) bool {
	cmd := exec.Command("git", "check-ignore", file)
	cmd.Dir = *repoAbsolutePath
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Check git check-ignore --help
			// An exit code of 1 means that no files are to be ignored given the input pattern
			if err.(*exec.ExitError).ExitCode() == 1 {
				return false
			}
		}
		log.Fatalf("Could not check if file %s is git-ignored: %v", file, err)
	}
	return true
}

func filterFilesToBeRemoved(files []string) []string {
	filesToBeRemoved := make([]string, 0)
	for _, file := range files {
		if !isFileInRepoDir(file) || isFileGitIgnored(file) {
			filesToBeRemoved = append(filesToBeRemoved, file)
		}
	}
	return filesToBeRemoved
}

func printFilesToBeRemoved(files []string) {
	fmt.Print(ConsoleRedColorCode)
	fmt.Printf("All of the following files will be removed either because they are ignored by git or because they are not present in the repo directory on branch %s\n", *mainBranchName)
	fmt.Print(ConsoleResetColorCode)

	for _, file := range files {
		fmt.Println(file)
	}
}

func takeUserConsent() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(ConsoleRedColorCode)
	fmt.Println("Start cleaning up the git objects mentioned above? (Yes/No) [Default: No]")
	fmt.Print(ConsoleResetColorCode)

	text, _ := reader.ReadString('\n')
	if strings.TrimSpace(text) != "Yes" {
		log.Fatal("User hasn't accepted")
	}
}

func removeFilesFromHistory(files []string) {
	for _, file := range files {

		fmt.Println()
		fmt.Print(ConsoleGreenColorCode)
		fmt.Printf("Will starting removing file: %s\n", file)
		fmt.Println(ConsoleResetColorCode)

		cmd := exec.Command("git", "filter-repo", "--force", "--invert-paths", "--path", file)
		cmd.Dir = *repoAbsolutePath
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to rewrite history to remove file: %s: %v", file, err)
		}
	}
}

func printFiltering(stop <-chan struct{}) {
	dots := 1

	for {
		select {
		case <-stop:
			return
		default:
			fmt.Print("\033[2K\rFiltering files to be removed")
			for i := 0; i < dots; i++ {
				fmt.Print(".")
			}
			dots %= 5
			dots += 1
			time.Sleep(time.Duration(time.Millisecond * 500))
		}
	}
}

func main() {
	flag.Parse()

	if *repoAbsolutePath == "" {
		log.Fatal("repo-absolute-path flag must not be empty")
	}

	if *mainBranchName == "" {
		log.Fatal("main-branch-name flag must not be empty")
	}

	checkoutToMainBranch()

	files := getAllFilesSavedInGit()
	filesLenInitially := len(files)
	fmt.Printf("Git is currently saving objects for %d files.\n", filesLenInitially)

	filteringDone := make(chan struct{})
	go func() {
		printFiltering(filteringDone)
	}()

	filesToBeRemoved := filterFilesToBeRemoved(files)
	close(filteringDone)

	printFilesToBeRemoved(filesToBeRemoved)
	takeUserConsent()
	removeFilesFromHistory(filesToBeRemoved)

	fmt.Printf("Git was saving %d objects and now is saving %d objects.\n", filesLenInitially, len(getAllFilesSavedInGit()))
}
