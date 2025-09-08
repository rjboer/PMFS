package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"

	PMFS "github.com/rjboer/PMFS"
)

var heading = color.New(color.FgCyan, color.Bold).SprintfFunc()

// loadEnv loads environment variables from a .env file if it exists.
func loadEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		os.Setenv(key, val)
	}
	if err := s.Err(); err != nil {
		log.Printf("Error reading .env: %v", err)
	}
}

// menuSelect displays a menu with the provided title and options and returns the selected index.
func menuSelect(title string, options []string) (int, error) {
	prompt := promptui.Select{Label: title, Items: options}
	idx, _, err := prompt.Run()
	return idx, err
}

// clearScreen clears the terminal using ANSI escape codes.
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// This example demonstrates basic project interaction via a simple command loop. It allows adding requirements, importing/exporting to Excel and viewing the project's current state. The program reads GEMINI_API_KEY from the environment and, if present, loads it from a .env file in the working directory.
func main() {
	loadEnv()
	if _, ok := os.LookupEnv("GEMINI_API_KEY"); !ok {
		fmt.Fprintln(os.Stderr, "GEMINI_API_KEY environment variable not set")
		fmt.Fprintln(os.Stderr, "Press any key to continue...")
		_, _ = bufio.NewReader(os.Stdin).ReadByte()
		os.Exit(1)
	}

	path := "./RoelofCompany"
	if err := os.MkdirAll(path, 0o777); err != nil {
		log.Fatal(err)
	}

	if _, err := PMFS.LoadSetup(path); err != nil {
		log.Fatalf("LoadSetup: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		clearScreen()
		listProducts()
		options := []string{"Select product", "Create product", "Exit"}
		idx, err := menuSelect("Product menu", options)
		if err != nil {
			break
		}
		switch idx {
		case 0:
			p := selectProduct(scanner)
			if p != nil {
				productMenu(scanner, p)
			}
		case 1:
			if p := createProduct(scanner); p != nil {
				productMenu(scanner, p)
			}
		case 2:
			fmt.Println("Goodbye!")
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("input error: %v", err)
	}
}
