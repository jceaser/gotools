package main

import ("fmt"
    //"bufio"
    "os"
    /*"io"
    "bytes"
    "log"*/
    "flag"
    "sort"
    "io/ioutil"
    //"math"
    "strconv"
    //"os/exec"
    "strings"
    /*"syscall"
    "unsafe"*/
    "path/filepath"
    "encoding/json"
    "github.com/peterh/liner"
    )

/****/
type winsize struct {
    Row    uint16
    Col    uint16
    Xpixel uint16
    Ypixel uint16
}

type App_Data struct {
    data map[string]interface{}
    verbose bool
    active_file string
}

var (
    history_fn = filepath.Join(os.TempDir(), ".rpn_history")    //used by liner
    names      = []string{"Create", "Read", "Update", "Delete"} //used by liner
)

var app_data = App_Data{active_file:"", verbose:false}

const (
    RuneSterling = '£'
    RuneDArrow   = '↓'
    RuneLArrow   = '←'
    RuneRArrow   = '→'
    RuneUArrow   = '↑'
    RuneBullet   = '·'
    RuneBoard    = '░'
    RuneCkBoard  = '▒'
    RuneDegree   = '°'
    RuneDiamond  = '◆'
    RuneGEqual   = '≥'
    RunePi       = 'π'
    RuneHLine    = '─'
    RuneLantern  = '§'
    RunePlus     = '┼'
    RuneLEqual   = '≤'
    RuneLLCorner = '└'
    RuneLRCorner = '┘'
    RuneNEqual   = '≠'
    RunePlMinus  = '±'
    RuneS1       = '⎺'
    RuneS3       = '⎻'
    RuneS7       = '⎼'
    RuneS9       = '⎽'
    RuneBlock    = '█'
    RuneTTee     = '┬'
    RuneRTee     = '┤'
    RuneLTee     = '├'
    RuneBTee     = '┴'
    RuneULCorner = '┌'
    RuneURCorner = '┐'
    RuneVLine    = '│' //'│'
    RuneUVLine   = '╷'
    RuneDVLine   = '╵'
)

const (
    ESC_SAVE_SCREEN = "?47h"
    ESC_RESTORE_SCREEN = "?47l"
    
    ESC_SAVE_CURSOR = "s"
    ESC_RESTORE_CURSOR = "u"
    
    ESC_BOLD_ON = "1m"
    ESC_BOLD_OFF = "0m"
    
    ESC_CURSOR_ON = "?25h"
    ESC_CURSOR_OFF = "?25l"
    
    ESC_CLEAR_SCREEN = "2J"
    ESC_CLEAR_LINE = "2K"
)

//#mark - functions

func v(format string, args ...string) {
    if app_data.verbose {
        fmt.Printf(format, args)
    }
}

func jsonToMap(raw string) interface{} {
    var json_data interface{}
    json.Unmarshal([]byte(raw), &json_data)
    return json_data
}

/** return sorted keys from a map of interfaces */
func sorted_keys(data map[string]interface{}) []string {
    keys := make([]string, len(data))
    i := 0
    for k := range data {
        keys[i] = k
        i++
    }
    sort.Strings(keys)
    return keys
}

func load(file string) *os.File {
    json_raw, err := os.Open(file)
    if err!=nil {
        if os.IsNotExist(err) {
            //create the file because it does not exist
            v("Creating data file %s\n", file)
            sample := []byte("{}")
            err := ioutil.WriteFile(file, sample, 0644)
            if err!=nil {
                fmt.Printf("Error: %s\n", err)
                return nil
            }
            //try to open it a second time
            json_raw, err = os.Open(file)
            if err!=nil {
                fmt.Printf("Error: %s\n", err)
                return nil
            }
        } else {
            fmt.Printf("Error: %s\n", err)
            return nil
        }
    }
    //defer json_raw.Close()
    return json_raw
}

