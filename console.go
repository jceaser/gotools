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

type win_size struct {
    Row    uint16
    Col    uint16
    Xpixel uint16
    Ypixel uint16
}

/******************************************************************************/
// MARK: - Functions

func GetWidth() int {
    ws := &win_size{}
    retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
        uintptr(syscall.Stdin),
        uintptr(syscall.TIOCGWINSZ),
        uintptr(unsafe.Pointer(ws)))
    if int(retCode) == -1 {
        panic(errno)
    }
    return int(ws.Col)
}

func GetHeight() int {
    ws := &win_size{}
    retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
        uintptr(syscall.Stdin),
        uintptr(syscall.TIOCGWINSZ),
        uintptr(unsafe.Pointer(ws)))
    if int(retCode) == -1 {
        panic(errno)
    }
    return int(ws.Row)
}

func MaxInt(left, right int) int {
    if left<right {
        return right
    }
    return left
}

/******************************************************************************/
// MARK: - Application

func main() {
    heightMode := flag.Bool("height", false, "Height mode")
    widthMode := flag.Bool("width", false, "Width mode")
    adjust := flag.Int("adjust", 0, "Value to add to height or width")

    flag.Parse()

    if *heightMode {
        fmt.Printf("%d\n", MaxInt(0, GetHeight() + *adjust))
    } else if *widthMode {
        fmt.Printf("%d\n", MaxInt(0, GetWidth() + *adjust))
    } else {
        fmt.Printf("%dx%d\n",
            MaxInt(0, GetWidth() + *adjust),
            MaxInt(0, GetHeight() + *adjust))
    }
}
