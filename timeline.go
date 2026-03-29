/* **********************************************************************************************100

Example JSON input:
{
    "Title": "My Timeline",
    "Nodes": [
        {"Start": -5000, "End": -4000, "Label": "Ancient Era"},
        {"Start": -3000, "End": -2000, "Label": "Bronze Age"},
        {"Start": 0, "End": 2024, "Label": "Common Era"},
        {"Now": true, "Mark": true, "Label": "Today"}
    ]
}

Each node can specify a start and end year, or use "Now" to indicate the current year.
If "Mark" is true, it will print a single point instead of a range.

Compile and run with:
cat timeline.json | go run timeline.go

format: gofmt -d timeline.go
*/

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// ***************************************************************************80
// MARK: Structures

const (
	RuneCkBoard  = '▒'
	RuneUnderline = '─'
	RuneVertical  = '│'
	RuneIntersection = '┼'
	RuneDiamond   = '◆'
	RuneLArrow	= '←'
)

type TimeLine struct {
	Title string
	Nodes []TimeLineNode
}

type TimeLineNode struct {
	Start int
	Now   bool
	Mark  bool
	End   int
	Label string
}

type win_size struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// ***************************************************************************80
// MARK: Functions

func GetWidth() int {
	ws := &win_size{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))
	if int(retCode) == -1 {
		panic(errno)
	}
	return int(ws.Col)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

/** limit string to width */
func limit(str string, width int) string {
	if len(str) <= width {
		return str
	}
	return str[:width]
}

/** print string limited to width */
func print(str string, width int) {
	fmt.Printf("%s", limit(str, width))
}

// ***********************************40
// MARK: JSON Helpers

func JsonToStruct[Target any](bytes []byte) (Target, error) {
	var data Target
	err := json.Unmarshal(bytes, &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func StructToJson[T any](data T, useIndent bool) ([]byte, error) {
	var bytes []byte
	var err error
	if useIndent {
		bytes, err = json.MarshalIndent(data, "", strings.Repeat(" ", 4))
	} else {
		bytes, err = json.Marshal(data)
	}
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// ***************************************************************************80
// MARK: application functions

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

func toHolocene(year int) int {
	astYear := toAstronomical(year)
	return astYear + 10000
}

func printHeader() {
	max_width := (GetWidth() - 7) * 100
	for i := 0; i < max_width; i++ {
		mark := i % 1000
		if mark != 0 {
			continue
		}
		fmt.Printf("│%-9d", i) // 1+8 is 9 chars wide
	}
	fmt.Println()
	for i := 0; i < (max_width / 100); i++ {
		mark := i % 10
		if mark != 0 {
			fmt.Printf(string(RuneUnderline))
		} else {
			fmt.Printf(string(RuneIntersection))
		}
	}
	fmt.Println()
}

func processTimeLineRow(node TimeLineNode) int {
	width := GetWidth()
	scale := 100 //number of years per column

	if node.Now {
		node.Start = time.Now().Year()
		if node.Label == "" {
			node.Label = "Today"
		}
	}

	if node.Mark {
		node.End = node.Start
	}

	astYear := toAstronomical(node.Start)
	holocene := astYear + 10000

	astEnd := toAstronomical(node.End)
	holoceneEnd := astEnd + 10000

	durration := holoceneEnd - holocene

	// calculate positions, maybe round in the future
	mark1 := max(0, holocene/scale)
	mark2 := max(0, durration/scale)
	spacer := strings.Repeat(" ", mark1)
	span := strings.Repeat(string(RuneCkBoard), mark2)
	if node.Mark {
		print(fmt.Sprintf("%s◆ <-- %d : %s", spacer, holocene, node.Label), width)
	} else {
		print(fmt.Sprintf("%s│%s│ <-- %d to %d (%d years): %s",
			spacer, span, holocene, holoceneEnd, durration, node.Label), width)
	}
	fmt.Println()
	return durration
}

// If stdin is from a pipe or file:
func hasInput() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	return (info.Mode() & os.ModeCharDevice) == 0
}

func doTimeLine(timeline TimeLine) {
	fmt.Printf("%s\n", timeline.Title)
	printHeader()

	total := 0
	for _, seg := range timeline.Nodes {
		total += processTimeLineRow(seg)
	}

	fmt.Printf("Average durration: %d years.\n", total/len(timeline.Nodes))
}

// ***************************************************************************80
// MARK: main

func main() {

	if !hasInput() {
		fmt.Println("Please provide timeline JSON data via stdin.")
		return
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
        fmt.Println("Error reading stdin:", err)
        return
    }

	timeline, err := JsonToStruct[TimeLine]([]byte(data))
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	doTimeLine(timeline)
}
