package main

/*
find files in the pattern of name.ext and roll them back to name.num.ext. Any name.num.ext should also be rolled back by one
*/

import ("testing")

/******************************************************************************/

func TestVprintf(*testing.T) {
    app_data.verbose = false
    vprintf("")
}

func TestExists(*testing.T) {
    result := exists("roll.go")
    if result != true {
        t.Error("Expected to find roll.go, got ", result)
    }
}