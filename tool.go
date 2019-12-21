package main

import ("fmt"
    "bufio"
    "os"
    /*"io"
    "bytes"*/
    "log"
    "flag"
    "math"
    "strconv"
    "os/exec"
    "strings"
    "syscall"
    "unsafe"
    )

/****/
type winsize struct {
    Row    uint16
    Col    uint16
    Xpixel uint16
    Ypixel uint16
}

type screen_buffers struct {
    left_hud string
    right_hud string
    content string
}

type App_Data struct {
    backlog_command string
    worker_command string
    backlog_list []string
}

var buffers = screen_buffers{left_hud: "", right_hud: "", content: ""}
var app_data = App_Data{backlog_command:"", worker_command:""}

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

//#mark - hi

// #mark

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

func PrintCtrOnErr(esc string) {
    fmt.Fprintf(os.Stderr, "\033[%s", esc)
}

func PrintCtrOnErrAt(esc string, y, x int) {
    fmt.Fprintf(os.Stderr, "\033[%d;%dH\033[%s", y, x)
}

func PrintStrOnErrAt(msg string, y, x int) {
    fmt.Fprintf(os.Stderr, "\033[%d;%dH%s", y, x, msg)
}

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

func PrintCtrOnOut(esc string) {
    fmt.Fprintf(os.Stderr, "\033[%s", esc)
}

func PrintCtrOnOutAt(esc string, y, x int) {
    fmt.Fprintf(os.Stderr, "\033[%d;%dH\033[%s", y, x)
}

func PrintStrOnOutAt(msg string, y, x int) {
    fmt.Fprintf(os.Stderr, "\033[%d;%dH%s", y, x, msg)
}

func PrintLeftBox(msg string, y int, x int, width int, height int) {
    side := strings.Repeat(" ", width-1) + "│"
    bottom := strings.Repeat("─", width-1) + "┘"
    PrintSideBox(msg, y, x, width, height, side, bottom)
}

func PrintRightBox(msg string, y int, x int, width int, height int) {
    side := "│" + strings.Repeat(" ", width-1)
    bottom := "└" + strings.Repeat("─", width-1)
    PrintSideBox(msg, y, x, width, height, side, bottom)
}

func PrintSideBox(msg string, y, x int, COLS int, ROWS int, side string, bottom string) {    
    msgs := strings.Split(msg, "\n")
    for i:=0 ; i<ROWS ; i++ {
        PrintStrOnOutAt(side, y+i,x)
        if i<len(msgs) {
            line := msgs[i]
            size := int( math.Min(float64(COLS)-2, float64(len(line))) )
            PrintStrOnOutAt(line[0:size], y+i, x+1)
        }
    }
    PrintStrOnOutAt(bottom, y+ROWS-1,x)
}

func do_command(script string) string {
    cmd := exec.Command("bash", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		//buffers.content = string(err)
        log.Fatalf("cmd.Run() failed with %s\n", err)
	}
    return string(out)
}

func load_work() {
    raw := do_command(app_data.backlog_command)
    app_data.backlog_list = strings.Split(raw, "\n")

buffers.content = strings.Join(app_data.backlog_list, "\n")

    buffers.left_hud = "c:" + strconv.FormatInt(int64(len(app_data.backlog_list)), 10)
}

func GetInput() string {
    out := ""
    buf := bufio.NewReader(os.Stdin)
    fmt.Print("> ")
    sentence, err := buf.ReadBytes('\n')
    if err != nil {
        fmt.Println(err)
    } else {
        out = string(sentence)
    }
    return out
}

func buffer_crop(input string, width int, height int) string{
    lines := strings.Split(input, "\n")
    start := int(math.Max(0,math.Min(float64(height), float64(len(lines)-height-3))))
    output := strings.Join(lines[start:],"\n")
    buffers.right_hud = strconv.FormatInt(int64(len(lines)), 10) + ", " +
        strconv.FormatInt(int64(len(lines[start:])), 10)
    return output
}

func Commands(cmd string) bool {
    running := true
    switch strings.Trim(cmd, "\n") {
    case "":
    case "nop":
        ;
	case "exit":
        running = false
	case "quit":
        fmt.Println("quit everything")
        running = false
    case "load":
        load_work()
    case "other":
        buffers.right_hud += "other\n"
    case "text":
        buffers.content += "All work and no play make jack a dull boy\n"
	default:
		running = true
	}
    return running
}

/****/

func main() {
    //backlogCommand := flag.String("load", "ps -ef | grep java", "command to generate work")
    backlogCommand := flag.String("load", "ps -ef", "command to generate work")
    workerCommand := flag.String("work", "echo %s", "command to work off the load")

    flag.Parse()
    
    h := int(getHeight())
    w := int(getWidth())

    ScrSave()

    //buffers.content = "\n\n\n\n\n\n" + *backlogCommand + "\n" + *workerCommand
    buffers.content = ""
    app_data.backlog_command = *backlogCommand
    app_data.worker_command = *workerCommand
    
    PrintStrOnOutAt("", int(h)-1, 0)
    
    //application loop
    loop := true // run at least once
    for loop==true {
        //content
        PrintStrOnOutAt(buffer_crop(buffers.content, w, h ), 0, 0)

        //hud
        PrintLeftBox(buffers.left_hud, 1,1, 10, 6)
        PrintRightBox(buffers.right_hud, 1,w-10, 10, 5)
        
        //Prompt
        PrintStrOnOutAt(strings.Repeat(" ", w-1), h-1, 0)
        PrintStrOnOutAt("", h-1, 0)
        loop = Commands(GetInput())
    }
    
    ScrRestore()
}
