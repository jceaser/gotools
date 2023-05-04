// This command will report on the width and/or hight of the console

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
    heightMode := flag.Bool("height", false, "Height mode")
    widthMode := flag.Bool("width", false, "Width mode")
    adjust := flag.Int("adjust", 0, "Value to add to height or width")
    
    flag.Parse()
    
    if *heightMode {
        fmt.Printf("%d\n", GetHeight() + uint(*adjust))
    } else if *widthMode {
        fmt.Printf("%d\n", GetWidth() + uint(*adjust))
    } else {
        fmt.Printf("%dx%d\n", GetWidth(), GetHeight())
    }
}
