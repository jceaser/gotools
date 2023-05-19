package main

/*
find files in the pattern of name.ext and roll them back to name.num.ext. Any name.num.ext should also be rolled back by one
*/

import (
    "bytes"
    "strings"
    "testing"
    //"time"
)

/******************************************************************************/

func init() {
}

func TestNow(t *testing.T) {
    td := TemplateData{
        Title: "Page Title",
        SafeTitle: "page_title",
        Content: "<b>content</b>",
        Date: "2023-05-10 07:47 PM",
        Path: "some-file-here"}

    if td.Exists("null") {
        t.Errorf("File should not exist")
    }

    td.Path = "/dev"
    if !td.Exists("null") {
        t.Errorf("File does not exist")
    }

    if len(td.Random("fish")) > 0 {
        t.Errorf("value when there should not be")
    }
}

func TestBurnList(t *testing.T) {
    burn := BurnList{}

    //empty tests
    if !burn.Available(42) {
        t.Errorf("42 should be available")
    }

    if burn.Burned(42) {
        t.Errorf("no number should be burned yet")
    }


    for i := 0 ; i < 9 ; i++ {
        burn.Burn(i)
    }

    if !burn.Available(9) {
        t.Errorf("last one not available: %v", burn)
    }

    if len(burn) != 9 {
        t.Errorf("9 items, 0-8 should be burned: %v", burn)
    }

    actual := burn.NotRepeated(10)
    if 9 != actual {
        t.Errorf("last number not available %d %v.", actual, burn)
    }

    if len(burn) != 10 {
        t.Errorf("10 items, 0-9 should be burned: %v", burn)
    }

    actual = burn.NotRepeated(10)
    if actual != -1 {
        t.Errorf("9 items already burned: %v", burn)
    }
}


func TestRender(t *testing.T) {
    data := TemplateData{
        Title: "Page Title",
        SafeTitle: "page_title",
        Content: "<b>content</b>",
        Date: "2023-05-10 07:47 PM",
        Path: "some-file-here"}

    test := func(msg, input, expected, logMsg string) bool {
        var output bytes.Buffer
        Log.Error.SetOutput(&output)
        Log.Warn.SetOutput(&output)

        actual := Render(input, data)
        if expected != actual {
            t.Errorf("%s: '%s' vs '%s'", msg, expected, actual)
        }
        if !strings.Contains(output.String(), logMsg) {
            t.Errorf("%s: '%s' vs '%s'", msg, logMsg, output.String())
        }
        return true
    }
    test("Title", "* {{.Title}}", "* Page Title", "")
    test("Safe Title", "* {{.SafeTitle}}", "* page_title", "")
    test("Content", "`{{.Content}}`", "`<b>content</b>`", "")
    test("Date", "{{.Date}}", "2023-05-10 07:47 PM", "")
    test("Error", "{{.Fake}}", "",
        "can't evaluate field Fake in type main.TemplateData")
}
