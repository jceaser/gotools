package main

import ("fmt"
    //"bufio"
    "os"
    /*"io"
    "bytes"*/
    "flag"
    "time"
    "math/rand"
    "unsafe"
    "syscall"
    "os/signal"
    "strconv"
    )

/******************************************************************************/
// MARK: - Consts and Structures

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

/****/
type winsize struct {
    Row    uint16
    Col    uint16
    Xpixel uint16
    Ypixel uint16
}

/******************************************************************************/
// MARK: - Terminal Functions

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

func getHeight() uint {
    ws := &winsize{}
    retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
        uintptr(syscall.Stdin),
        uintptr(syscall.TIOCGWINSZ),
        uintptr(unsafe.Pointer(ws)))

    if int(retCode) == -1 {
        panic(errno)
    }
    return uint(ws.Row)
}

func TerminalCommand(command string) {
    fmt.Printf("\033[%s", command)
}

func PrintAt(x, y int, text string) {
    fmt.Printf("\033[%d;%dH%s", x, y, text)
}

func FormatAt(x,y int, text string) string {
    return fmt.Sprintf("\033[%d;%dH%s", x, y, text)
}

func CommandAdd(command, text string) string {
    return fmt.Sprintf("\033[%s%s", command, text)
}

func CommandWrap(begin, text, end string) string {
    return fmt.Sprintf("\033[%s%s\033[%s", begin, text, end)
}

func ColorText(text string, color int) string {
    encoded := fmt.Sprintf("\033[0;%dm%s\033[0m", color, text)
    return encoded
}

/******************************************************************************/
// - MARK: Functions

/** create a string containing a number on the base 36 scale */
func letter(n int64) string {
    return strconv.FormatInt(n, 36)
}

func run(n int, c chan string) {
    loopcount := 1
    for {
        r := rand.Intn(42)//0<letters<36<spaces<50
        
        r = rand.Intn(128-28)+36
        
        rs := " "
        //if r <= 35 {
        if 32 < r && r < 128{
            //rs = letter(int64(r))
            rs = string(r)
            c <- rs
        } else {
            //add some spaces
            space_space := rand.Intn(16) + 16
            for sp:=0; sp<space_space; sp++ {
                c <- " "
            }
        }
        loopcount += 1
    }
}

func PaintScreen (screen [][]string) {
    for row:=0 ; row<int(getHeight()); row++ {
        for col:=0 ; col<int(getWidth()); col++ {
            msg := FormatAt(row, col, screen[row][col])
            
            r := rand.Intn(10)
            if r==1 {
                msg = CommandWrap(ESC_BOLD_ON, msg, ESC_BOLD_OFF)
            } else if r==2 || r==3 {
                msg = ColorText(msg, 0)
            }
            
            msg = ColorText(msg, 32)
            fmt.Println(msg)
        }
    }

}

/******************************************************************************/
// MARK: - Application Functions

// Handle ctr-c events
func HandleControlC() {
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Printf("\033[2K\r") //clear line and move back to beginning
        fmt.Printf("\n\033[2K\r") //clear line and move back to beginning
        fmt.Printf("\n\033[2K\r") //clear line and move back to beginning
        
        fmt.Printf("%dx%d\n", getWidth(), getHeight())        
        Shutdown()
        os.Exit(0)
    }()

}

func Startup() {
    TerminalCommand(ESC_SAVE_SCREEN)
    TerminalCommand(ESC_CURSOR_OFF)
    TerminalCommand(ESC_CLEAR_SCREEN)
}

func Shutdown() {
    TerminalCommand(ESC_CURSOR_ON)
    TerminalCommand(ESC_RESTORE_SCREEN)
}

func main() {
    //bothMode := flag.Bool("both", true, "display both width and height")
    //heightMode := flag.Bool("height", false, "height mode")
    //widthMode := flag.Bool("width", false, "width mode")
    
    flag.Parse()

    HandleControlC()
    var c chan string = make(chan string)
    
    Startup()
    defer Shutdown() //defer this just in case
    
    //setup the columns
    var i uint
    for i=0; i<getWidth(); i++ {
        go run(int(i), c)
        break
    }
    //draw a horizontal line down the left
    for i=0; i<getHeight();i++ {
        fmt.Println(" ")
    }
    
    screen := make([][]string, getHeight())
    for i := range(screen) {
        screen[i] = make([]string,getWidth())
    }
    
    /*****************/
    // -- Main Loop --
    
    //columns first
    col := 1
    for {
        //dont move all columns down
        if rand.Intn(10)<4 {
            for row := int(getHeight()-1) ; 0<row; row-- {
                if row==1 {
                    msg := <- c
                    screen[row][col] = msg
                } else {
                    screen[row][col] = screen[row-1][col]
                }
            }
        }
        col += 2
        if int(getWidth())<=col {
            col = 1
            PaintScreen(screen)
            time.Sleep(time.Millisecond * 250)
        }
    }
}
