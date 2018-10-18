package main

import ("os"
    "fmt"
    "flag"
    "time"
	"math"
    "unsafe"
    "syscall"
	"strconv"
	"strings"
	"bufio"
    )

type winsize struct {
    rows    uint16
    cols    uint16
    xpixels uint16
    ypixels uint16
}

const (
	Level5       = 0x6f
    Level4       = 0x70
    Level3       = 0x71
    Level2       = 0x72
	Level1       = 0x73
)

/** what is the size of the terminal
@param pointer to output stream
@return width, height
*/
func get_term_size(fd uintptr) (int, int) {
    var sz winsize
    _, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
        fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))
    return int(sz.cols), int(sz.rows)
}

/** ESC Commands */
const (
    ESC_SAVE_SCREEN = "?47h"
    ESC_RESTORE_SCREEN = "?47l"
    
    ESC_SAVE_CURSOR = "s"
    ESC_RESTORE_CURSOR = "u"
    
    ESC_BOLD_ON = "1m"
    ESC_BOLD_OFF = "0m"
    
    ESC_CURSOR_ON = "?25h"
    ESC_CURSOR_OFF = "?25l"
    
    ESC_CLEAR_SCREEN = "2J"
    ESC_CLEAR_LINE = "2K"
)

/** some common symbols */
const (
    RuneSterling = '£'
    RuneDArrow   = '↓'
    RuneLArrow   = '←'
    RuneRArrow   = '→'
    RuneUArrow   = '↑'
    RuneBullet   = '·'
    RuneBoard    = '░'
    RuneCkBoard  = '▒'
    RuneDegree   = '°'
    RuneDiamond  = '◆'
    RuneGEqual   = '≥'
    RunePi       = 'π'
    RuneHLine    = '─'
    RuneLantern  = '§'
    RunePlus     = '┼'
    RuneLEqual   = '≤'
    RuneLLCorner = '└'
    RuneLRCorner = '┘'
    RuneNEqual   = '≠'
    RunePlMinus  = '±'
    RuneS1       = '⎺'
    RuneS3       = '⎻'
    RuneS7       = '⎼'
    RuneS9       = '⎽'
    RuneBlock    = '█'
    RuneTTee     = '┬'
    RuneRTee     = '┤'
    RuneLTee     = '├'
    RuneBTee     = '┴'
    RuneULCorner = '┌'
    RuneURCorner = '┐'
    RuneVLine    = '│' //'│'
    RuneUVLine   = '╷'
    RuneDVLine   = '╵'
)

func Print(esc string) {
	fmt.Fprintf(os.Stdout, "%s", esc)
}

func PrintCtrAt(esc string, y, x int) {
    fmt.Fprintf(os.Stdout, "\033[%d;%dH\033[%s", y, x)
}

func PrintAlt(esc string) {
	//printf "\x1b(0\x73\x72\x71\x70\x6f\x1b(B\n"
	//fmt.Fprintf(os.Stdout, "\033(0%s\033(B", esc)
	fmt.Fprintf(os.Stdout, "\x1b(0%s\x1b(B", esc)
}

/**
print out an esc control
@param esc control code to print out
*/
func PrintCtrOnErr(esc string) {
    fmt.Fprintf(os.Stderr, "\033[%s", esc)
}

func PrintCtrOnErrAt(esc string, y, x int) {
    fmt.Fprintf(os.Stderr, "\033[%d;%dH\033[%s", y, x)
}

/**
print a string on the error console at a specific location on the screen
@param msg text to print out
@param y row to print on
@param x col to print one
*/
func PrintStrOnErrAt(msg string, y, x int) {
    fmt.Fprintf(os.Stderr, "\033[%d;%dH%s", y, x, msg)
}

/** Save the screen setup at the start of the app */
func ScrSave() {
    PrintCtrOnErr(ESC_SAVE_SCREEN)
    PrintCtrOnErr(ESC_SAVE_CURSOR)
    PrintCtrOnErr(ESC_CURSOR_OFF)
    PrintCtrOnErr(ESC_CLEAR_SCREEN)
}

/** Restore the screen setup from SrcSave() */
func ScrRestore() {
    PrintCtrOnErr(ESC_CURSOR_ON)
    PrintCtrOnErr(ESC_RESTORE_CURSOR)
    PrintCtrOnErr(ESC_RESTORE_SCREEN)
}

/**
get the current time as a timestamp
@return timestamp
*/
func makeTimestamp() int64 {
    return time.Now().UnixNano() / int64(time.Millisecond)
}

/** sleep for a second */
func WaitSecond() {time.Sleep(1*1000 * time.Millisecond)}

func PrintScale(min float64, num float64, max float64) {
	/*  */
	var ans = math.Round((num*5.0)/(max-min))
	//fmt.Printf("ans:%f num:%f max:%f min:%f\n", ans, num, max-min, min)
	if (ans<=1.0){
		PrintAlt(string(RuneS9))
	} else if (ans<=2.0) {
		PrintAlt(string(RuneS7))
	} else if (ans<=3.0) {
		PrintAlt(string(RuneHLine))
	} else if (ans<=4.0) {
		PrintAlt(string(RuneS3))
	} else if (ans<=5.0) {
		PrintAlt(string(RuneS1))
	}
}

func main() {
    //var up = flag.Int("up", -1, "count up time.")
    //var done = flag.String("done", "", "output when done")

    var ceil *float64 = flag.Float64("ceil", 1.0, "highest possible value, assume 1.0")
	var floor *float64 = flag.Float64("floor", 0.0, "lowest possible value, assume 0.0")
    
	var dynamic = flag.Bool("dynamic", true, "Adjust the floor and ceiling dynamically")

    flag.Parse()
    
    /*if 0==flag.NFlag() || *help {
        flag.PrintDefaults()
    }*/
        
    //ScrSave()
    
	scanner := bufio.NewScanner(os.Stdin)
	var buffer = ""
	
	for scanner.Scan() {
		var raw = scanner.Text()
		var list = strings.Split(raw, " ")
		for _, item := range list {
			f, _ := strconv.ParseFloat(item, 64)
			buffer += item
			if (*dynamic) {
				if f<*floor {
					buffer += string(RuneDArrow)
					Print(string(RuneDArrow))
					*floor = f
				}
				if *ceil<f {
					buffer += string(RuneUArrow)
					Print(string(RuneUArrow))
					*ceil = f
				}
			}
			//Print("\033[0J")
			//Print("\033[1D")
			PrintScale(*floor, f, *ceil)
			
			if (len(buffer)>80) {
				buffer = buffer[:80]
			}
			//Print(buffer)

		}
	}

	//fmt.Printf("\nDone: %f %f \n", *floor, *ceil)
	
}