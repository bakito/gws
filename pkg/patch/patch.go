package patch

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"

	"github.com/bakito/gws/pkg/types"
)

func Patch(id string, filePatch types.FilePatch) error {
	slog.Info("Patching file", "id", id)
	// Read the content of the file
	lines, err := readLines(filePatch.File)
	if err != nil {
		return err
	}

	// Process the lines, replacing the block if found
	processedLines, changed := processMultilineBlock(lines, filePatch.OldBlock, filePatch.NewBlock)

	if changed {
		// Write the processed lines to a temporary file
		tempFile := filePatch.File + ".tmp"
		err = writeLines(tempFile, processedLines)
		if err != nil {
			return err
		}

		// Backup the original file
		backupFileName := filePatch.File + ".bak"
		fmt.Println("Backup created:", backupFileName)
		slog.Info("Original file back-upped", "id", id, "backup", backupFileName)
		err = backupFile(filePatch.File, backupFileName)
		if err != nil {
			return err
		}

		// Replace the original file with the temporary file
		err = replaceFile(filePatch.File, tempFile)
		if err != nil {
			return err
		}

		slog.Info("Successfully patched", "id", id)
	} else {
		slog.Info("No patching required", "id", id)
	}
	return nil
}

// backupFile creates a backup of the original file
func backupFile(original, backup string) error {
	// Copy the original file to backup
	input, err := os.ReadFile(os.ExpandEnv(original))
	if err != nil {
		return err
	}
	err = os.WriteFile(os.ExpandEnv(backup), input, 0o600)
	if err != nil {
		return err
	}
	return nil
}

// readLines reads a file and returns the lines as a slice of strings
func readLines(filename string) ([]string, error) {
	file, err := os.Open(os.ExpandEnv(filename))
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
	file, err := os.Create(os.ExpandEnv(filename))
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
func processMultilineBlock(lines []string, oldBlock []string, newBlock []string) ([]string, bool) {
	var result []string
	inBlockIndex := 0
	changed := false

	for _, line := range lines {
		if inBlockIndex >= len(oldBlock) {
			result = append(result, newBlock...)
			inBlockIndex = 0
			changed = true
		} else if line == oldBlock[inBlockIndex] {
			inBlockIndex++
			continue
		} else {
			inBlockIndex = 0
		}

		// If not in the block, add the original line to the result
		result = append(result, line)
	}

	if inBlockIndex >= len(oldBlock) {
		result = append(result, newBlock...)
		changed = true
	}

	return result, changed
}

// replaceFile replaces the original file with the temporary file
func replaceFile(original, tempFile string) error {
	err := os.Rename(os.ExpandEnv(tempFile), os.ExpandEnv(original))
	if err != nil {
		return err
	}
	return nil
}
