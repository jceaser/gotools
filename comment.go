package main

/***************************************************************************************************
follow up the command with a call to column -t -s $'\t'
***************************************************************************************************/

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

var Author = "thomas.cherry@gmail.com"

/************************************************/
// MARK: Structs

// Application context, setup for the current instance
type AppContext struct {
    Directory *string
    Tag *string
}

// Shorthand for reading in a json array of strings.
type StringArrayWrapper struct {
	Values []string `json:"-"`
}

// Custom unmarshaler for the json string array
func (s *StringArrayWrapper) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.Values)
}

// Simple function to dump out the values of the array
func (s StringArrayWrapper) Dump() string {
    return strings.Join(s.Values, ", ")
}

// Pretty print out the values of the array
func (s StringArrayWrapper) String() string {
	var cleaned []string
	for _, val := range s.Values {
		// Split by newline and take the first part
		lines := strings.SplitN(val, "\n", 2)
		cleaned = append(cleaned, lines[0])
	}
	return strings.Join(cleaned, ", ")
}

/************************************************/
// MARK: - Functions

// Read in an array of raw bytes and turn them into a StringArrayWrapper
func GetStringArrayWrapper(data []byte) StringArrayWrapper {
    var wrapper StringArrayWrapper
	if err := json.Unmarshal([]byte(data), &wrapper); err != nil {
		panic(err)
	}
	return wrapper
}

// Return the comment from a given file. If no comment exists, return empty string
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

// Return a string with the list of tags for a given file. Return empty string of none are found
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

/**************************************************************************************************/
// MARK: - Application

func SetupApp() AppContext {
    cxt := AppContext{}

    flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Comment by %s:\n", Author)
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nExample:")
		fmt.Fprintf(os.Stderr, "%s -tag green | column -t -s $'\t'", os.Args[0])
	}

    cxt.Directory = flag.String("path", ".", "Path to scan")
    cxt.Tag = flag.String("tag", "", "Tags to search for")
    flag.Parse()

    return cxt
}

func main() {
	cxt := SetupApp()
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

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

	fmt.Printf("%1s\t%1s\t%1s\n", "File", "Tags", "Comment")

	// Iterate through each file
	for _, file := range files {
		filePath := filepath.Join(currentDir, file.Name())

        //get tags
		tags, err := getFinderTag(filePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Tag Error\n")
			tags = ""
		}

        //get comment
		comment, err := getFinderComment(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting comment for %s: %v\n", file.Name(), err)
			comment = ""
		}

		// nothing to see, move on
		if len(comment)<1 && len(tags)<1 {
		    continue
		}
        if  *cxt.Tag=="" || strings.Contains(strings.ToLower(tags), strings.ToLower(*cxt.Tag)) {
			fmt.Printf("%1s\t%1s\t%1s\n", file.Name(), tags, comment)
		}
	}
}
