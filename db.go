package main

import ("fmt"
    //"bufio"
    "os"
    /*"io"
    "bytes"
    "log"*/
    "flag"
    "io/ioutil"
    /*"math"
    "strconv"
    "os/exec"*/
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

type screen_buffers struct {
    left_hud string
    right_hud string
    content string
}

type App_Data struct {
    backlog_command string
    worker_command string
    backlog_list []string

    data map[string]interface{}
    verbose bool
    active_file string
}

var (
    history_fn = filepath.Join(os.TempDir(), ".rpn_history")    //used by liner
    names      = []string{"Create", "Read", "Update", "Delete"} //used by liner
)

var buffers = screen_buffers{left_hud: "", right_hud: "", content: ""}
var app_data = App_Data{backlog_command:"", worker_command:"", verbose:false}

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

//#mark - hi

/*func v(msg string) {
    if app_data.verbose {
        fmt.Printf("%s\n", [msg])
    }
}*/

func v(format string, args ...string) {
    if app_data.verbose {
        fmt.Printf(format, args)
    }
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
    fmt.Printf("here: %s\n", json_text)
    if err!=nil {
        fmt.Printf("error: %s\n", err)
    }
    err = ioutil.WriteFile(file, json_text, 0644)
    if err!=nil {
        fmt.Printf("Error: %s\n", err)
    } else {
        v("File %s has been saved\n", file)
    }
}

func List(data map[string]interface{}) {
    fmt.Printf("List: ")
    for k,v := range data {
        fmt.Printf("%s=%s ", k, v)
    }
    fmt.Printf("\n")
}

func Create(data map[string]interface{}, key string, value string) {
    data[key] = value
}

func Read(data map[string]interface{}, key string) {
    fmt.Printf("%s=%s\n", key, data[key])
}

func Update(data map[string]interface{}, key string, value string) {
    data[key] = value
}

func Delete(data map[string]interface{}, key string) {
    delete (data, key)
}

func Table() {
    header := ""
    rows := ""
    for k,v := range app_data.data {
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
            ProcessLine(input, data)  //use it
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

func ProcessLine(raw string, data map[string]interface{}) {
    list := strings.Split(raw, " ")
    command := list[0]
    args := list[1:]
    switch command {
        case "q", "quit", "exit":
            if app_data.verbose { fmt.Printf("getting out of here\n") }
            os.Exit(0)
        case "e", "echo":
            fmt.Printf("%s => %s\n", command, strings.Join(args, ",") )
        
        case "c", "create":
            Create(data, args[0], strings.Join(args[1:], " ") )
        case "r", "read":
            Read(data, args[0])
        case "u", "update":
            Update(data, args[0], strings.Join(args[1:], " ") )
        case "d", "delete":
            Delete(data, args[0])
        
        case "l", "list":
            List(data)
        
        case "t", "table":
            Table()
        case "s", "save":
            Save(data, app_data.active_file)
    }
}

// #mark
func main() {
    //backlogCommand := flag.String("load", "ps -ef | grep java", "command to generate work")
    //backlogCommand := flag.String("load", "ps -ef", "command to generate work")
    //workerCommand := flag.String("work", "echo %s", "command to work off the load")
    verbose := flag.Bool("verbose", false, "verbose")
    file_name := flag.String("file", "data.json", "data file")
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
    
    if app_data.verbose { List(data) }
    InteractiveAdvance(line, data)
    if app_data.verbose { List(data) }
}
