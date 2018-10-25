package main

import ("os"
    "bufio"
    "bytes"
    "flag"
    "fmt"
    "math"
    "strconv"
    "strings"
    "os/signal"
    "syscall"
    "time"
    "unsafe"
    )

/****/
type winsize struct {
    Row    uint16
    Col    uint16
    Xpixel uint16
    Ypixel uint16
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
    return int(sz.Col), int(sz.Row)
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
    
    ESC_HOME = "H"
    ESC_CLEAR_DOWN = "J"
    ESC_CLEAR_EOL = "K"
    ESC_CLEAR_SCREEN = "2J"
    ESC_CLEAR_LINE = "2K"
    
    ESC_BOLD = "1m"
    ESC_UNDER_LINE = "4m"
    ESC_BLINK = "5m"
    ESC_REVERSE = "7m"
    ESC_END_FORMAT = "0m"
    
    ESC_BACK_X = "D"
    ESC_START_LINE = "\r"
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
    PrintAlt(Scale(min, num, max))
}

func Scale(min float64, num float64, max float64) string {
    /*  */
    var ans = math.Round((num*5.0)/(max-min))
    var out = ""
    //fmt.Printf("ans:%f num:%f max:%f min:%f\n", ans, num, max-min, min)
    if (ans<=1.0){
        out = string(RuneS9)
    } else if (ans<=2.0) {
        out = string(RuneS7)
    } else if (ans<=3.0) {
        out = string(RuneHLine)
    } else if (ans<=4.0) {
        out = string(RuneS3)
    } else if (ans<=5.0) {
        out = string(RuneS1)
    }
    return out
}

func getWidth() uint {
    ws := &winsize{}
    retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
        uintptr(syscall.Stdin),
        uintptr(syscall.TIOCGWINSZ),
        uintptr(unsafe.Pointer(ws)))

    if int(retCode) == -1 {
        panic(errno)
    }
    return uint(ws.Col)
}

func main() {
    
    //trap control - c
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Printf("\033[2K\r") //clear line and move back to beginning
        fmt.Printf("\n\033[2K\r") //clear line and move back to beginning
        fmt.Printf("\n\033[2K\r") //clear line and move back to beginning

        os.Exit(0)
    }()

    
    var ceil *float64 = flag.Float64("ceil", 1.0, "highest possible value, assume 1.0")
    var floor *float64 = flag.Float64("floor", 0.0, "lowest possible value, assume 0.0")
    var dynamic = flag.Bool("dynamic", true, "Adjust the floor and ceiling dynamically")
    var dump = *flag.Bool("dump", false, "Don't use terminal esc sequences")
    var wait = flag.Int("wait", 2000, "delay between reads in milliseconds")
    
    flag.Parse()
    
    /*if 0==flag.NFlag() || *help {
        flag.PrintDefaults()
    }*/
        
    //ScrSave()
    
    scanner := bufio.NewScanner(os.Stdin)
    var buf bytes.Buffer
    var avg_sum float64 = 0.0
    var avg_count int = 0
    var avg float64 = 0.0

    for scanner.Scan() {
        var raw = scanner.Text()
        var list = strings.Split(raw, " ")
        
        for _, item := range list {
            f, _ := strconv.ParseFloat(item, 64)
            avg_sum += f
            avg_count++
            avg = avg_sum / float64(avg_count)
            if (*dynamic) {
                if f<*floor {
                    //buf.WriteString("\\033[31m")
                    buf.WriteRune(RuneDArrow)
                    //buf.WriteString("\\033[0m")
                    *floor = f
                }
                if *ceil<f {
                    buf.WriteRune(RuneUArrow)
                    *ceil = f
                }
            }
            buf.WriteString(Scale(*floor, f, *ceil))
            
            //limit buffer to this size
            var width, _ = get_term_size(uintptr(syscall.Stdout))
            var buf_count = strings.Count(buf.String(), "") + 0
            if (buf_count>=width) {
                var hold_it string = buf.String()[buf_count-width:]
                buf.Reset()
                buf.WriteString(hold_it)
            }
            
            /******************************************************/
            
            //print first line
            if (!dump) {
                fmt.Printf("\033[2K\r") //clear line and move back to beginning
            }
            fmt.Printf(buf.String())
            
            //print second line
            if (!dump) {
                fmt.Printf("\n\033[2K\r")
                fmt.Printf("floor: %f, avg: %f, ceil: %f", *floor, avg, *ceil)
            }

            //print third line
            if (!dump) {
                fmt.Printf("\n\033[2K\r")
                fmt.Printf("w=%d, c=%d", buf_count, avg_count)
            }
            
            //reset back to first line
            if (!dump) {
                fmt.Printf("\n\033[3A")
                time.Sleep( time.Duration(*wait) * time.Millisecond)
            }
        }
    }
    if (dump) {
        fmt.Printf("\nfloor: %f, avg: %f, ceil: %f", *floor, avg, *ceil)
    } else {
        fmt.Printf("\n\n\n")
    }
}