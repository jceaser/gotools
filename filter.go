package main

// Filters out lines from standard in which contain text or match a pattern

import (
    "fmt"
    "bufio"
    "flag"
    "io"
    "os"
    "regexp"
    "strings"
    )

type matchFunc func(string, string) bool

func matchString(text, match string) bool {
	return strings.Contains(text, match)
}

func matchRegexp(text, match string) bool {
	matcher := regexp.MustCompile(match)
	found := matcher.Match([]byte(text))
	return found
}

func filter (input io.Reader, output *io.PipeWriter, match string,
        matcher matchFunc) {
	scanner := bufio.NewScanner(input)
	defer output.Close()
	for scanner.Scan() {
		line := scanner.Text()
		if len(match)>0 && matcher(line, match) {
			continue
		}
		output.Write([]byte(line + "\n"))
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func main() {

    flag.Usage = func() {
        fmt.Printf("Filter out lines of text from standard in\n")
        fmt.Printf("By thomas.cherry@gmail.com\n\n")
        raw_app_name := os.Args[0]
        index_of_slash := strings.LastIndex(os.Args[0], "/") + 1
        app_name := [index_of_slash:]
	    fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", app_name)
	    flag.PrintDefaults()
    }

    contains :=  flag.Bool("contains", false, "match is raw text")
    match := flag.String("match", "", "text to match in each line")
    reg := flag.Bool("regexp", false, "match is regex")
    flag.Parse()

	var matcher matchFunc
	if *reg {
	    matcher = matchRegexp
	} else if *contains {
	    matcher = matchString
	} else {
	    return
	}

	piper, pipew := io.Pipe()
	go filter(os.Stdin, pipew, *match, matcher)
	io.Copy(os.Stdout, piper)
}
