package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"
)

func main() {
	varsDefault, err := readPythonFile(
		"config.py",
		"CLOUDSDK_CLIENT_ID",
		"CLOUDSDK_CLIENT_NOTSOSECRET",
		"CLOUDSDK_SCOPES",
	)
	if err != nil {
		return
	}
	tmpl, err := template.ParseFiles("auth_config.go.tpl")
	if err != nil {
		_, _ = fmt.Println("Error parsing template file:", err)
		return
	}
	err = tmpl.Execute(os.Stdout, map[string]any{
		"DefaultClientID":     varsDefault["CLOUDSDK_CLIENT_ID"][0],
		"DefaultClientSecret": varsDefault["CLOUDSDK_CLIENT_NOTSOSECRET"][0],
		"DefaultClientScopes": varsDefault["CLOUDSDK_SCOPES"],
	})
	if err != nil {
		_, _ = fmt.Println("Error processing template file:", err)
	}
}

func readPythonFile(name string, keys ...string) (map[string][]string, error) {
	// Open the file
	file, err := os.Open(name) // Replace with your actual file name
	if err != nil {
		_, _ = fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	// Define keys to extract
	targetKeys := make(map[string]bool)
	for _, key := range keys {
		targetKeys[key] = true
	}

	// Regex to match key-value pairs
	re := regexp.MustCompile(`(?m)^(\w+)\s*=\s*(.*)$`)

	// Store extracted variables
	vars := make(map[string][]string)

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var multiLineKey string
	var multiLineValues []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignore empty lines and full-line comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Remove inline comments (anything after #)
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		// Handle multi-line tuples (CLOUDSDK_SCOPES)
		if strings.HasSuffix(line, "(") {
			multiLineKey = strings.Split(line, "=")[0]
			multiLineKey = strings.TrimSpace(multiLineKey)
			if targetKeys[multiLineKey] {
				multiLineValues = []string{}
			} else {
				multiLineKey = ""
			}
			continue
		}
		if multiLineKey != "" {
			// End of multi-line tuple
			if strings.HasSuffix(line, ")") {
				line = cleanValue(strings.TrimSuffix(line, ")"))
				if line != "" {
					multiLineValues = append(multiLineValues, line)
				}
				vars[multiLineKey] = multiLineValues
				multiLineKey = ""
				continue
			}
			line = cleanValue(line)
			if line != "" {
				multiLineValues = append(multiLineValues, line)
			}
			continue
		}

		// Match single-line key-value pairs
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			if targetKeys[key] {
				value := cleanValue(matches[2])
				vars[key] = []string{value}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		_, _ = fmt.Println("Error reading file:", err)
		return nil, err
	}
	return vars, nil
}

// cleanValue removes unwanted characters (' and ,).
func cleanValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "'", "") // Remove single quotes
	value = strings.ReplaceAll(value, ",", "") // Remove commas
	return value
}
