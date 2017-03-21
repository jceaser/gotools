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
    
    bothMode := flag.Bool("both", true, "display both width and height")
    heightMode := flag.Bool("height", false, "height mode")
    widthMode := flag.Bool("width", false, "width mode")
    
    flag.Parse()
    
    if *bothMode {
        fmt.Printf("%dx%d\n", getWidth(), getHeight())
    } else if *heightMode {//only first and last
        fmt.Printf("%d\n", getHeight())
    } else if *widthMode {
        fmt.Printf("%d\n", getWidth())
    }
}
