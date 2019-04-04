package main

import ("fmt"
    "bufio"
    "os"
    "io"
    "bytes"
    "flag"
    "math"
    "strconv"
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

var stack []float64
var memory = make(map[string]float64)

func main() {
    //args := os.Args
    
    formula := flag.String("formula", "print", "math formula in RPN format")
    useStream := flag.Bool("stream", false, "use input stream")
    interactive := flag.Bool("interactive", false, "interactive mode")
    verbose := flag.Bool("verbose", false, "verbose")
    
    flag.Parse()
    
    if *useStream {
        scanner := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
	        a := fmt.Sprintf("%s %s", scanner.Text(), *formula)
            formula = &a
        }
    }

    if *interactive {
        str := ReadLines(os.Stdin, Filter)
        fmt.Printf("%s", str)
    } else  {
        for _, segment:= range strings.Split(*formula, " ") {
            cmd := segment
            muliplyer := 1
            if strings.Contains(cmd, ":") {
                cmd_parts := strings.Split(cmd, ":")
                //if len(cmd_parts)==2 {
                    raw_action := cmd_parts[0]
                    raw_multiplyer := cmd_parts[1]
                    times, err := strconv.Atoi(raw_multiplyer)
                    if err!=nil { times = 1; }
                    if 0<times {
                        muliplyer = times
                        cmd = raw_action
                    }
                //}
            }
            for i := 0; i< muliplyer ; i++ {
                Action(cmd, verbose)
            }
        }
    }
}

func Action (segment string, verbose *bool) {
    value, err := strconv.ParseFloat(segment, 64)
    if err==nil {
        Push(value)
    } else {
        switch segment {
            case "+": Plus()
            case "-": Minus()
            case "*": Times()
            case "/": Divide()
            case "%": Remainder()
            case "^": Power()
            case "min": Min()
            case "max": Max()

            case "rand":
            case "--": Decrement()
            case "++": Increment()
            case "^2": Square()
            case "print": Print()
            case "<>": Swap()
            case "<<": RotateLeft()
            case ">>": RotateRight()

            case "?<": IfLess()
            case "?>": IfOver()

            case "a","b","c","d","e","f","g","h","i","j","k","l","m",
                    "n","o","p","q","r","s","t","u","v","w","x","y","z":
                MemoryLoad(segment)
            case "A","B","C","D","E","F","G","H","I","J","K","L","M",
                    "N","O","P","Q","R","S","T","U","V","W","X","Y","Z":
                MemoryStore( strings.ToLower(segment) )
            case "dump": Dump()

            default:
                fmt.Printf("%s is an unknown command", segment)
        }
    }
    if *verbose {
        fmt.Printf("%v\n", stack)
    }

}

func MemoryLoad (key string) {//recall value
    var value = memory[key]     
    Push(value)
}

func MemoryStore (key string) {//save value
    memory[key] = Peek()
}

func Dump() {
    fmt.Printf("(%v)\n", memory);
}

func Push(value float64) {
    stack = append(stack, value)
}

func Pop() float64 {
    n := len(stack) - 1
    value := 0.0
    value, stack = stack[n], stack[:n]
    return value
}

func Peek() float64 {
    n := len(stack)-1
    return stack[n]
}

func PopQueue() float64 {
    value := stack[0]
    stack = stack[1:]
    return value
}

func Print() {
    fmt.Printf("(%v)\n", stack)
}

/**/
/* trinaries */

func IfLess() {
    fail := Pop()
    pass := Pop()
    test := Pop()
    if test<0 {
        Push(pass)
    } else {
        Push(fail)
    }
}

func IfOver() {
    fail := Pop()
    pass := Pop()
    test := Pop()
    if test>0 {
        Push(pass)
    } else {
        Push(fail)
    }
}

/**/
/* Doubles */

func Plus() {
    right := Pop()
    left := Pop()
    Push(left+right)
}
func Minus() {
    right := Pop()
    left := Pop()
    Push(left-right)
}
func Times() {
    right := Pop()
    left := Pop()
    Push(left*right)
}
func Divide() {
    right := Pop()
    left := Pop()
    Push(left/right)
}
func Power() {
    power := Pop()
    base := Pop()
    Push( math.Pow(base,power) )
}
func Min() {
    right := Pop()
    left := Pop()
    Push( math.Min(left,right) )
}
func Max() {
    right := Pop()
    left := Pop()
    Push( math.Max(left,right) )
}

func Remainder() {
    right := Pop()
    left := Pop()
    Push( math.Remainder(left,right) )
}

func Swap() {
    right := Pop()
    left := Pop()
    Push( right )
    Push( left )
}

/*****/
/* singletons */

func SquareRoot() {
    base := Pop()
    Push( math.Sqrt(base) )
}

func Square() {
    power := 2.0
    base := Pop()
    Push( math.Pow(base,power) )
}

func Round() {
    top := Pop()
    Push( math.Round(top) )
}

func RotateLeft() {
    //  <<  take off end and put at beginning
    //value := stack[0]
    //stack = stack[1:]
    value := PopQueue()
    Push(value)
}
func RotateRight() {
    //  >>  take off beginning and put at end
    value := Pop()
    stack = append([]float64{value}, stack...)
}

func Increment() {
    value := Pop()
    Push(value + 1)
}

func Decrement() {
    value := Pop()
    Push(value - 1)
}

func Filter (line string) string{
    return line
}

