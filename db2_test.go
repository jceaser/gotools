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

/*
table;append 1 2 3 ; table ; append 1.0 ; table ; append 5 4 3 2 1 ; table
*/
func TestAppend(t *testing.T) {
    data := InitDataBase()
    //SetData(data)

    expected := []string {"9","8","7"}
    AppendTable(expected, data)
    row := DataLength(data)-1
    b := interface_to_string(data.Columns["bar"][row])
    f := interface_to_string(data.Columns["foo"][row])
    r := interface_to_string(data.Columns["rab"][row])
    ans := []string{b,f,r}
    check_three(t, expected, ans, "append test - exact")

    source := []string {"8","7","6","5"}
    expected = []string {"8","7","6"}
    AppendTable(source, data)
    row = DataLength(data)-1
    b = interface_to_string(data.Columns["bar"][row])
    f = interface_to_string(data.Columns["foo"][row])
    r = interface_to_string(data.Columns["rab"][row])
    ans = []string{b,f,r}
    check_three(t, expected, ans, "append test - to many given")

    source = []string {"4"}
    expected = []string {"4","0.000000","0.000000"}
    AppendTable(source, data)
    row = DataLength(data)-1
    b = interface_to_string(data.Columns["bar"][row])
    f = interface_to_string(data.Columns["foo"][row])
    r = interface_to_string(data.Columns["rab"][row])
    ans = []string{b,f,r}
    check_three(t, expected, ans,
        "append test - not enough given - %s != expected[%d]=%s")
}

/**************************************/

func length(data map[string][]interface{}) int {
    length := -1
    for _ , v := range data {
        length = len(v)
        break
    }
    return length
}

func check_three(t *testing.T, expected []string, ans []string, msg string) {
    for i,v := range ans {
        if v!=expected[i] {
            t.Errorf(msg, v, i, expected[i])
            break
        }
    }
}

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