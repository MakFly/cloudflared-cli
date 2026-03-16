package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm asks the user a yes/no question and returns their response.
func Confirm(message string) bool {
	fmt.Fprintf(out, "%s [y/N]: ", message)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// ConfirmWithDefault asks a yes/no question with a default value.
func ConfirmWithDefault(message string, defaultYes bool) bool {
	prompt := "[y/N]"
	if defaultYes {
		prompt = "[Y/n]"
	}

	fmt.Fprintf(out, "%s %s: ", message, prompt)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response == "" {
		return defaultYes
	}
	return response == "y" || response == "yes"
}

// Prompt asks for text input and returns the response.
func Prompt(message string) string {
	fmt.Fprintf(out, "%s: ", message)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	return strings.TrimSpace(response)
}

// PromptWithDefault asks for text input with a default value.
func PromptWithDefault(message, defaultVal string) string {
	fmt.Fprintf(out, "%s [%s]: ", message, defaultVal)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(response)
	if response == "" {
		return defaultVal
	}
	return response
}
