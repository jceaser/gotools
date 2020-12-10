package main

import ("os"
    "fmt"
    "flag"
    "time"
    "unsafe"
    //"strconv"
    "strings"
    "syscall"
    "os/exec"
    )

type winsize struct {
    rows    uint16
    cols    uint16
    xpixels uint16
    ypixels uint16
}

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
    _ESC_SAVE_SCREEN = "?47h"
    _ESC_RESTORE_SCREEN = "?47l"
    
    _ESC_SAVE_CURSOR = "s"
    _ESC_RESTORE_CURSOR = "u"
    
    _ESC_BOLD_ON = "1m"
    _ESC_BOLD_OFF = "0m"
    
    _ESC_CURSOR_ON = "?25h"
    _ESC_CURSOR_OFF = "?25l"
    
    _ESC_CLEAR_SCREEN = "2J"
    _ESC_CLEAR_LINE = "2K"
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

/** numbers */
var letters = [][][]rune {
    {/*0*/
        {RuneULCorner, RuneHLine, RuneURCorner}, 
        {RuneVLine, ' ', RuneVLine},
        {RuneVLine, ' ', RuneVLine},
        {RuneVLine, ' ', RuneVLine},
        {RuneLLCorner, RuneHLine, RuneLRCorner} },
    {/*1*/
        {' ', RuneUVLine, ' '},
        {' ', RuneVLine, ' '},
        {' ', RuneVLine, ' '},
        {' ', RuneVLine, ' '},
        {' ', RuneDVLine, ' '} },
    {/*2*/
        {RuneHLine, RuneHLine, RuneURCorner},
        {' ', ' ', RuneVLine},
        {RuneULCorner, RuneHLine, RuneLRCorner},
        {RuneVLine, ' ', ' '},
        {RuneLLCorner, RuneHLine, RuneHLine} },
    {/*3*/
        {RuneHLine, RuneHLine, RuneURCorner},
        {' ', ' ', RuneVLine},
        {RuneHLine, RuneHLine, RuneRTee},
        {' ', ' ', RuneVLine},
        {RuneHLine, RuneHLine, RuneLRCorner} },
    {/*4*/
        {RuneTTee, ' ', RuneTTee},
        {RuneVLine, ' ', RuneVLine},
        {RuneLLCorner, RuneHLine, RuneRTee},
        {' ', ' ', RuneVLine},
        {' ', ' ', RuneBTee} },
    {/*5*/
        {RuneULCorner, RuneHLine, RuneHLine},
        {RuneVLine, ' ', ' '},
        {RuneLLCorner, RuneHLine, RuneURCorner},
        {' ', ' ', RuneVLine},
        {RuneHLine, RuneHLine, RuneLRCorner} },
    {/*6*/
        {RuneULCorner, RuneHLine, RuneHLine},
        {RuneVLine, ' ', ' '},
        {RuneLTee, RuneHLine, RuneURCorner},
        {RuneVLine, ' ', RuneVLine},
        {RuneLLCorner, RuneHLine, RuneLRCorner} },
    {/*7*/
        {RuneHLine, RuneHLine, RuneURCorner},
        {' ', ' ', RuneVLine},
        {' ', ' ', RuneVLine},
        {' ', ' ', RuneVLine},
        {' ', ' ', RuneBTee} },
    {/*8*/
        {RuneULCorner, RuneHLine, RuneURCorner},
        {RuneVLine, ' ', RuneVLine},
        {RuneLTee, RuneHLine, RuneRTee},
        {RuneVLine, ' ', RuneVLine},
        {RuneLLCorner, RuneHLine, RuneLRCorner} },
    {/*9*/
        {RuneULCorner, RuneHLine, RuneURCorner},
        {RuneVLine, ' ', RuneVLine},
        {RuneLLCorner, RuneHLine, RuneRTee},
        {' ', ' ', RuneVLine},
        {' ', ' ', RuneBTee} },
    {/*blank*/
        {' ', ' ', ' '},
        {' ', ' ', ' '},
        {' ', ' ', ' '},
        {' ', ' ', ' '},
        {' ', ' ', ' '} } }

