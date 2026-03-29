package main

import ("fmt"
    "bufio"
    "os"
    "io"
    "bytes"
    "flag"
    "syscall"
    "unsafe"
    "strings"
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

type appdata struct {
    streamMode bool
    lineMode bool
    edgeMode bool
    firstMode int
    lastMode int
    trimLeft bool
    trimRight bool
    all bool
}

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
    app_data := appdata{}
    
    streamMode := flag.Bool("stream", true, "steam mode, read and process each line and dump to stream")
    lineMode := flag.Bool("line", false, "line mode, process each line separately")
    edgeMode := flag.Bool("edge", false, "edge mode, trim just edges")
    firstMode := flag.Int("first", -1, "trim the first few charactors")
    lastMode := flag.Int("last", -1, "trim the last few charactors")
    trimLeft := flag.Bool("left", false, "trim leading spaces")
    trimRight := flag.Bool("right", false, "trim trailing spaces")
    all := flag.Bool("all", false, "remove all spaces")
    
    flag.Parse()

    app_data = appdata{
        streamMode: *streamMode,
        lineMode: *lineMode,
        edgeMode: *edgeMode,
        firstMode: *firstMode,
        lastMode: *lastMode,
        trimLeft: *trimLeft,
        trimRight: *trimRight,
        all: *all}
    
    if *streamMode {
        buf := bufio.NewReader(os.Stdin)
        bytelist, err := buf.ReadBytes('\n')
        for err == nil {
            line := string(bytelist)
            fmt.Printf("%s\n", workOnLine(line, app_data))
            bytelist, err = buf.ReadBytes('\n')
        }
    } else if *lineMode {
        str := ReadLines(os.Stdin, pass)
        scanner := bufio.NewScanner(strings.NewReader(str))
        output := ""
        for scanner.Scan() {
            line := scanner.Text()
            output = output + "\n" + workOnLine(line, app_data)
        }
        fmt.Printf("%s", output)
    } else {
        str := ReadLines(os.Stdin, pass)
        str = workOnLine(str, app_data)
        fmt.Printf("%s", str)
    }
}

func workOnLine(line string, data appdata) string {
    if data.all {
        line = strings.ReplaceAll(line, " ", "")
        line = strings.ReplaceAll(line, "\t", "")
    } else {
        //outer-space prefix in-space <whats-left> in-space postfix outer-space

        //remove outer-space
        if data.trimLeft {line = strings.TrimLeft(line, " \t")}
        if data.trimRight {line = strings.TrimRight(line, " \t\n")}
        
        //remove pre/post fix
        if data.firstMode != -1 {line = cutLeft(line, data.firstMode)}
        if data.lastMode != -1 {line = cutRight(line, data.lastMode)}
        
        //trim what is left
        if data.edgeMode {line = sansSpace(line)}
        
        //whats-left is all there is
    }
    return line
}

func pass(line string) string{return line}
func cutLeft(line string, count int) string {return line[count:]}
func cutRight(line string, count int) string {return line[0:len(line)-(1+count)]}
func sansSpace(line string) string{return string(bytes.TrimSpace([]byte(line)))}