func Load(file string) map[string]interface{} {
    v("Loading file %s\n", file ) 
    json_raw := load(file)
    if json_raw==nil {
        fmt.Printf("No data\n")
    } else {
        defer json_raw.Close()
        var json_data map[string]interface{}
        bytes, err := ioutil.ReadAll(json_raw)
        if err!=nil {
            fmt.Printf("Error: %s\n", err)
        } else {
            json.Unmarshal([]byte(bytes), &json_data)
            return json_data
        }
    }
    return nil
}

func Save(data map[string]interface{}, file string) {
    json_text, err := json.Marshal(data)
    if err!=nil {
        fmt.Printf("error: %s\n", err)
        return
    }
    err = ioutil.WriteFile(file, json_text, 0644)
    if err!=nil {
        fmt.Printf("Error: %s\n", err)
    } else {
        v("File %s has been saved\n", file)
    }
}

func List(data map[string]interface{}) {
    //v("List: ")
    for i, k := range sorted_keys(data) {
        if i>0 {
            fmt.Printf(", ")
        }
        fmt.Printf("%s=%v", k, data[k])
    }
    fmt.Printf("\n")
}

func Create(data map[string]interface{}, key string, value string) {
    if data[key] == nil {
        //try to parse the value, turn numbers into a number
        if number, err := strconv.ParseFloat(value, 64) ; err==nil {
            //no error, value is a number
            data[key] = number
        } else {
            data[key] = jsonToMap(value)
        }
        //data[key] = value
    } else {
        fmt.Printf ("key already exists\n")
    }
}

func Read(data map[string]interface{}, key string) {
    if data[key] == nil {
        fmt.Printf("key does not exist\n")
    } else {
        //v("%s=", key)
        value := fmt.Sprintf("%v", data[key])
        if number, err := strconv.ParseFloat(value, 64) ; err==nil {
            fmt.Printf("%f\n", number)
        } else {
            fmt.Printf("%s\n", value)
        }
    }
}

func Update(data map[string]interface{}, key string, value string) {
    if data[key] != nil {
        //try to parse the value, turn numbers into a number
        if number, err := strconv.ParseFloat(value, 64) ; err==nil {
            //no error, value is a number
            data[key] = number
        } else {
            data[key] = jsonToMap(value)
        }
    } else {
        fmt.Printf ("key does not exists\n")
    }
}

func Delete(data map[string]interface{}, key string) {
    delete (data, key)
}

func Dump(data map[string]interface{}) {
    json_text, err := json.Marshal(data)
    if err!=nil {
        fmt.Printf("error: %s\n", err)
    } else {
        fmt.Printf("%s\n", json_text)
    }
}

func Table(data map[string]interface{}) {
    header := ""
    rows := ""
    for k,v := range data {
        if len(header)>0 {
            header = fmt.Sprintf("%s, %s", header, k)
        } else {
            header = k
        }

        if len(rows)>0 {
            rows = fmt.Sprintf("%s, %v", rows, v)
        } else {
            rows = fmt.Sprintf("%v", v)
        }
    }
    fmt.Printf("%s\n", header)
    fmt.Printf("%s\n", rows)
}

func Math(data map[string]interface{}, key string,
        operation func(float64, float64) float64) {
    if data[key] == nil {
        data[key] = 0.0
    } else {
        value := fmt.Sprintf("%v", data[key])
        if number, err := strconv.ParseFloat(value, 64) ; err==nil {
            //data[key] = number + 1
            data[key] = operation(number, 1)
        }
    }
}

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

