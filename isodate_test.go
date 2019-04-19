package main

/*
find files in the pattern of name.ext and roll them back to name.num.ext. Any name.num.ext should also be rolled back by one
*/

import (
    "testing"
    "time"
    //"fmt"
)

/******************************************************************************/

func init() {
    test_now = time.Now()
}

func TestNow(t *testing.T) {
    ans := Now()
    expected := test_now
    if ans != expected {
        t.Errorf("no type case did not work %s = %s", ans, expected)
    }
}

func TestExists(t *testing.T) {
    ans := Now().Format(NowByFormat(""))
    expected := test_now.Format(time.RFC3339)
    
    if ans != expected {
        t.Errorf("no type case did not work %s = %s", ans, expected)
    }
}

func TestFormat(t *testing.T) {
    pline(t, "ansic", time.ANSIC, "ancic format")
    pline(t, "unix", time.UnixDate, "unix format")
    pline(t, "ruby", time.RubyDate, "ruby format")
    pline(t, "822", time.RFC822, "822 format")
    pline(t, "822z", time.RFC822Z, "822z format")
    pline(t, "850", time.RFC850, "850 format")
    pline(t, "1123", time.RFC1123, "1123 format")
    pline(t, "1123z", time.RFC1123Z, "1123z format")
    pline(t, "iso", time.RFC3339, "iso format")
    pline(t, "3339", time.RFC3339, "other iso format")
    pline(t, "3339nano", time.RFC3339Nano, "nano iso format")
    pline(t, "kitchen", time.Kitchen, "kitchen format")
    pline(t, "stamp", time.Stamp, "stamp format")
    pline(t, "milli", time.StampMilli, "milli second format")
    pline(t, "micro", time.StampMicro, "micro second format")
    pline(t, "nano", time.StampNano, "nano format")
}

func pline(t *testing.T, a string, b string, msg string) {
    ans := Now().Format(NowByFormat(a))
    expected := test_now.Format(b)    
    
    if ans!=expected {
       t.Errorf(msg + " - %s = %s\n", ans, expected)
    }
}
