package main

/*
find files in the pattern of name.ext and roll them back to name.num.ext. Any name.num.ext should also be rolled back by one
*/

import (
    "testing"
    "math"
)

/******************************************************************************/

/*func TestVprintf(*testing.T) {
    app_data.verbose = false
    vprintf("")
}*/

func TestExists(*testing.T) {
    /*result := exists("rpn.go")
    if result != true {
        t.Error("Expected to find rpn.go, got ", result)
    }*/
}

func TestPush(t *testing.T) {
    Push(3.14)
    if stack[0] != 3.14 {
        t.Errorf("Push did not work")
    }
}

func TestPop(t *testing.T) {
    Push(3.14)
    ans := Pop()
    if ans != 3.14 {
        t.Errorf("Push did not work")
    }
}

func TestMed(t *testing.T) {
    //odd use case
    Push(3) ; Push(5) ; Push(9) ; Push(8) ; Push(6)
    Push(58) ; Push(9) ; Push(4) ; Push(10)
    Median()
    ans := Pop()
    if ans != 7.0 {
        t.Errorf("Off number of item use case did not work")
   }

    //even case
    Push(3) ; Push(5) ; Push(9) ; Push(8) ; Push(6)
    Push(58) ; Push(9) ; Push(4)
    Median()
    ans = Pop()
    if ans != 8.0 {
        t.Errorf("Even number of item use case did not work: %f!=8.0", ans)
   }
}

func TestStandardDeviation(t *testing.T) {
    //even case
    Push(3) ; Push(5) ; Push(9) ; Push(1) ; Push(8) ; Push(6) ; Push(58)
    Push(9) ; Push(4) ; Push(10)
    StandardDeviation()
    ans := Pop()
    expected := 15.8117045254457
    if ans != expected {
        t.Errorf("Standard Deviation did not work: %f!=%f", ans, expected)
   }
}

func TestFact(t *testing.T) {
    Push(5.9)
    Factorial()
    ans := Pop()
    expected := 120.0
    if ans != expected {
       t.Errorf("Factorial was incorrect, got: %f, want: %f.", ans, expected)
    }
}

func TestSum(t *testing.T) {
    Push(2.0)
    Push(3.0)
    Plus()
    ans := Pop()
    if ans!=5.0 {
       t.Errorf("Sum was incorrect, got: %f, want: %f.", ans, 5.0)
    }

    /*total := Sum(5, 5)
    if total != 10 {
       t.Errorf("Sum was incorrect, got: %d, want: %d.", total, 10)
    }*/
}

func TestProcessLine(t *testing.T) {
    InitializeActions()
    pline(t, "6 2 +", 8.0, "Simple addition %f==%f")
    pline(t, "6 2 -", 4.0, "Simple subtraction %f==%f")
    pline(t, "6 2 *", 12.0, "Simple multiplication %f==%f")
    pline(t, "6 2 /", 3.0, "Simple division %f==%f")
    pline(t, "2 6 ^", 64.0, "alt division %f==%f")
    pline(t, "6 2 %", 0.0, "alt division %f==%f")
    pline(t, "2 6 min", 2.0, "find min %f==%f")
    pline(t, "2 6 max", 6.0, "find max %f==%f")
    pline(t, "2 6 <>", 2.0, "swap %f==%f")

    pline(t, "2 6 drop", 2.0, "drop value %f==%f")
    pline(t, "6 --", 5.0, "decrement %f==%f")
    pline(t, "6 ++", 7.0, "increment %f==%f")
    pline(t, "3 ^2", 9.0, "square %f==%f")
    pline(t, "3.14 integer", 3.0, "integer part of the number %f==%f")
    //pline(t, "2.6 decimal", 0.6, "decimal part of the number %f==%f");

    pline(t, "1 2 3 >>", 2.0, "rotate right %f==%f")
    pline(t, "1 2 3 <<", 1.0, "rotate left %f==%f")
    pline(t, "3 2 1 sort", 3.0, "sort %f==%f")
    //pline(t, "1 2 3 clear", math.NaN(), "Empty the stack %f!=%f")
}

func TestProcessLine_Bad(t *testing.T) {
    ProcessLine("+", false)
    expected := math.NaN()
    ans := Pop()
    Print()
    if !math.IsNaN(ans) {
       t.Errorf("no values %f==%f\n", ans, expected)
    }

}

func pline(t *testing.T, formula string, expected float64, msg string) {
    ProcessLine(formula, false)
    ans := Pop()
    Clear()
    if ans!=expected {
       t.Errorf(msg + "\n", ans, expected)
    }
}