/**
print out a large number at a location
@param num 0-11, 11 is a blank
@param y row to print on
@param x col to print on
*/
func PrintNumber(num, y, x int) {
    if (0<=num && 0<=y && 0<=x) {
    for i:= range letters[num] {
        for j:= range letters[num][i] {
            PrintStrOnErrAt(string(letters[num][i][j]), y+i, x+j+1)
        }
    }
    }
}

/**
print out an esc control
@param esc control code to print out
*/
func PrintCtrOnErr(esc string) {
    fmt.Fprintf(os.Stderr, "\033[%s", esc)
}

func PrintCtrOnErrAt(esc string, y, x int) {
    fmt.Fprintf(os.Stderr, "\033[%d;%dH\033[%s", y, x, esc)
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
    PrintCtrOnErr(_ESC_SAVE_SCREEN)
    PrintCtrOnErr(_ESC_SAVE_CURSOR)
    PrintCtrOnErr(_ESC_CURSOR_OFF)
    PrintCtrOnErr(_ESC_CLEAR_SCREEN)
}

/** Restore the screen setup from SrcSave() */
func ScrRestore() {
    PrintCtrOnErr(_ESC_CURSOR_ON)
    PrintCtrOnErr(_ESC_RESTORE_CURSOR)
    PrintCtrOnErr(_ESC_RESTORE_SCREEN)
}

/**
Print the time
@param direction word to show direction
@param i number to print out
*/
func PrintTime(direction string, i int) {
    var x, y = get_term_size(uintptr(syscall.Stdin))
    
    PrintStrOnErrAt("counting " + direction, y/5, x/5)
    
    //PrintCtrOnErrAt(ESC_CLEAR_LINE, y/2-2, x/2-6)
    //PrintCtrOnErrAt(ESC_CLEAR_LINE, y/2-2, x/2-2)
    //PrintCtrOnErrAt(ESC_CLEAR_LINE, y/2-2, x/2+2)
    //PrintCtrOnErrAt(ESC_CLEAR_LINE, y/2-2, x/2+6)
    
    PrintCtrOnErr(_ESC_BOLD_ON)
    fmt.Fprintf(os.Stderr, "\033[%d;%dH", (y/2), (x/2))
    
    PrintNumber( (i/1000)%10, y/2-2, x/2-6)
    PrintNumber( (i/100)%10, y/2-2, x/2-2)
    PrintNumber( (i/10)%10, y/2-2, x/2+2)
    PrintNumber( (i/1)%10, y/2-2, x/2+6)
    
    
    PrintCtrOnErr(_ESC_BOLD_OFF)
    
    PrintStrOnErrAt("", (y), (x-1))
    
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

func DoCommand(command string) {
    s := strings.Split(command, " ")
    cmd := exec.Command(s[0], s[1:]...)
    cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
    err := cmd.Run()
    if err != nil {fmt.Printf("Error: %s\n", err)}
}

func main() {
    var up = flag.Int("up", -1, "Count up time.")
    var down = flag.Int("down", -1, "count down")
    var done = flag.String("done", "", "output when done")
    var cmd = flag.String("command", "", "command to execute")
    var help = flag.Bool("help", false, "Print this help message")
    
    flag.Parse()
    
    if 0==flag.NFlag() || *help {
        flag.PrintDefaults()
    }
    
    //var start = makeTimestamp();
    
    ScrSave()

    PrintTime("down", 1000)
    PrintTime("down", 100)
    PrintTime("down", 10)
    PrintTime("down", 1)
    PrintTime("down", 0)
    PrintTime("down", -1)

    if 0<*up {/*
        for i:=0; i<*up; i++ {
            var now = makeTimestamp();
            PrintTime("up", int(now-start)/1000)
            WaitSecond()
        }*/
    } else if 0<*down {/*
        for i:=*down; 0<=i; i-- {
            var now = makeTimestamp();
            PrintTime("down", *down - int(now-start)/1000)
            WaitSecond()
        }*/
    }
    fmt.Fprintf(os.Stderr, "\r")
    ScrRestore()
    
    if *done!="" {
        fmt.Printf("%s\n", *done)
    }

    if *cmd!="" {
        DoCommand(*cmd)
    }
}