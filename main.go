package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultHome   = "zettelkasten"
	noteExtension = ".md"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	zettelHome, err := getZettelHome()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "new":
		createNewNote(zettelHome)
	case "edit":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a note ID")
			os.Exit(1)
		}
		editNote(zettelHome, os.Args[2])
	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Please provide a search query")
			os.Exit(1)
		}
		searchNotes(zettelHome, os.Args[2])
	case "link":
		if len(os.Args) < 4 {
			fmt.Println("Please provide source and target IDs")
			os.Exit(1)
		}
		linkNotes(zettelHome, os.Args[2], os.Args[3])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Zettelkasten CLI

Usage:
  zettel new                Create new note
  zettel edit <ID>          Edit existing note
  zettel search <query>     Search notes
  zettel link <src> <dest>  Link two notes

Environment variables:
  ZETTEL_HOME   Notes directory (default: ~/zettelkasten)
  EDITOR        Preferred text editor`)
}

func getZettelHome() (string, error) {
	if home := os.Getenv("ZETTEL_HOME"); home != "" {
		return home, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, defaultHome), nil
}

func generateID() string {
	return time.Now().Format("20060102150405")
}

func createNewNote(zettelHome string) {
	if err := os.MkdirAll(zettelHome, 0755); err != nil {
		fmt.Println("Error creating directory:", err)
		os.Exit(1)
	}

	id := generateID()
	notePath := filepath.Join(zettelHome, id+noteExtension)

	if err := os.WriteFile(notePath, []byte("# "+id+"\n"), 0644); err != nil {
		fmt.Println("Error creating note:", err)
		os.Exit(1)
	}

	if err := openEditor(notePath); err != nil {
		fmt.Println("Error opening editor:", err)
		os.Exit(1)
	}

	fmt.Println("Created new note:", id)
}

func editNote(zettelHome, id string) {
	notePath := filepath.Join(zettelHome, id+noteExtension)
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		fmt.Println("Note does not exist:", id)
		os.Exit(1)
	}

	if err := openEditor(notePath); err != nil {
		fmt.Println("Error opening editor:", err)
		os.Exit(1)
	}
}

func searchNotes(zettelHome, query string) {
	err := filepath.Walk(zettelHome, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == noteExtension {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if strings.Contains(string(content), query) {
				fmt.Println("Found in:", filepath.Base(path[:len(path)-len(noteExtension)]))
			}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Search error:", err)
	}
}

func linkNotes(zettelHome, src, dest string) {
	srcPath := filepath.Join(zettelHome, src+noteExtension)
	destPath := filepath.Join(zettelHome, dest+noteExtension)

	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		fmt.Println("Source note does not exist:", src)
		os.Exit(1)
	}

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		fmt.Println("Destination note does not exist:", dest)
		os.Exit(1)
	}

	f, err := os.OpenFile(srcPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening note:", err)
		os.Exit(1)
	}
	defer f.Close()

	if _, err = f.WriteString(fmt.Sprintf("\n[[%s]]\n", dest)); err != nil {
		fmt.Println("Error writing link:", err)
		os.Exit(1)
	}

	fmt.Printf("Linked %s -> %s\n", src, dest)
}

func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return fmt.Errorf("EDITOR environment variable not set")
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
