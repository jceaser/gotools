package main

import ("fmt"
    "os"
    //"reflect"
    "bufio"
    "errors"
    "io/ioutil"
    "regexp"
    "flag"
    "math"
    "math/rand"
    "sort"
    "strconv"
    "syscall"
    "unsafe"
    "strings"
    "path/filepath"
    "github.com/peterh/liner"
    )

/****/

/*
Function Data holds the function to call and it's description
*/
type func_data struct {
    cmd func()
    doc string
}

var (
    history_fn = filepath.Join(os.TempDir(), ".rpn_history")    //used by liner
    names      = []string{"print", "dump", "quit"}              //used by liner

    active_stack int
    stack [][]float64
    lastx float64
    memory = make(map[string]float64)
    modes = make(map[string]interface{})
    actions = make(map[string]func_data)
)

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

func setup_liner(line *liner.State) {
    line.SetCtrlCAborts(true)

    line.SetTabCompletionStyle(liner.TabPrints)
    line.SetCompleter(func(line string) (c []string) {
        for _, n := range names {
            fmt.Print(n)
            if strings.HasPrefix(n, strings.ToLower(line)) {
                c = append(c, n)
                fmt.Print(n)
            }
        }
        return
    })
    if f, err := os.Open(history_fn); err == nil {
        line.ReadHistory(f)
        f.Close()
    }
}

func a(key string, foo func(), doc string) {
    actions[key] = func_data{cmd:foo, doc:doc}
}

func InitializeStack() {
    active_stack = 0
    item := [][]float64{{}}
    stack = append(stack, item...)
}

func InitializeActions() {
    // push consts
    a("euler", E, "Decimal expansion of e - Euler's number")
    a("pi", Pi, "Decimal expansion of Pi (or, digits of Pi)")
    a("ln", Ln2, "Decimal expansion of the natural logarithm of 2")
    a("ln10", Ln10, "Decimal expansion of natural logarithm of 10")
    a("everything", Everything, "Answer to Everything")
    a("sol", func(){Push(299792458)}, "Speed of Light")
    a("lastx", func(){Push(lastx)}, "Last X value")

    // unary actions
    a("drop", Drop, "Remove the top item from the stack")
    a("pord", Pord, "Remove the bottom item from the stack (opposite of drop)")
    a("--", Decrement, "subtract one from the top of the stack")
    a("++", Increment, "add one to the top of the stack")
    a("^2", Square, "square the item at the top of the stack")
    a("rand", Random, "Generates a random number from 0-1")
    a("int", Truncate, "Return the integer part of the number")
    a("frac", Exponent, "Return the decimal part of the number");
    a("!", Factorial, "Factorial")
    a("rad", Rad, "Decimal to Radians")
    a("sin", Sin, "Sine of the radian")
    a("cos", Cos, "Cosine of the radian")
    a("tan", Tan, "Tangent of the radian")
    a("fmt", Format, "")
    a("dup", Dup, "Duplicate x")

    // binary actions
    a("&", And, "AND values")
    a("|", Or, "OR values")
    a("+", Plus, "add two numbers")
    a("-", Minus, "subtract two numbers")
    a("–", ReverseMinus, "subtract in reverse")
    a("*", Times, "multiply two numbers")
    a("/", Divide, "divide two numbers")
    a("\\", ReverseDivide, "divide two numbers, but reversed")
    a("%", Remainder, "divide two numbers return only the remainder")
    a("^", Power, "take the power of two numbers first^second")
    a("min", Min, "return the smaller of two numbers")
    a("max", Max, "return the larger of two numbers")
    a("<>", Swap, "swap the top two stack items")

    // ternary actions
    a("?<", IfLess, "if s[2]<0 then s[1] else s[0], consumes all three")
    a("?>", IfOver, "if s[2]>0 then s[1] else s[0], consumes all three")

    // Whole stack actions
    a("<<", RotateLeft, "shift stack down, wrapping")
    a(">>", RotateRight, "shift stack up, wrapping")
    a("avg", Average, "Take the average of the entire stack")
    a("stddev", StandardDeviation, "Take the standard deviation of the stack")
    a("sort", Sort, "sort the stack")
    a("med", Median, "find the medium value on the stack")
    a("clear", Clear, "Empty the stack")

    //other actions
    a("sswap", SwapStacks, "swap stacks")
    a("sprint", PrintStacks, "print all stacks")
    a("sprinti", PrintStacksInfo, "Print stack info")
    a("sadd", AddStack, "add a stack")

    a("quit", Exit, "Quit application, same as exit")
    a("exit", Exit, "Quit application, same as quit")
    a("help", Help, "Display a help document")
    a("print", Print, "print the stack")
    a("vprint", PrintV, "print the stack")
    a("dump", Dump, "print out memory storage")
}

