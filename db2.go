package main

import ("fmt"
    //"bufio"
    "os"
    /*"io" */
    "bytes"
    /* "log"*/
    "flag"
    "io/ioutil"
    /*"math"*/
    "strconv"
    /*"os/exec"*/
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

type Form struct {
    Name string
    Columns []string
    Settings map[string]string
}

type DataBase struct {
    Columns map[string][]interface{}
    Forms map[string][]string
    Calculations map[string]string
    Settings map[string]string
}

type App_Data struct {
    backlog_command string
    worker_command string
    backlog_list []string

    //data map[string]interface{}
    data DataBase
    verbose bool
    indent_file bool
    active_file string
    running bool
}

var (
    history_fn = filepath.Join(os.TempDir(), ".rpn_history")    //used by liner
    names      = []string{"Create", "Read", "Update", "Delete"} //used by liner
)

var buffers = screen_buffers{left_hud: "", right_hud: "", content: ""}
var app_data = App_Data{backlog_command:"",
    worker_command:"",
    indent_file:false,
    verbose:false }

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

// #mark hi

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

func Load(file string) *DataBase {
    v("Loading file %s\n", file ) 
    json_raw := load(file)
    if json_raw==nil {
        fmt.Printf("No data\n")
    } else {
        defer json_raw.Close()
        var json_data = DataBase{}
        bytes, err := ioutil.ReadAll(json_raw)
        if err!=nil {
            fmt.Printf("Error: %s\n", err)
        } else {
            json.Unmarshal([]byte(bytes), &json_data)
            return &json_data
        }
    }
    return nil
}

func Save(data DataBase, file string) {
    var json_text []byte
    var err error
    if app_data.indent_file {
        json_text, err = json.MarshalIndent(data, "", "    ")
    } else {
        json_text, err = json.Marshal(data)
    }

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

func contains(arr []string, str string) bool {
   for _, a := range arr {
      if a == str {
         return true
      }
   }
   return false
}

//util method to find the length of the 'first' column
func data_length() int {
    length := -1
    for _ , v := range app_data.data.Columns {
        length = len(v)
        break
    }
    return length
}

func interface_to_float(raw interface{}) float64 {
    ret := 0.0
    switch i := raw.(type) {
        case float64:
            ret = float64(i)
        case float32:
            ret = float64(i)
        case int64:
            ret = float64(i)
    }
    return ret
}

/******************************************************************************/
// #mark Commands

func List(data DataBase) {
    fmt.Printf("List: ")
    for k,v := range data.Columns {
        fmt.Printf("%s=%+v ", k, v)
    }
    fmt.Printf("\n")
}

func CreateColumn(column string) {
    size := data_length()
    app_data.data.Columns[column] = make( []interface{}, 0)
    for i:=0 ; i<size; i++ {
        app_data.data.Columns[column]=append(app_data.data.Columns[column],0.0)
    }
}

// add a row to all columns.
func Create() {
    data := app_data.data.Columns
    for k,v := range data {
        app_data.data.Columns[k] = append(v, 0.0)
    }
}

//read a specific value from the column table
func Read(key string, row int) {
    fmt.Printf("%s[%d]=%+v\n", key, row, app_data.data.Columns[key][row])
}

//update a specific value from the column table
func Update(key string, row int, value string) {
    //if value can be turned into a number, then stuf it as a number
    number, err := strconv.ParseFloat(value, 64)
    if err==nil {
        app_data.data.Columns[key][row] = number
    } else {
        app_data.data.Columns[key][row] = value
    }
}

//delete a row from all columns
func Delete(row int) {
    for k,v := range app_data.data.Columns {
        copy( v[row:], v[row+1:] )
        v[len(v)-1] = ""
        v = v[:len(v)-1]
        app_data.data.Columns[k] = v
    }
}

//Dump table of all columns
//* @param form name of the form to dump out, empty for entire table
func Table(form string) {
    var header bytes.Buffer
    var rows []bytes.Buffer

    first := true

    keys := make([]string, 0, len(app_data.data.Columns))
    for k := range app_data.data.Columns {
        keys = append(keys, k)
    }
    if 0<len(form) {
        keys = app_data.data.Forms[form]
    }
    
    for _,k := range keys {
        v := app_data.data.Columns[k]
        //if v==nil {
        //    v = app_data.data.Calculations[k]
        //}
        if !first {
            header.WriteString( "," )
        }
        header.WriteString( k )

        for i := range v {
            if first {
                rows = append( rows, bytes.Buffer{})
            } else {
                rows[i].WriteString( "," )
            }
            column := ""
            if i<len(v) {
                column = fmt.Sprintf("%v", v[i])
            }
            rows[i].WriteString( column )
        }
        first = false
    }
    fmt.Printf("%v\n", string(header.Bytes()))
    for i := range rows {
        fmt.Printf("%v\n", string(rows[i].Bytes()))
    }
}

// Summaries a form by printing out a table, first row is header, last row is
// summary row. Each column is represented on the summary row based on data
// example: sum main avg,avg
// * @param form name of form to summarize
// * @param args dash delimitated list of summarize functions
func Summary(form string, args string) {
    if 0<len(form) {
        fmt.Printf("sumarize form %s with %s\n", form, args)
        Table(form)
        
        alist := strings.Split(args, ",")
        for i,v := range alist {
            if i<len(app_data.data.Forms) {
                field := app_data.data.Forms[form][i]
                data := app_data.data.Columns[field]
                if 0<i {
                    fmt.Printf(",")
                }
                switch v {
                case "avg":
                    total := 0.0
                    count := 0
                    average := 0.0
                    for _, value := range data {
                        total = total + interface_to_float(value)
                        count = count + 1
                    }
                    average = total / float64(count)
                    fmt.Printf("%f", average)
                case "sum":
                    total := 0.0
                    for _,value := range data {
                        total = total + interface_to_float(value)
                    }
                    fmt.Printf("%f", total)
                case "count":
                    count := len(data)
                    fmt.Printf("%d", count)
                }
            }
        }
        fmt.Printf("\n")
    }
}

func Dash(args []string) {
    if len(args)<1 {
        fmt.Printf("----\n")
    } else if len(args)==1 {
        if 0==len(args[0]) { args[0] = "----" }
        fmt.Printf("%s\n", args[0])
    } else {
        letter := args[0]
        count, err := strconv.Atoi(args[1])
        if err==nil {
            fmt.Printf("%s\n", strings.Repeat(letter, count) )
        } else {
            fmt.Printf("%s\n", letter)
        }
    }
}

func Nop() {
    fmt.Printf("not implemented yet\n")
}

func Help() {
    format := "%4s %-7s %-12s %-40s\n"
    forty := strings.Repeat("-",40)
    fmt.Printf(format, "Flag", "Long", "Arguments", "Description")
    fmt.Printf(format,"----","-------","------------",forty)
    fmt.Printf(format, "c", "create", "", "create a row in each column")
    fmt.Printf(format, "r", "read", "col row", "read a column row")
    fmt.Printf(format, "u", "update", "col row val", "update a column row")
    fmt.Printf(format, "d", "delete", "row", "delete a row from each column")
    fmt.Printf("\n")

    fmt.Printf(format, "t", "table", "", "quit interactive mode")
    fmt.Printf(format, "sum", "summary", "form list",
        "sumarize a form with function list: avg,sum,count")
    fmt.Printf("\n")

    fmt.Printf(format, "q", "quit", "", "quit interactive mode")
    fmt.Printf(format, "", "exit", "", "quit interactive mode")
    fmt.Printf(format, "h", "help", "", "this output")
    fmt.Printf(format, "e", "echo", "string", "echo out something")
    fmt.Printf(format, "-", "----", "sep count", "print out a separator")
    fmt.Printf(format, "s", "save", "", "save database to file")

}

func value(form string, column int, row int) string {
    form_data := app_data.data.Forms[form]
    column_name := form_data[column]
    cell_data := app_data.data.Columns[column_name]
    value := ""
    if cell_data == nil {
        value = "calc"//app_data.data.Calculation[
    } else {
        p := fmt.Sprintf("%f", interface_to_float( cell_data[row] ) )
        value = p
    }
    return value
}

//test code
func Sub(form string) {
    //build a grid, but how big?
    column_count := len( app_data.data.Forms[form] )
    row_count := data_length()

    fmt.Printf("Form: %s - %dx%d\n", form, column_count, row_count)
    
    //make a blank grid
    grid := make( [][]string, 0 )
    for r:=0; r<row_count; r++ {
        tmp := make( []string, 0 )
        for c:=0; c<column_count; c++ {
            tmp = append( tmp, value(form, c, r) )
        }
        grid = append( grid, tmp )
    }

    //print out the grid
    for i,_ := range grid { //rows
        if i==0 { //first line, print header
            for ii,vv := range app_data.data.Forms[form] {
                if ii==0 {
                    fmt.Printf("| %10s |", vv)
                } else {
                    fmt.Printf(" %10s |", vv)
                }
            }
            fmt.Printf("\n")
            for ii,_ := range app_data.data.Forms[form] {
                if ii==0 {
                    fmt.Printf("| %10s |", strings.Repeat("-", 10) )
                } else {
                    fmt.Printf(" %10s |", strings.Repeat("-", 10) )
                }
            }
            fmt.Printf("\n")
        }
        for ii,vv := range grid[i] {
            if ii==0 {
                fmt.Printf("| %10s |", vv)
            } else {
                fmt.Printf(" %10s |", vv)
            }
        }
        fmt.Printf("\n")
    }

}

//create a sample database with 3x2 columns and rows, 2 forms, one setting
func Initialize() {
    data := DataBase{}
    
    data.Columns = make( map[string][]interface{} )
    data.Columns["foo"] = make( []interface{}, 2 )
    data.Columns["foo"] = []interface{}{1.0,2.0}
    data.Columns["bar"] = make( []interface{}, 2 )
    data.Columns["bar"] = []interface{}{3.0,4.0}
    data.Columns["rab"] = make( []interface{}, 2 )
    data.Columns["rab"] = []interface{}{5.0,6.0}

    data.Forms = make( map[string][]string )
    data.Forms["main"] = []string{"foo","bar","foobar"}
    data.Forms["alt"] = []string{"bar","rab","foobar"}

    data.Calculations = make ( map[string]string )
    data.Calculations["foobar"] = "foo bar +"

    data.Settings = make ( map[string]string )
    data.Settings["author"] = "thomas.cherry@gmail.com"
    data.Settings["main.summary"] = "avg,sum"
    data.Settings["alt.summary"] = "sum,avg"

    fmt.Printf("the database is %+v\n", data)

    file := "data.json"

    json_text, err := json.MarshalIndent(data, "", "    ")
    //v("here: %s\n", json_text)
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

/******************************************************************************/
// #mark - application functions

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
print out an esc control
@param esc control code to print out
*/
func PrintCtrOnErr(esc string) {
    fmt.Fprintf(os.Stderr, "\033[%s", esc)
}

/**
print out an esc control
@param esc control code to print out
*/
func PrintCtrOnOut(esc string) {
    fmt.Fprintf(os.Stdout, "\033[%s", esc)
}

/**
run the interactive mode using the third party readline library. Help the 
library stor history, take each line and send it to ProcessLine()
*/
func InteractiveAdvance(line *liner.State, data DataBase) {
    fmt.Printf("Database by thomas.cherry@gmail.com\n")
    app_data.running = true
    for app_data.running==true {
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
    PrintCtrOnOut(ESC_CURSOR_ON)
    v("\ndone\n")
}

//Process a line with a command and arguments
// * @param raw line to posible execute
// * @param data database to operate on
func ProcessLine(raw string, data DataBase) {
    list := strings.Split(raw, " ")
    command := list[0]
    args := []string{""}
    if len(list)>1 {
        args = list[1:]
    }
    switch command {
        case "h", "help":
            Help()
        case "q", "quit", "exit":
            if app_data.verbose { fmt.Printf("getting out of here\n") }
            app_data.running = false
            //os.Exit(0)
        case "e", "echo":
            fmt.Printf("%s => %s\n", command, strings.Join(args, ",") )
        case "-", "----":
            Dash(args)

        /**************************************************************/
        /* CRUD */
        case "c", "create":     //create ; add row to all columns
            if len(args[0])==0 {
                Create()
            } else {            //create column
                CreateColumn(args[0])
            }
        case "r", "read":       //read column row
            column := args[0]
            row, err := strconv.Atoi(args[1])
            if err==nil {
                Read( column, row )
            }
        case "u", "update":     //update column row value
            column := args[0]
            row, row_err := strconv.Atoi(args[1])
            value := args[2]
            if row_err==nil {
                Update( column, row, value)
            }
        case "d", "delete":     //delete row
            row, err := strconv.Atoi(args[0])
            if err==nil {
                Delete(row)
            } else {            //delete column
                Nop() //TODO: add way to delete column
            }
        case "t", "table":
            Table(args[0])
        case "sum", "summary":
            if len(args)>=2 {
                Summary(args[0], args[1])
            } else {
                fmt.Printf("here with %s.\n", args)
            }
        case "initialize":
            Initialize()
        case "l", "list":
            List(data)
        case "sub":
            Sub(args[0]) //- test function

        case "f", "forms":
            fmt.Printf("%+v\n", app_data.data.Forms)//TODO: make this pretty
        case "cf", "create-form":
            Nop()
        case "df", "delete-form":
            Nop()
        
        case "cs", "calcs":
            Nop()
        case "cc", "create-calc":
            Nop()
        case "rc", "read-calc":
            Nop()
        case "uc", "update-calc":
            Nop()
        case "dc", "delete-calc":


        case "s", "save":
            Save(data, app_data.active_file)
    }
}

// #mark
func main() {
    verbose := flag.Bool("verbose", false, "verbose")
    file_name := flag.String("file", "data.json", "data file")
    init_command := flag.String("command", "", "Run one command and exit")
    flag.Parse()

    app_data.verbose = *verbose
    app_data.active_file = *file_name
    data := Load(app_data.active_file)

    if data == nil {
        fmt.Printf("Could not load data\n")
        os.Exit(1)
    } else {
        app_data.data = *data
    }
    if 0<len(*init_command) {
        commands := strings.Split(*init_command, ";")
        for c := range commands {
            ProcessLine(commands[c], app_data.data)
        }
    } else {
        //readline setup
        line := liner.NewLiner()
        defer line.Close()
        setup_liner(line)

        //h := int(getHeight())
        //w := int(getWidth())

        if app_data.verbose { List(app_data.data) }
        InteractiveAdvance(line, app_data.data)
        if app_data.verbose { List(app_data.data) }
    }
}
