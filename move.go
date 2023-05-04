package main

import ("os"
    "fmt"
    "io"
    "errors"
    //"flag"
    //"time"
    "unsafe"
    //"strconv"
    //"strings"
    "syscall"
    "io/ioutil"
    "os/exec"
    //"os/signal"

    //"bufio"

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

/**
Util Function Read a file
Not tested!
@param full path to read
@return empty string on error, file contents otherwise
*/
func readFile(file_path string) string {
    content, err := ioutil.ReadFile(file_path)
    if err != nil {
        os.Stderr.WriteString(err.Error() + "\n")
        return ""
    }
    return string(content)
}

func processLetter(b byte) {
    if b != '\033' {
        fmt.Printf("got one letter %s\n", string(b))
    }
}

func processLine(line string) {
    fmt.Printf("got one line %s\n", line)
}

func processCode(line string) {
    fmt.Printf("got one code %s\n", line)
}


func startKeyListener(c chan byte) {
    defer close(c)
    
    exec.Command("stty", "-f", "/dev/tty", "cbreak", "min", "1").Run()
    exec.Command("stty", "-f", "/dev/tty", "-echo").Run()
    defer exec.Command("stty", "-f", "/dev/tty", "echo").Run()
    
    f, err := os.Open("/dev/tty")
    if err != nil {
        fmt.Printf("open error\n")
        os.Stderr.WriteString(err.Error() + "\n")
        panic(err)
    }
    defer f.Close()
    
    line := ""
    three := ""
    five := ""
    b := make([]byte, 1)
    for {
        _, err := f.Read(b)
        if err != nil && !errors.Is(err, io.EOF) {
            fmt.Printf("read error %v", err)
            break
        }
        
        if b[0] == '\n' {
            if line == "quit" {
                break
            }
            processLine(line)
            line = ""
        } else {
            c <-b[0]
            processLetter(b[0])
            line = line + string(b[0])
            three = three + string(b[0])
            five = five + string(b[0])
            
            if len(three)>3 {
                three = three[1:4]
            }
            if len(five)>5 {
                five = five[1:6]
            }
            switch three {
            case "\033[A":
                processCode("up")
                three = ""
            case "\033[B":
                processCode("down")
                three = ""
            case "\033[C":
                processCode("right")
                three = ""
            case "\033[D":
                processCode("left")
                three = ""
            case "\033OP":
                processCode("F1")
                three = ""
            case "\033OQ":
                processCode("F2")
                three = ""
            case "\033OR":
                processCode("F3")
                three = ""
            case "\033OS":
                processCode("F4")
                three = ""
            }
            
            switch five {
            case "\033[15~":
                processCode("F5")
                three = ""
            case "\033[17~":
                processCode("F6")
                three = ""
            case "\033[18~":
                processCode("F7")
                three = ""
            case "\033[19~":
                processCode("F8")
                three = ""
            case "\033[20~":
                processCode("F9")
                three = ""
            case "\033[21~":
                processCode("F10")
                three = ""
            case "\033[22~", "\033[23~":
                processCode("F11")
                three = ""
            case "\033[24~":
                processCode("F12")
                three = ""
            }
        }

        if err != nil {
            // end of file
            fmt.Printf("end of file\n")
            break
        }
    }
}

func init() {
}

func main() {
    ch := make(chan byte)
    go startKeyListener(ch)
    
    running := true
    for running {
        if letter, okay := <-ch ; okay {
            if letter == 0x0000 {
                fmt.Println (string(letter))
            }
        } else {
            running = false
        }
    }
}