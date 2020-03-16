package main

/*
find files in the pattern of name.ext and roll them back to name.num.ext. Any name.num.ext should also be rolled back by one
*/

import (
    "testing"
    //"math"
    //"fmt"
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

func TestConvert(t *testing.T) {
    //var data interface{}
    
    var f64 float64
    f64 = 3.145920
    ans := interface_to_string(f64)
    sline(t, "3.145920", ans, "Float64 to string does not equal %s from %s.")
    
    var f32 float32
    f32 = 3.145920
    ans = interface_to_string(f32)
    sline(t, "3.145920", ans, "Float32 to string does not equal %s from %s.")
    
    ans = interface_to_string(3.145920)
    sline(t, "3.145920", ans, "Float to string does not equal %s from %s.")
    
    ans = interface_to_string(1)
    sline(t, "1", ans, "Int to string does not equal %s from '%s'.")

    ans = interface_to_string("3.145920")
    sline(t, "3.145920", ans, "String to string does not equal %s from %s.")
}

func TestFloatConvert(t *testing.T) {
    ans := interface_to_float(3.14592)
    pline(t, 3.145920, ans, "Raw to float does not equal %f from %f.")

    ans = interface_to_float("3.14592")
    pline(t, 0.0, ans, "String test %f from %f.")

}

/**************************************/

func sline(t *testing.T, expected string, ans string, msg string) {
    if ans!=expected {
       t.Errorf(msg + "\n", expected, ans)
    }
}

func pline(t *testing.T, expected float64, ans float64, msg string) {
    if ans!=expected {
       t.Errorf(msg + "\n", expected, ans)
    }
}