package main

/*
find files in the pattern of name.ext and roll them back to name.num.ext. Any name.num.ext should also be rolled back by one
*/

import (
    "testing"
    //"math"
)

/******************************************************************************/

func init() {
}

/**************************************/

func TestExists(*testing.T) {
    //Clear()
    /*result := exists("rpn.go")
    if result != true {
        t.Error("Expected to find rpn.go, got ", result)
    }*/
}

func TestAvergae(t *testing.T) {
    data := []interface{}{1.0,2.0,3.0,10.0}
    ans := Average(data)
    pline(t, 4.0, ans, "Average does not equal %f, got %f")
}
func TestMedian(t *testing.T) {
    data := []interface{}{1.0,2.0,3.0,4.0,5.0,6.0}
    ans := Median(data)
    pline(t, 3.5, ans, "Median does not %f, got %f")
    
    data = []interface{}{1.0,2.0,3.0}
    ans = Median(data)
    pline(t, 2.0, ans, "Median does not equal %f, got %f")
}

func TestMode(t *testing.T) {
    data := []interface{}{0.0,1.0,2.0,3.0,3.0,10.0}
    ans := Mode(data)
    pline(t, 3.0, ans, "Mode does not equal %f")
}

/**************************************/

func pline(t *testing.T, expected float64, ans float64, msg string) {
    if ans!=expected {
       t.Errorf(msg + "\n", expected, ans)
    }
}