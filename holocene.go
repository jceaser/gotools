package main

import (
	"fmt"
	"os"
	"strconv"
    "strings"
	"time"
)

func toAstronomical(year int) int {
	// AD years (1 AD = 1, etc.)
	if year >= 1 {
		return year
	}

	// BC years: input year = -N means "N BC"
	// But astronomical year 0 = 1 BC
	// So: N BC -> astronomical year = -(N - 1)
	// Example: -500 -> 500 BC -> astronomical year = -(500 - 1) = -499
	return -(abs(year) - 1)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func work(year int) {
	astYear := toAstronomical(year)
	holocene := astYear + 10000

	fmt.Printf("Input year: %d -> Astronomical year: %d -> Holocene Era: %d HE : column %d\n",
		year, astYear, holocene, holocene/100)

}
func main() {
	var year int
	var err error

	// If year provided on command line:
	if len(os.Args) > 1 {
        parts := strings.Split(os.Args[1], ",")
        for _, part := range parts {
		    year, err = strconv.Atoi(part)
		    if err != nil {
			    fmt.Println("Error: Year must be an integer. Use negative numbers for BC.")
			    return
		    }
            work(year)
        }
	} else {
		// Otherwise use the current year
		year = time.Now().Year()
        work(year)
	}
}
