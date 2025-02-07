package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	oldSection = `    if platforms.OperatingSystem.IsWindows():
      suite = Suite.PUTTY
      bin_path = _SdkHelperBin()
    else:
      suite = Suite.OPENSSH
      bin_path = None
    return Environment(suite, bin_path)`
	newSection = `    suite = Suite.OPENSSH
    bin_path = None
    return Environment(suite, bin_path)`
)

func main() {
	// Set the path to your ssh.py file
	file := `ssh.py`

	// Backup the original file
	err := backupFile(file, file+".bak")
	if err != nil {
		fmt.Println("Error creating backup:", err)
		return
	}

	// Read the content of the file
	lines, err := readLines(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Process the lines, replacing the block if found
	processedLines := processMultilineBlock(lines, oldSection, newSection)

	// Write the processed lines to a temporary file
	tempFile := file + ".tmp"
	err = writeLines(tempFile, processedLines)
	if err != nil {
		fmt.Println("Error writing to temporary file:", err)
		return
	}

	// Replace the original file with the temporary file
	err = replaceFile(file, tempFile)
	if err != nil {
		fmt.Println("Error replacing original file:", err)
		return
	}

	fmt.Println("Backup created:", backupFile)
	fmt.Println("Replacement complete.")
}

// backupFile creates a backup of the original file
func backupFile(original, backup string) error {
	// Copy the original file to backup
	input, err := ioutil.ReadFile(original)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(backup, input, 0644)
	if err != nil {
		return err
	}
	return nil
}

// readLines reads a file and returns the lines as a slice of strings
func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes a slice of strings to a file
func writeLines(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range lines {
		_, err := file.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

// processMultilineBlock processes the lines and replaces the old block with the new one
func processMultilineBlock(lines []string, oldBlock string, newBlock string) []string {
	var result []string
	inBlock := false

	for _, line := range lines {
		// Check if the start of the block is found
		if line == oldBlockStart {
			inBlock = true
			// Skip the old block and add the new block
			result = append(result, newBlock...)
			continue
		}

		// If inside the block, we check for the end of the block
		if inBlock {
			// When we encounter the end of the block, stop replacing
			if line == oldBlockEnd {
				inBlock = false
			}
			// Skip this line since it's part of the old block
			continue
		}

		// If not in the block, add the original line to the result
		result = append(result, line)
	}

	return result
}

// replaceFile replaces the original file with the temporary file
func replaceFile(original, tempFile string) error {
	err := os.Rename(tempFile, original)
	if err != nil {
		return err
	}
	return nil
}

func lines(multiLineString string) []string {
	var l []string
	scanner := bufio.NewScanner(strings.NewReader(multiLineString))
	for scanner.Scan() {
		l = append(l, scanner.Text())
	}
	return l
}
