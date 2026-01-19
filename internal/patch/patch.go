package patch

import (
	"bufio"
	"os"
	"strings"

	"github.com/bakito/gws/internal/env"
	"github.com/bakito/gws/internal/log"
	"github.com/bakito/gws/internal/types"
)

func Patch(id string, filePatch types.FilePatch) error {
	log.Logf("Patching file %q", id)
	// Read the content of the file
	lines, err := readLines(filePatch.File)
	if err != nil {
		return err
	}

	var processedLines []string
	var changed bool

	if filePatch.OldBlock == "" {
		processedLines, changed = appendToFile(lines, filePatch.NewBlock, filePatch.Indent)
	} else {
		// Process the lines, replacing the block if found
		processedLines, changed = processMultilineBlock(lines, filePatch.OldBlock, filePatch.NewBlock, filePatch.Indent)
	}
	if changed {
		// Write the processed lines to a temporary file
		tempFile := filePatch.File + ".tmp"
		err = writeLines(tempFile, processedLines)
		if err != nil {
			return err
		}

		// Backup the original file
		backupFileName := filePatch.File + ".bak"
		log.Logf("Backup created: %s", backupFileName)
		log.Logf("Original file %q back-upped to %s", id, backupFileName)
		err = backupFile(filePatch.File, backupFileName)
		if err != nil {
			return err
		}

		// Replace the original file with the temporary file
		err = replaceFile(filePatch.File, tempFile)
		if err != nil {
			return err
		}

		log.Logf("Successfully patched %q", id)
	} else {
		log.Logf("No patching required %q", id)
	}
	return nil
}

func appendToFile(lines []string, toAppend, indent string) ([]string, bool) {
	content := strings.Join(lines, "\n")
	changed := false

	toAppend = strings.Join(splitWithIndent(toAppend, indent), "\n")

	if !strings.Contains(content, toAppend) {
		changed = true
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += toAppend
	}
	return strings.Split(content, "\n"), changed
}

// backupFile creates a backup of the original file.
func backupFile(original, backup string) error {
	// Copy the original file to back up
	input, err := os.ReadFile(env.ExpandEnv(original))
	if err != nil {
		return err
	}
	err = os.WriteFile(env.ExpandEnv(backup), input, 0o600)
	if err != nil {
		return err
	}
	return nil
}

// readLines reads a file and returns the lines as a slice of strings.
func readLines(filename string) ([]string, error) {
	file, err := os.Open(env.ExpandEnv(filename))
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

// writeLines writes a slice of strings to a file.
func writeLines(filename string, lines []string) error {
	file, err := os.Create(env.ExpandEnv(filename))
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

// processMultilineBlock processes the lines and replaces the old block with the new one.
func processMultilineBlock(lines []string, oldBlock, newBlock, indent string) ([]string, bool) {
	var result []string
	inBlockIndex := 0
	changed := false

	oldSlice := splitWithIndent(oldBlock, indent)
	newSlice := splitWithIndent(newBlock, indent)

	for _, line := range lines {
		//nolint:gocritic
		if inBlockIndex >= len(oldSlice) {
			result = append(result, newSlice...)
			inBlockIndex = 0
			changed = true
		} else if line == oldSlice[inBlockIndex] {
			inBlockIndex++
			continue
		} else {
			inBlockIndex = 0
		}

		// If not in the block, add the original line to the result
		result = append(result, line)
	}

	if inBlockIndex >= len(oldSlice) {
		result = append(result, newSlice...)
		changed = true
	}

	return result, changed
}

// replaceFile replaces the original file with the temporary file.
func replaceFile(original, tempFile string) error {
	err := os.Rename(env.ExpandEnv(tempFile), env.ExpandEnv(original))
	if err != nil {
		return err
	}
	return nil
}

func splitWithIndent(content, indent string) []string {
	var lines []string

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		lines = append(lines, indent+scanner.Text())
	}

	return lines
}
