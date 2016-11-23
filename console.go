package main

import ("fmt"
    /*"bufio"
    "os"
    "io"
    "bytes"*/
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


/****/


func main() {
    //args := os.Args
    
    lineMode := flag.Bool("line", true, "line mode, trim each line")
    edgeMode := flag.Bool("edge", false, "edge mode, trim just edges")
    allMode := flag.Bool("all", false, "all mode, trim everything")
    
    flag.Parse()
    
    str := "text"
    
    if *lineMode {
        fmt.Printf("%dx%d", getWidth(), getHeight())
    } else if *edgeMode {//only first and last
        fmt.Printf("%s", str)
    } else if *allMode {
        fmt.Printf("%s", str)
    }
}
