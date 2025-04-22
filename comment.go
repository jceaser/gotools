package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// Read all files in the current directory
	files, err := os.ReadDir(currentDir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	// Iterate through each file
	for _, file := range files {
		filePath := filepath.Join(currentDir, file.Name())
		comment, err := getFinderComment(filePath)
		if err != nil {
			fmt.Printf("Error getting comment for %s: %v\n", file.Name(), err)
			continue
		}

		if comment != "" {
			fmt.Printf("%s: %s\n", file.Name(), comment)
		}

		flag, err := getFinderTag(filePath)
		if err != nil {
			fmt.Printf("Error\n")
			continue
		}
		if flag != "" {
			fmt.Printf("%s: %s\n", file.Name(), flag)
		}

	}
}

func getFinderComment(filePath string) (string, error) {
	cmd := exec.Command("xattr", "-p", "com.apple.metadata:kMDItemFinderComment", filePath)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// xattr returns exit status 1 if the attribute doesn't exist
			if exitError.ExitCode() == 1 {
				return "", nil
			}
		}
		return "", err
	}

	//bplist00
	prefix := []byte{0x62, 0x70, 0x6C, 0x69, 0x73, 0x74, 0x30, 0x30}
	if bytes.HasPrefix(output, prefix) {
		output = output[11:]
	}

	// Trim any whitespace and newlines from the output
	comment := strings.TrimSpace(string(output))
	return comment, nil
}

func getFinderTag(filePath string) (string, error) {
	cmd := exec.Command("xattr", "-p", "com.apple.metadata:_kMDItemUserTags", filePath)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// xattr returns exit status 1 if the attribute doesn't exist
			if exitError.ExitCode() == 1 {
				return "", nil
			}
		}
		return "", err
	}

	// Trim any whitespace and newlines from the output
	comment := strings.TrimSpace(string(output))
	return comment, nil
}
