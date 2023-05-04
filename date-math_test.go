/*******************************************************************************
Tests for date-math.go
*******************************************************************************/

package main

import (
    "testing"
    "time"
)

/******************************************************************************/
// MARK: - Constants

const (
    ISO = "2006-01-02T15:04:05"
)

/******************************************************************************/
// MARK: - Functions

func init() {
}

/******************************************************************************/
// MARK: - Tests

func TestStringToTime(t *testing.T) {
    tester := func(format, input, msg string) {
        result := StringToTime(format, input)
        
        if format=="" {
            format = ISO
        }
        expected, err := time.Parse(format, input)
        if err != nil {
            t.Error("Expected no error, got ", err)
        }
        if result != expected {
            t.Error(msg, "Expected date to match, got ", result)
        }
    }
    
    tester("2006-01-02",            "2022-11-11",           "Short Date")
    tester(ISO,                     "2022-11-11T02:04:05",  "ISO Date")
    tester("2006-01-02 15:04:05",   "2022-11-11 02:04:05",  "Human Date")
    tester("",                      "2022-11-11T02:04:05",  "Assume format")
}

func TestTime(t *testing.T) {
    tester := func (given time.Time,
            num int,
            unit time.Duration,
            expected_str string,
            msg string) {
        result := Add(given, num, unit)
        expected, err := time.Parse(ISO, expected_str)
        if err != nil {
            t.Error (msg, ": error parsing expected string", expected_str)
        }
        if result != expected {
            t.Error(msg, "result [", result, "] != [", expected, "]")
        }
    }
    base, _ := time.Parse(ISO, "2022-01-02T08:10:15")
    
    tester(base, 0, time.Hour, "2022-01-02T08:10:15", "No change")
    tester(base, 5, time.Hour, "2022-01-02T13:10:15", "Five Hours Ahead")
    tester(base, 3, time.Minute, "2022-01-02T08:13:15", "Three Minutes Ahead")
    tester(base, 1, time.Second, "2022-01-02T08:10:16", "One Second Ahead")
}
