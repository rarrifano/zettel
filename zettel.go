package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
        "regexp"
)

// Base directory
var ZettelDir = filepath.Join(os.Getenv("HOME"), "Zettelkasten")

// Version number
const version = "0.1"

// Usage info
func usage() {
	fmt.Println(`Usage: zettel [OPTION] [ARGUMENT]

Options:
  -n, --new TITLE           Create a new note with the given title
  -o, --open QUERY          Open notes by matching filename or content
  -l, --list                List all notes
  -i, --index TITLE TAGS... Create an index based on one or more tags
  -t, --tags                List all unique tags
      --completion          Generate bash completion script
  -V, --version             Display version information
  -h, --help                Display this help message`)
}

// Ensure directory exists
func checkDirectory() error {
	if _, err := os.Stat(ZettelDir); os.IsNotExist(err) {
		return os.MkdirAll(ZettelDir, os.ModePerm)
	}
	return nil
}

// Create a new note
func newNote(title string) error {
	if title == "" {
		return errors.New("missing title")
	}

	fileName := fmt.Sprintf("%s-%s.md", time.Now().Format("200601021504"), strings.ReplaceAll(title, " ", "-"))
	filePath := filepath.Join(ZettelDir, fileName)

	content := fmt.Sprintf("# %s\n\n#tagme\n\n", title)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return err
	}

	fmt.Println(fileName) // Minimal output for scripting
	return openEditor(filePath)
}

// Open notes by query
func openNotes(query string) error {
	files, err := filepath.Glob(filepath.Join(ZettelDir, "*.md"))
	if err != nil {
		return err
	}

	var matches []string
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		if strings.Contains(filepath.Base(file), query) || strings.Contains(string(content), query) {
			matches = append(matches, file)
		}
	}

	switch len(matches) {
	case 0:
		return errors.New("no matching notes found")
	case 1:
		return openEditor(matches[0])
	default:
		for i, match := range matches {
			fmt.Printf("%d. %s\n", i+1, filepath.Base(match))
		}
		fmt.Print("Select a note: ")
		var choice int
		_, err := fmt.Scanf("%d", &choice)
		if err != nil || choice < 1 || choice > len(matches) {
			return errors.New("invalid choice")
		}
		return openEditor(matches[choice-1])
	}
}

// List all notes
func listNotes() error {
	files, err := filepath.Glob(filepath.Join(ZettelDir, "*.md"))
	if err != nil {
		return err
	}
	for _, file := range files {
		fmt.Println(filepath.Base(file))
	}
	return nil
}

// Create an index based on tags
func createIndex(title string, tags []string) error {
	if title == "" || len(tags) == 0 {
		return errors.New("title and at least one tag are required")
	}

	fileName := fmt.Sprintf("%s-%s.md", time.Now().Format("200601021504"), strings.ReplaceAll(title, " ", "-"))
	filePath := filepath.Join(ZettelDir, fileName)

	files, err := filepath.Glob(filepath.Join(ZettelDir, "*.md"))
	if err != nil {
		return err
	}

	var links []string
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		for _, tag := range tags {
			if strings.Contains(string(content), "#"+tag) {
				links = append(links, fmt.Sprintf("- [[%s]]", filepath.Base(file)))
				break
			}
		}
	}

	content := fmt.Sprintf("# %s\n\n%s\n", title, strings.Join(links, "\n"))
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return err
	}

	fmt.Println(fileName)
	return openEditor(filePath)
}

// List all unique tags
func listTags() error {
	files, err := filepath.Glob(filepath.Join(ZettelDir, "*.md"))
	if err != nil {
		return err
	}

	// Regular expression to match valid tags (letters, digits, and '#' only)
	validTagRegex := regexp.MustCompile(`^#[a-zA-Z0-9]+$`)

	tags := make(map[string]bool)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		for _, word := range strings.Fields(string(content)) {
			if strings.HasPrefix(word, "#") && validTagRegex.MatchString(word) {
				tags[word] = true
			}
		}
	}

	for tag := range tags {
		fmt.Println(tag)
	}
	return nil
}

// Open editor
func openEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Generate bash completion
func generateCompletion() {
	fmt.Println(`# Bash Completion
_zettel_completion() {
    local cur opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts="--new --open --list --index --tags --completion -n -o -l -i -t -V -h"

    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
}
complete -F _zettel_completion zettel`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	option := os.Args[1]
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	err := checkDirectory()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch option {
	case "-n", "--new":
		err = newNote(strings.Join(args, " "))
	case "-o", "--open":
		err = openNotes(strings.Join(args, " "))
	case "-l", "--list":
		err = listNotes()
	case "-i", "--index":
		err = createIndex(args[0], args[1:])
	case "-t", "--tags":
		err = listTags()
	case "--completion":
		generateCompletion()
	case "-V", "--version":
		fmt.Println("zettel version", version)
	case "-h", "--help":
		usage()
	default:
		usage()
		err = errors.New("invalid option")
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