func fish() {
    no_args := "complete --command rpn --long-option %s --description \"%s\"\n"
    with_args := "complete --command rpn --long-option %s --arguments \"%s\" --description \"%s\"\n"

    fmt.Printf(with_args, "formula", "code", "math formula to compute")
    fmt.Printf(no_args, "interactive", "ask user for input and loop")
    fmt.Printf(no_args, "verbose", "print out more logs")
    fmt.Printf(no_args, "pop", "Return the final stack value")
}

func main() {
    InitializeStack()
    InitializeActions()

    //readline setup
    line := liner.NewLiner()
    defer line.Close()
    setup_liner(line)

    //flag setup
    formula := flag.String("formula", "print",
        "math formula in RPN format, interpreted before stream")
    file := flag.String("file", "", "print out function manual")

    interactive := flag.Bool("interactive", false,
        "interactive mode using a readline style interface")
    verbose := flag.Bool("verbose", false, "verbose")
    final_pop := flag.Bool("pop", false,
        "output a final pop")
    show_fish := flag.Bool("_completion", false, "shell completion help")
    manual := flag.Bool("manual", false, "print out function manual")

    flag.Parse()

    if *show_fish {
        fish()
         os.Exit(0)
    }

    if 0<len(*file) {
        _, error := os.Stat(*file)
        if !errors.Is(error, os.ErrNotExist) {
            fileContent, err := ioutil.ReadFile(*file)
            if err == nil {
                text := string(fileContent)
                if len(*formula)>0 {
                    together := text + " " + *formula
                    formula = &together
                } else {
                    formula = &text
                }
            }
        }
    }

    if *manual {
        Help()
    }

    //process stream if it exists
    stat, _ := os.Stdin.Stat()
    if  ( (stat.Mode() & os.ModeCharDevice) == 0 ) {
        scanner := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
            a := fmt.Sprintf("%s %s", *formula, scanner.Text())
            formula = &a
        }
    }

    if *interactive {
        InteractiveAdvance(line, verbose)
        //InteractiveBasic()
    } else  {
        ProcessLine(*formula, *verbose)
    }
    if *final_pop {
        fmt.Println(Pop())
    }
}

/**
run the interactive mode using the third party readline library. Help the
library stor history, take each line and send it to ProcessLine()
*/
func InteractiveAdvance(line *liner.State, verbose *bool) {
    fmt.Printf("RPN Calculator by thomas.cherry@gmail.com\n")
    for {
        if name, err := line.Prompt(">"); err == nil {
            input := strings.Trim(name, " ")    //clean it
            line.AppendHistory(name)            //save it
            ProcessLine(input, *verbose)        //use it
        } else if err == liner.ErrPromptAborted {
            fmt.Print("Aborted")
        } else {
            fmt.Print("Error reading line: ", err)
        }
        //save the history
        if f, err := os.Create(history_fn); err != nil {
            fmt.Print("Error creating history file: ", err)
        } else {
            line.WriteHistory(f)
            f.Close()
        }
    }
}

/**
this is a fall back method in case it is decided to not use a third party
interface
*/
func InteractiveBasic(verbose *bool) {
    fmt.Print("Enter text: ")
    for {
        fmt.Printf(">")
        var input string
        input = ReadFormula()
        input = strings.Trim(input, " ")
        ProcessLine(input, *verbose)
    }
}

