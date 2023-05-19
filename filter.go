package main

// Filters lines which contain text

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

func matchRegexpBuild(text, match) matchFunc {
	matcher := regexp.MustCompile(match)

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
    match := flag.String("match", "", "text to match in each line")
    reg := flag.Bool("regexp", false, "text to match in each line")
    contains :=  flag.Bool("contains", false, "text to match in each line")
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
