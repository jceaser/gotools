package main

import (
	"bytes"
    "flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"encoding/json"
)

type StringArrayWrapper struct {
	Values []string `json:"-"`
}

// Custom unmarshaler
func (s *StringArrayWrapper) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.Values)
}

func (s StringArrayWrapper) Dump() string {
    return strings.Join(s.Values, ", ")
}

func (s StringArrayWrapper) String() string {
	var cleaned []string
	for _, val := range s.Values {
		// Split by newline and take the first part
		lines := strings.SplitN(val, "\n", 2)
		cleaned = append(cleaned, lines[0])
	}
	return strings.Join(cleaned, ", ")
}

func GetStringArrayWrapper(data []byte) StringArrayWrapper {
    var wrapper StringArrayWrapper
	if err := json.Unmarshal([]byte(data), &wrapper); err != nil {
		panic(err)
	}
	return wrapper
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

/*
    xattr -px com.apple.metadata:_kMDItemUserTags comment.go | \
        xxd -r -p | \
        plutil -convert json -o - -
*/

func getFinderTag(filePath string) (string, error) {

    full_cmd := fmt.Sprintf(
        "/usr/bin/xattr -px com.apple.metadata:_kMDItemUserTags %q | " +
            "/usr/bin/xxd -r -p | " +
            "/usr/bin/plutil -convert json -o - - 2>&1",
        filePath,
    )

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("/bin/sh", "-c", full_cmd)

	cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return "", nil
			}
		}
	}

    output := []byte(stdout.String())

    tags := GetStringArrayWrapper(output)

    comment := fmt.Sprintf("%s", string(tags.String()))
	return comment, nil
}

type AppContext struct {
    Directory *string
    Tag *string
}

func SetupApp() AppContext {
    cxt := AppContext{}

    cxt.Directory = flag.String("path", ".", "Path to scan")
    cxt.Tag = flag.String("tag", "", "Tags to search for")
    flag.Parse()

    return cxt
}

func main() {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	cxt := SetupApp()
	if len(*cxt.Directory)>0 && *cxt.Directory != "." {

    	absPath, err := filepath.Abs(*cxt.Directory)
	    if err != nil {
		    panic(err)
	    }
	    currentDir = absPath
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
			fmt.Printf("%s\t%s\t%s\n", file.Name(), "comment", comment)
		}

		flag, err := getFinderTag(filePath)
		if err != nil {
			fmt.Printf("Tag Error\n")
			continue
		}
		if flag != "" {
			fmt.Printf("%s\t%s\t%s\n", file.Name(), "tag", flag)
		}
	}
}
