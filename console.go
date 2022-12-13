package main

import (
    "fmt"
    "flag"
    "syscall"
    "unsafe"
    )

/******************************************************************************/
// MARK: Structures

type winsize struct {
    Row    uint16
    Col    uint16
    Xpixel uint16
    Ypixel uint16
}

/******************************************************************************/
// MARK: - Functions

func GetWidth() uint {
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

func GetHeight() uint {
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

/******************************************************************************/
// MARK: - Application

func main() {
    heightMode := flag.Bool("height", false, "height mode")
    widthMode := flag.Bool("width", false, "width mode")
    
    flag.Parse()
    
    if *heightMode {
        fmt.Printf("%d\n", GetHeight())
    } else if *widthMode {
        fmt.Printf("%d\n", GetWidth())
    } else {
        fmt.Printf("%dx%d\n", GetWidth(), GetHeight())
    }
}
