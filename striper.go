package main

import ("fmt"
    "bufio"
    "os"
    "io"
    "bytes"
    "flag"
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
/****/


/**
pull each line and process
*/
func ReadLines(reader io.Reader, foo func(string) string) string {
    buf := bufio.NewReader(reader)
    var buffer bytes.Buffer
    line, err := buf.ReadBytes('\n');
    for err == nil {
        if foo!=nil {
            line = []byte(foo(string(line)))
        }
        buffer.Write(line)
        line, err = buf.ReadBytes('\n')
    }
    return buffer.String()
}

func main() {
    //args := os.Args
    
    lineMode := flag.Bool("line", false, "line mode, trim each line")
    edgeMode := flag.Bool("edge", false, "edge mode, trim just edges")
    allMode := flag.Bool("all", false, "all mode, trim everything")
    
    flag.Parse()
    
    if *lineMode {//each line gets trimmed, you get a trim, you get a trim
        str := ReadLines(os.Stdin, sansSpace)
        fmt.Printf("%s", str)
    } else if *edgeMode {//only first and last
        str := ReadLines(os.Stdin, pass)
        fmt.Printf("%s", string(bytes.TrimSpace([]byte(str))))
    } else if *otherMode {
        str := ReadLines(os.Stdin, all)
        fmt.Printf("%s", string(bytes.TrimSpace([]byte(str))))
    }
}

func pass(line string) string{return line}

func sansSpace(line string) string{
    return string(bytes.TrimSpace([]byte(line))) + "\n"
}

func all(line string) string{
    return string(bytes.TrimSpace([]byte(line)))
}
