/* ****************************************************************************
Performs math on date/times
Author: thomas.cherry@gmail.com
Date: 2022-11-20
**************************************************************************** */

package main

import (
    "os"
    "fmt"
    "time"
    "flag"
)

/* ************************************************************************** */
// MARK: - Util functions

/**
Converts a formated string to a time structure
*/
func StringToTime(format string, input string) time.Time {
    if len(format)<1 {
        format = "2006-01-02T15:04:05"
    }
    parseTime, err := time.Parse(format, input)
    if err != nil {
        os.Stderr.WriteString(err.Error() + "\n")
        parseTime, err = time.Parse("2006-01-02", input)
        if err != nil {
            fmt.Println(err)
        }
    }
    
    return parseTime
}

/**
Adds a durration to a time, returning the new time
*/
func Add(working time.Time, num int, unit time.Duration) time.Time {
    if num!=0 {
        working = working.Add(time.Duration(num) * unit)
    }
    return working
}

/* ************************************************************************** */
// MARK: - App functions

func HelpMessageCallback() {
    fmt.Fprintf(flag.CommandLine.Output(),
        "date-math by thomas.cherry@gmail.com\n\n")
    fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n\n", os.Args[0])
    fmt.Fprintf(flag.CommandLine.Output(),
        "\tcmd --older <date> --newer <newer>\n")
    fmt.Fprintf(flag.CommandLine.Output(),
        "\tcmd --date <date> --years <n> --days <n> --hours <n> --minutes <n> --seconds <n>\n")
    fmt.Fprintf(flag.CommandLine.Output(), "\nFlags:\n")
    flag.PrintDefaults()
}

func main() {

	//overwrite the usage function
    flag.Usage = HelpMessageCallback

    //process command line arguments
    format := flag.String("format", "2006-01-02T15:04:05",
        "Golang date/time format")
    older := flag.String("old", "",
        "oldest date/time in 2006-01-02T15:04:05 format")
    newer := flag.String("new", "",
        "newest date/time in 2006-01-02T15:04:05 format")
    date := flag.String("date", "",
        "Date to work on in 2006-01-02T15:04:05 format")
    years := flag.Int("years", 0, "years to add or subtract")
    months := flag.Int("months", 0, "months to add or subtract")
    days := flag.Int("days", 0, "days to add or subtract")
    hours := flag.Int("hours", 0, "hours to add or subtract")
    minutes := flag.Int("minutes", 0, "minutes to add or subtract")
    seconds := flag.Int("seconds", 0, "seconds to add or subtract")

    flag.Parse()

    working := time.Now()
    
    if *date!="" {
        working = StringToTime(*format, *date)
    } else if *newer!="" && *older!=""{
        //find the durration between two date points
        n := StringToTime(*format, *newer)
        o := StringToTime(*format, *older)
        dur := n.Sub(o)
        
        d := int(dur.Hours()/24)
        h := int(dur.Hours() - float64(d*24))
        m := int(dur.Minutes() - float64(d*24*60) - float64(h*60))
        s := int(dur.Seconds() -
            float64(d*24*60*60) -
            float64(h*60*60) -
            float64(m*60))
        fmt.Printf ("%dd %dh %dm %ds\n", d, h, m, s)
        os.Exit(0)
    } else if *newer!="" {
        working = StringToTime(*format, *newer)
    } else if *older!="" {
        working = StringToTime(*format, *older)
    }
    
    orig := working

    if *years!=0 {working = working.AddDate(*years, 0, 0)} 
    if *months!=0 {working = working.AddDate(0, *months, 0)}
    if *days!=0 {working = working.AddDate(0, 0, *days)}
    working = Add(working, *hours, time.Hour)
    working = Add(working, *minutes, time.Minute)
    working = Add(working, *seconds, time.Second)

    fmt.Printf("orig\t%s\nnew\t%s\n", orig.Format(*format),
        working.Format(*format))
}