package main

import (
	"bufio"
	"fmt"
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
	input, err := os.ReadFile(original)
	if err != nil {
		return err
	}
	err = os.WriteFile(backup, input, 0644)
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
	oldBlockLines := splitLines(oldBlock)

	var result []string
	inBlockIndex := 0

	for _, line := range lines {
		if inBlockIndex >= len(oldBlockLines) {
			result = append(result, splitLines(newBlock)...)
			inBlockIndex = 0
		} else if line == oldBlockLines[inBlockIndex] {
			inBlockIndex++
			continue
		} else {
			inBlockIndex = 0
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

func splitLines(multiLineString string) []string {
	var l []string
	scanner := bufio.NewScanner(strings.NewReader(multiLineString))
	for scanner.Scan() {
		l = append(l, scanner.Text())
	}
	return l
}
