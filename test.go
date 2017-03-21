package main

import ("fmt"
    /*"bufio"
    "os"
    "io"
    "bytes"*/
    "flag"
    "math/rand"
    "unsafe"
    "syscall"
    "strconv"
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

func letter(n int64) string {
    return strconv.FormatInt(n, 36)
}

func run(n int, c chan string) {
    //fmt.Println("\033[5;%dHHello\n", n)
    
    if n%2==0 {
        r := rand.Intn(64)
        rs := letter(int64(r))
        
        c <- fmt.Sprintf("\033[5;%dH%s\n", n, rs)
    }
}

/****/


func main() {
    //args := os.Args
    
    for i:=0 ; i<36; i++ {
        fmt.Printf("%d = %s\n", i, letter(int64(i)))
    }
    
    //fmt.Println("\033[5;5HHello\n")
    
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