/**
run the interactive mode using the third party readline library. Help the 
library stor history, take each line and send it to ProcessLine()
*/
func InteractiveAdvance(line *liner.State, data map[string]interface{}) {
    fmt.Printf("Database by thomas.cherry@gmail.com\n")
    for {
        if name, err := line.Prompt(">"); err == nil {
            input := strings.Trim(name, " ")    //clean it
            line.AppendHistory(name)            //save it
            ProcessManyLines(input, data)  //use it
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

func Help() {
    fmt.Printf("Database by thomas.cherry@gmail.com\n")
    fmt.Printf("Manage table data with optional form display.\n")
    fmt.Printf("\nNote: Arguments with ? are optional\n\n")

    format := "%4s %-14s %-14s %-40s\n"

    forty := strings.Repeat("-",40)
    fmt.Printf(format, "Flag", "Long", "Arguments", "Description")
    fmt.Printf(format,"----","------------","------------",forty)
    fmt.Printf(format, "c", "create", "name value", "create a name and value")
    fmt.Printf(format, "r", "read", "name", "read a named value")
    fmt.Printf(format, "u", "update", "name value", "update a named value")
    fmt.Printf(format, "d", "delete", "name", "delete a named value")
    fmt.Printf(format, "", "", "", "")

    fmt.Printf(format, "", "dump", "", "return current JSON")
    fmt.Printf(format, "e", "echo", "text", "echo out text")
    fmt.Printf(format, "h", "help", "", "Display this help")
    fmt.Printf(format, "l", "list", "", "List table")
    fmt.Printf(format, "", "ls", "", "List table")
    fmt.Printf(format, "L", "load", "file", "load new active file")
    fmt.Printf(format, "q", "quit", "", "quit application")
    fmt.Printf(format, "", "exit", "", "quit application")
    fmt.Printf(format, "S", "save", "", "Save active file")
    fmt.Printf(format, "t", "table", "", "display output as a table")
}

/**
Takes a raw command which may contain multiple instructions and break them up
into single commands which can be processed by ProcessLine()
*/
func ProcessManyLines(raw_line string, data map[string]interface{}) {
    if 0<len(raw_line) {
        commands := strings.Split(raw_line, ";")
        for _, raw_command := range commands {
            command := strings.Trim(raw_command, " ")
            if 0<len(command) {
                ProcessLine(command, data)
            }
        }
    }
}

/**
Take a raw string which may contain a command and execute it
*/
func ProcessLine(raw string, data map[string]interface{}) {
    list := strings.Split(raw, " ")
    command := list[0]
    args := list[1:]
    switch command {
            case "q", "quit", "exit":
            if app_data.verbose { fmt.Printf("getting out of here\n") }
            os.Exit(0)
        case "e", "echo":
            fmt.Printf("%s\n", strings.Join(args, ",") )
        
        case "c", "create":
            Create(data, args[0], strings.Join(args[1:], " ") )
        case "r", "read":
            Read(data, args[0])
        case "u", "update":
            Update(data, args[0], strings.Join(args[1:], " ") )
        case "d", "delete":
            Delete(data, args[0])
        
        case "add":
            Math(data, args[0], func(left float64, right float64) float64 {
                return left + right
            })

        case "sub":
            Math(data, args[0], func(left float64, right float64) float64 {
                return left - right
            })

        case "dump":
            Dump(data)
        case "l", "ls", "list":
            List(data)
        case "L", "load":
            app_data.active_file = args[0]
            data := Load(app_data.active_file)
            app_data.data = data
        
        case "t", "table":
            Table(data)
        case "S", "save":
            Save(data, app_data.active_file)
        case "h", "help":
            Help()
        default:
            Help()
    }
}

// #mark
func main() {
    verbose := flag.Bool("verbose", false, "verbose")
    file_name := flag.String("file", "db.json", "data file")
    help_flag := flag.Bool("manual", false, "Display help")
    command := flag.String("command", "", "Command to run")
    flag.Parse()

    app_data.verbose = *verbose
    app_data.active_file = *file_name
    data := Load(app_data.active_file)
    app_data.data = data

    //readline setup
    line := liner.NewLiner()
    defer line.Close()
    setup_liner(line)

    //h := int(getHeight())
    //w := int(getWidth())
    
    if *help_flag {
        Help()
        return
    }

    if *command == "" {
        if app_data.verbose { List(data) }
        InteractiveAdvance(line, data)
        if app_data.verbose { List(data) }
    } else {
        ProcessManyLines(*command, data)
    }
}
