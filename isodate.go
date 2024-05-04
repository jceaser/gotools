package main

/*******************************************************************************
Output the current date and time in a formatted way.
Author: thomas.cherry@gmail.com
*******************************************************************************/

import (
    "time"
    "flag"
    "fmt"
)

/* pulled from documentation
const (
        ANSIC       = "Mon Jan _2 15:04:05 2006"
        UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
        RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
        RFC822      = "02 Jan 06 15:04 MST"
        RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
        RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
        RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
        RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
        RFC3339     = "2006-01-02T15:04:05Z07:00"
        RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
        Kitchen     = "3:04PM"
        // Handy time stamps.
        Stamp      = "Jan _2 15:04:05"
        StampMilli = "Jan _2 15:04:05.000"
        StampMicro = "Jan _2 15:04:05.000000"
        StampNano  = "Jan _2 15:04:05.000000000"
)
*/

var (
    verbose *bool
    test_now time.Time
)

func main() {
    dump := flag.Bool("dump", false, "dump all dates in all formats with names")
    format := flag.String("format", "iso", "defaults to ISO, format to output")
    verbose = flag.Bool("verbose", false, "verbose")
    
    flag.Parse()

    if *dump {
        Dump()
    } else {
        Action(*format)
    }
}

/******************************************************************************/
// #mark - Actions

/**
Primary action, what is run by default. Displays one date string in the
requested format
@param format to display
*/
func Action (format string) {
    selectedFormat := NowByFormat (format)
    vf("%s = ", format)
    fmt.Println(Now().Format(selectedFormat))
}

/**
Additional action, displays all the date formats names and outputs
*/
func Dump() {
    vln("Default format is 'iso' which is an alias for 3339:")
    DumpLine("iso")

    vln("\nDump of all other formats")
    
    DumpLine("1123")
    DumpLine("1123z")
    DumpLine("3339")
    DumpLine("3339nano")
    DumpLine("822")
    DumpLine("822z")
    DumpLine("850")
    DumpLine("ansic")
    DumpLine("kitchen")
    DumpLine("micro")
    DumpLine("milli")
    DumpLine("nano")
    DumpLine("ruby")
    DumpLine("stamp")
    DumpLine("unix")
}

/******************************************************************************/
// #mark - Helpers

/**
writes a "dump" line to the console, writes the format name and the result
@param format date format name to output
*/
func DumpLine(format string) {
    var now = NowByFormat(format)
    fmt.Printf("%9s = %s\n", format, Now().Format(now))
}

/*
gives the current time in the requested format
@param format date format to output
@return the current date formatted in the requested format
*/
func NowByFormat(format string) string{
    var now = "unknown"
    switch (format) {
        case "ansic"        : now = time.ANSIC
        case "unix"         : now = time.UnixDate
        case "ruby"         : now = time.RubyDate
		case "822"          : now = time.RFC822
		case "822z"         : now = time.RFC822Z
		case "850"          : now = time.RFC850
		case "1123"         : now = time.RFC1123
		case "1123z"        : now = time.RFC1123Z
        case "iso", "3339"  : now = time.RFC3339
		case "3339nano"     : now = time.RFC3339Nano
		case "kitchen"      : now = time.Kitchen

		case "stamp"        : now = time.Stamp
		case "milli"        : now = time.StampMilli
		case "micro"        : now = time.StampMicro
		case "nano"         : now = time.StampNano
 
        default:     now = time.RFC3339
    }
    return now
}

func Now() time.Time {
    if test_now.Year() != 1 {
        return test_now
    } else {
        return time.Now()
    }
}

/**
Prints a message line to the console if app is in verbose mode
@param msg message to print out
*/
func vln(msg string) {
    if *verbose {
        fmt.Println(msg)
    }
}

/**
Prints a message using Printf to the console if app is in verbose mode
@param format message format - see Printf()
@param args arguments for the massage format, variable list - see Printf()
*/
func vf(format string, args ...interface{}) {
    if *verbose {
        fmt.Printf(format, args...)
    }
}