/** used */
func ReadFormula() string {
    /*var b []byte = make([]byte, 1)*/

    fmt.Print("Enter text: ")
    for {
        fmt.Printf(">")
        var input string
        var ascii int
        getChar(ascii /*, keyCode, err*/)
        fmt.Printf(string(ascii))
        fmt.Printf(input)
    }
}

/** used */
func getChar(ascii int) {
    reader := bufio.NewReader(os.Stdin)
    // ...
    ch, _, err := reader.ReadRune()
    fmt.Printf(string(ch))
    if err != nil {
        fmt.Println("Error reading key...", err)
    }
}

func Calculate(formula string) float64 {
    ProcessLine(formula, false)
    return Pop()
}

func split_on_spaces(raw string) []string {
    cleaned := strings.TrimSpace(raw)
    rule := regexp.MustCompile("\\s+")
    list := rule.Split(cleaned, -1)
    return list
}

/**
process a formula line
*/
func ProcessLine(formula string, verbose bool) {
    labels := make(map[int]int)
    commands := split_on_spaces(formula)

    for lc := 0 ; lc <= len(commands)-1; lc {
        segment := commands[lc]
        cmd := segment
        if cmd=="lbl" {
            labels[int(Peek())] = lc
        }
    }

    fmt.Println(labels)

    for pc := 0 ; pc <= len(commands)-1; pc++ {
        segment := commands[pc]
        cmd := segment
        if cmd=="lbl" {
            label := int(Pop())
            labels[label] = pc
            continue
        } else if cmd=="j-" {
            label := int(Pop())
            test := Pop()
            if test<0.0 {
                pc = labels[label]
            }
            continue
        } else if cmd=="j+" {
            label := int(Pop())
            test := Pop()
            if test>0.0 {
                pc = labels[label]
            }
            continue
        } else if cmd=="j=" {
            label := int(Pop())
            test := Pop()
            if test==0.0 {
                pc = labels[label]
            }
            continue
        } else if cmd=="goto" {
            label := int(Pop())
            pc = labels[label]
            continue
        }

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

/**
Perform a single action/operation. Either store a value or operate on it
*/
func Action (segment string, verbose bool) {
    value, err := strconv.ParseFloat(segment, 64)
    if err==nil {
        //must be a number, put it on the stack
        Push(value)
    } else {
        //it is a command
        foo := actions[segment].cmd
        doc := actions[segment].doc

        //fmt.Println("foo = ", reflect.ValueOf(foo).Kind())

        if verbose && len(doc)>0 {
            fmt.Printf("%s\n", doc)
        }

        if foo!=nil {
            foo()
        } else {
            switch segment {
                case "a","b","c","d","e","f","g","h","i","j","k","l","m",
                        "n","o","p","q","r","s","t","u","v","w","x","y","z":
                    MemoryLoad(segment)
                case "A","B","C","D","E","F","G","H","I","J","K","L","M",
                        "N","O","P","Q","R","S","T","U","V","W","X","Y","Z":
                    MemoryStore( strings.ToLower(segment) )
                default:
                    fmt.Printf("%s is an unknown command\n", segment)
            }
        }
    }

    if verbose {
        fmt.Printf("%v\n", stack)
    }

}

/******************************************************************************/
// #mark - General Operations

// memory functions

func cleanKey(raw string) string {
    var ans string
        cleaner := strings.ToLower(strings.Trim(raw, " "))
        ans = cleaner
    return ans
}

/**
Recall a value from memory to the stack
@param key value name
*/
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

// stack functions

func ActiveStack(params ...[]float64) []float64 {
    return stack[active_stack]
}

func ActiveStackUpdate(params ...[]float64) {
    if params!=nil && len(params)>0 && params[0]!=nil {
        stack[active_stack] = params[0]
    }
}

func Items() bool {
    return len(stack[active_stack])>0
}

func Empty() bool {
    return len(stack[active_stack])<1
}

func Push(value float64) {
    if !math.IsNaN(value) {
        stack[active_stack] = append(stack[active_stack], value)
        //ActiveStackUpdate(append(ActiveStack(), value))
    }
}

/**
Remove the top of the stack and return it
*/
func Pop() float64 {
    l := len(stack[active_stack])
    if l < 1 {
        fmt.Printf("stack is empty %d\n", l)
        return math.NaN()
    }
    n := l - 1
    value := math.NaN()
    value, stack[active_stack] = stack[active_stack][n], stack[active_stack][:n]
    return value
}

/**
Return the top of the stack, but don't remove it
*/
func Peek() float64 {
    n := len(stack[active_stack])-1
    return stack[active_stack][n]
}

func PopQueue() float64 {
    value := stack[active_stack][0]
    stack[active_stack] = stack[active_stack][1:]
    return value
}

// print functions
func Print() {
    raw_format, okay := modes["format"]
    format := "%v" + "\n"
    if okay {
        switch raw_format.(type) {
        case string:
            format = raw_format.(string)
        default:
            format = "%v" + "\n"
        }
    }
    fmt.Printf(format, stack[active_stack])
    //fmt.Printf("%v\n", stack[active_stack])
}

func PrintV() {
    l := len(stack[active_stack])
    //for i := len(stack[active_stack])-1 ; 0<=i ; i-- {
    //for i, v := range stack[active_stack] {
    for i := 0 ; i < l ; i++ {
        v := stack[active_stack][i]
        fmt.Printf("%d: %v\n", l-i-1, v)
    }
    fmt.Printf("%v\n", lastx)
}

func Help() {
    fmt.Printf("%s\n", strings.Repeat("*", 80) )

    keys := make( []string, 0 )
    for k := range actions {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    for _, k := range keys {
        doc := actions[k].doc
        fmt.Printf("%10s = %s\n", k, doc)
    }
    fmt.Printf("%s\n", strings.Repeat("* ", 40) )

    fmt.Printf("%s -- %s\n",
        "a-z will load values (read)",
        "A-Z will store values (write)")
    fmt.Printf("%s\n", "`action:count` will call action 'count' times : `+:3`")
}

func AddStack() {
    item := [][]float64{{}}
    stack = append(stack, item...)
}

func SwapStacks() {
    if len(stack)==1 {AddStack()}

    if active_stack < len(stack)-1 {
        active_stack++
    } else {
        active_stack = 0
    }
}

func PrintStacksInfo() {
    fmt.Printf( "looking at %d of %d stacks\n", active_stack+1, len(stack) )
}

func PrintStacks() {
    fmt.Printf("%v\n", stack)
}

// system functions

func Exit(){
    os.Exit(0)
}

/**************************************/
// #mark - ternary operators

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

/**************************************/
// #mark - binary operators

func And() {
    right := Pop()
    left := Pop()
    lastx = right
    Push(float64(int(left)&int(right)))
}
func Or() {
    right := Pop()
    left := Pop()
    lastx = right
    Push(float64(int(left)|int(right)))
}
func Xor() {
    right := Pop()
    left := Pop()
    lastx = right
    Push(float64(int(left)^int(right)))
}
func Plus() {
    right := Pop()
    left := Pop()
    lastx = right
    Push(left+right)
}
func Minus() {
    right := Pop()
    left := Pop()
    lastx = right
    Push(left-right)
}

func ReverseMinus() {
    left := Pop()
    right := Pop()
    lastx = left
    Push(left-right)
}

func Times() {
    right := Pop()
    left := Pop()
    lastx = right
    Push(left*right)
}
func Divide() {
    right := Pop()
    left := Pop()
    lastx = right
    Push(left/right)
}
func ReverseDivide() {
    left := Pop()
    right := Pop()
    lastx = right
    Push(left/right)
}
func Power() {
    power := Pop()
    base := Pop()
    lastx = power
    Push( math.Pow(base,power) )
}
func Min() {
    right := Pop()
    left := Pop()
    lastx = right
    Push( math.Min(left,right) )
}
func Max() {
    right := Pop()
    left := Pop()
    lastx = right
    Push( math.Max(left,right) )
}

func Remainder() {
    right := Pop()
    left := Pop()
    lastx = right
    Push( math.Remainder(left,right) )
}

func Swap() {
    right := Pop()
    left := Pop()
    Push( right )
    Push( left )
}

/**************************************/
// #mark - unary operators

func Format() {
    decimals := int64(Pop())
    if decimals < 0 {
        modes["format"] = nil
    } else {
        modes["format"] = fmt.Sprintf ("%%.%df\n", decimals);
    }
}

func Drop() {Pop()}

func Pord() {
    RotateLeft()
    Pop()
}

func SquareRoot() {
    x := Pop()
    lastx = x
    Push( math.Sqrt(x) )
}

func Square() {
    power := 2.0
    base := Pop()
    lastx = power
    Push( math.Pow(base,power) )
}

func Round() {Push( math.Round(Pop()) )}

func Increment() {
    value := Pop()
    lastx = value
    Push(value + 1)
}

func Decrement() {
    value := Pop()
    lastx = value
    Push(value - 1)
}

func Random() {Push(rand.Float64())}

func Truncate() {Push(math.Trunc(Pop()))}

func Exponent() {
    _, exp := math.Modf(Pop())
    Push(exp)
}

func Factorial(){
    ans:=1.0
    for i:=math.Floor(Pop()); i>0; i-- {
        ans = ans * i
    }
    Push(ans)
}

func Dup() {Push(Peek())}

//degrees = radians × 180° / π
//radians = degrees × π / 180°

func Rad(){Push(Pop()*math.Pi/180.0)}

func Sin(){Push(math.Sin(Pop()))}
func Cos(){Push(math.Cos(Pop()))}
func Tan(){Push(math.Tan(Pop()))}

func Everything() {Push(42.0)}
func E() {Push(math.E)}
func Pi() {Push(math.Pi)}
func Ln2() {Push(math.Ln2)}
func Ln10() {Push(math.Ln10)}

/**************************************/
/* full stack operators */

/** clear the stack ; empty the stack */
func Clear() {stack[active_stack] = []float64{}}

/** Rotate the stack by taking off the end and putting at the beginning */
func RotateLeft() {
    //  <<  take off end and put at beginning
    value := PopQueue()
    Push(value)
}

/** Rotate the stack by taking off beginning and putting at the end */
func RotateRight() {
    //  >>  take off beginning and put at end
    value := Pop()
    stack[active_stack] = append([]float64{value}, stack[active_stack]...)
}

/** Average the entire stack */
func Average() {
    var total = 0.0;
    var count = 0.0
    for Items() {
        total = total + Pop()
        count = count + 1
    }
    ans := total / count
    Push(ans)
}

/**
>3 5 9 1 8 6 58 9 4 10 stddev print
([15.8117045254457])
*/
func StandardDeviation() {
    var sum, mean, sd float64 = 0, 0, 0
    count_i := len(stack[active_stack])
    count_f := float64(count_i)

    for i:=0 ; i<count_i; i++ {
        sum += stack[active_stack][i]
    }
    mean = sum / count_f
    for i:=0 ; i<count_i; i++ {
        sd += math.Pow( stack[active_stack][i]-mean, 2)
    }
    sd = math.Sqrt( sd / count_f)
    Clear()
    Push(sd)
}

func Sort() {sort.Float64s(stack[active_stack])}

/**
>3 5 9 1 8 6 58 9 4 10 med print
([7])
>clear
>3 5 9 1 8 6 58 9 4 med print
([6])
*/
func Median() {
    Sort()
    med := 0.0
    count := len(stack[active_stack])/2
    if count % 2 == 0 {//odd use case
        med = stack[active_stack][count]
    } else {//even use case
        med = ( stack[active_stack][count] + stack[active_stack][count-1] ) / 2
    }
    Clear()
    Push(med)
}

func Filter (line string) string{
    return line
}

