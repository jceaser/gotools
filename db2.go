package main

import ("fmt"
    //"bufio"
    "os"
    /*"io" */
    "bytes"
    /* "log"*/
    "sort"
    "flag"
    "io/ioutil"
    "math"
    "strconv"
    "os/exec"
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
    rpn string
    //column_cache map[string][]interface{}
    column_cache map[string][]float64

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
    indent_file:true,
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

const (
    ERR_MSG_COL_NOT_FOUND = "Column %s not found\n"
    ERR_MSG_ROW_BETWEEN = "Row must be between 0 and %d\n"
    ERR_MSG_VALUE_NUM = "Value '%s' is not a number\n"
    ERR_MSG_CREATE_ARGS = "create <column_name>? - optional\n"
    ERR_MSG_READ_ARGS = "read <column_name> <row>\n"
    ERR_MSG_UPDATE_ARGS = "update <column_name> <row> <value>\n"
    ERR_MSG_DELETE_ARGS = "delete <row>\n"
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

//rpn -formula '2 3 +' -pop
func run(formula string) string {
    //fmt.Printf("%s\n", formula)
    out, err := exec.Command(app_data.rpn, "-formula", formula, "-pop").Output()
    if err != nil {
        fmt.Printf("%s", err)
    }
    output := strings.TrimSpace(string(out[:]))
    ret := output
    return ret
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

    if len(file)<1 {
        file = app_data.active_file
    }

    if err!=nil {
        fmt.Printf("error: %s - %s\n", file, err)
        return
    }
    err = ioutil.WriteFile(file, json_text, 0644)
    if err!=nil {
        fmt.Printf("Error: %s - %s\n", file, err)
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

/** Dump out a list of columns with their rows */
func List(data DataBase) {
    fmt.Printf("List: ")
    for k,v := range data.Columns {
        fmt.Printf("%s=%+v ", k, v)
    }
    for k,v := range app_data.column_cache {
        fmt.Printf("%s=%+v ", k, v)
    }
    fmt.Printf("\n")
}

/* Create a new column , called by c command with an argument */
func CreateColumn(column string) {
    size := data_length()
    app_data.data.Columns[column] = make( []interface{}, 0)
    for i:=0 ; i<size; i++ {
        app_data.data.Columns[column]=append(app_data.data.Columns[column],0.0)
    }
}

// add a row to all columns. called by c command with no arguments
func Create() {
    data := app_data.data.Columns
    for k,v := range data {
        app_data.data.Columns[k] = append(v, 0.0)
    }
}

//read a specific value from the column table ; called with read command
func Read(key string, row int) {
    if app_data.data.Columns[key]==nil {
        fmt.Fprintf(os.Stderr, ERR_MSG_COL_NOT_FOUND, key)
    } else {
        max := len(app_data.data.Columns[key])-1
        if max<row || row<0 {
            fmt.Fprintf(os.Stderr, ERR_MSG_ROW_BETWEEN, max)
        } else {
            data := app_data.data.Columns[key][row]
            fmt.Printf("%s[%d]=%+v\n", key, row, data)
        }
    }
}

//update a specific value from the column table
func Update(key string, row int, value string) {
    //if value can be turned into a number, then stuf it as a number
    number, err := strconv.ParseFloat(value, 64)
    if err==nil {
        //no error, value is a number
        column := app_data.data.Columns[key]
        if column == nil {
            fmt.Fprintf(os.Stderr, ERR_MSG_COL_NOT_FOUND, key)
        } else {
            max := len(column)-1
            if max<row || row<0 {
                fmt.Fprintf(os.Stderr, ERR_MSG_ROW_BETWEEN, max)
            } else {
                app_data.data.Columns[key][row] = number
            }
        }
    } else {
        //use as is, really do this?
        //app_data.data.Columns[key][row] = value
        fmt.Fprintf(os.Stderr, ERR_MSG_VALUE_NUM, value)
    }
}

//delete a row from all columns
func Delete(row int) {
    for k,v := range app_data.data.Columns {
        max := len(v)-1
        //while we have the first column, check the length before going on
        if max<row || row<0 {
            fmt.Fprintf(os.Stderr, ERR_MSG_ROW_BETWEEN, max)
            break
        } else {
            copy( v[row:], v[row+1:] )
            v[len(v)-1] = ""
            v = v[:len(v)-1]
            app_data.data.Columns[k] = v
        }
    }
}

func DeleteColumn(column string) {
    Nop()
}

func put_cache(key string, data []float64) {
    if app_data.column_cache==nil {
        app_data.column_cache = make(map[string][]float64)
    }
    app_data.column_cache[key] = data
}

func get_cache(key string) []float64 {
    if app_data.column_cache==nil {
        app_data.column_cache = make(map[string][]float64)
    }
    data := app_data.column_cache[key]
    
    return data
}

func Calculate() {
    var header bytes.Buffer
    var rows []bytes.Buffer
    first := true

    //find the first column and get its length, then initialize rows
    for _,v := range app_data.data.Columns {
        for i:=0 ; i<len(v) ; i++ {
            rows = append(rows, bytes.Buffer{})
        }
        break
    }

    for key,formula := range app_data.data.Calculations {
        if !first {
            header.WriteString( "," )
        }
        header_title := fmt.Sprintf("%s='%v'", key, formula )
        header.WriteString( header_title )

        var calc_values []float64
        for i,_ := range rows {
            if !first {
                rows[i].WriteString( "," )
            }
            result := formula_for_row(formula, i)
            result_as_float, _ := strconv.ParseFloat(result, 64)
            calc_values = append(calc_values, result_as_float)
            rows[i].WriteString( result )
        }
        put_cache(key, calc_values)
        first = false
    }
    fmt.Printf("%v\n", string(header.Bytes()))
    for i := range rows {
        fmt.Printf("%v\n", string(rows[i].Bytes()))
    }
}

func FormFiller(form string, action string) {
    dry_run := false
    if action=="dry-run" {
        dry_run = true
    }
    //var header bytes.Buffer
    //var row []bytes.Buffer
    line := liner.NewLiner()
    defer line.Close()
    if !dry_run {
        Create() //new row
    }
    row := data_length() - 1
    //values := make([]float64, 0, data_length())
    //index := 0
    for _,column := range app_data.data.Forms[form] {
        if 0<len(app_data.data.Calculations[column]) {
            continue // this is a calculation, skip it
        }
        asking := true
        answer := 0.0
        for asking {
            fmt.Printf("Enter in a number for %s\n", column)
            raw_response, _ := line.Prompt("#")
            if raw_response=="stop" {
                return
            }
            number, err := strconv.ParseFloat(raw_response, 64)
            if err!=nil {
                fmt.Printf("that was not a number. Try again.\n")
            } else {
                answer = number
                asking = false
            }
        }
        //values[index] = answer
        //index++
        if !dry_run {
            app_data.data.Columns[column][row] = answer
        }
    }
    //fmt.Printf("all answers %v\n", values)
}

//Dump table of all columns
//* @param form name of the form to dump out, empty for entire table
func Table(form string) {
    var header bytes.Buffer
    var rows []bytes.Buffer
    
    first := true
    keys := make([]string, 0, len(app_data.data.Columns))

    if 0<len(form) {
        //use the form list
        keys = app_data.data.Forms[form]
    } else {
        //use all columns
        for k := range app_data.data.Columns {
            keys = append(keys, k)
        }
    }
    
    //loop throug all the column or form keys
    for _,k := range keys {
        var formula = ""
        v := app_data.data.Columns[k] //return a list of strings

        // if v is nil, then not a column, search calculations
        if v==nil {
            formula = app_data.data.Calculations[k]
            if formula !="" {
                var calc_values []float64
                for i,_ := range rows {
                    if first {
                        rows = append(rows, bytes.Buffer{})
                    }
                    result := formula_for_row(formula, i)
                    result_as_float, _ := strconv.ParseFloat(result, 64)
                    calc_values = append(calc_values, result_as_float)
                    rows[i].WriteString( "," )
                    rows[i].WriteString( result )
                }
                put_cache(k, calc_values)
            } else {
                //blank key?
                continue
            }
        }
        if !first {
            header.WriteString( "," )
        }
        header.WriteString( k )

        for i := range v {
            //should this section be outside of the loop
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

/**
convert a formula to a value
@param formula calculation to make $c1 $c2 +
@param row 0 based row count
@return result
*/
func formula_for_row(formula string, row int) string {
    words := strings.Split(formula, " ")
    for i,v := range words {
        if strings.HasPrefix(v, "$") {
            key := v[1:]
            columns := app_data.data.Columns[key]
            if columns!=nil {
                column := fmt.Sprintf("%f",columns[row])
                words[i] = column
            }
        }
    }
    ret := strings.Join(words, " ")
    ret = run(ret)
    return ret
}

// Summaries a form by printing out a table, first row is header, last row is
// summary row. Each column is represented on the summary row based on data
// example: sum main avg,avg
// * @param form name of form to summarize
// * @param args dash delimitated list of summarize functions
func Summary(form string, args string) {
    var out bytes.Buffer
    if 0<len(form) {
        v("sumarize form %s with %s\n", form, args)
        Table(form)
        first_form := app_data.data.Forms[form][0]
        alist := strings.Split(args, ",")
        for i,v := range alist {
            if i<=len(app_data.data.Forms) {
                field := app_data.data.Forms[form][i]
                data := app_data.data.Columns[field]
                if data == nil {
                    /*
                    there is no column data, so try getting calculated values 
                    from the cache. Table caches it's last calculations for 
                    functions like summary to build on
                    */
                    row_count := len(app_data.data.Columns[first_form])
                    data = make([]interface{}, row_count)
                    raw := get_cache(field)
                    for i,v := range raw {
                        data[i] = v
                    }
                }
                if 0<i {
                    out.WriteString( "," )
                }

                switch v {
                    case "a", "avg":
                        out.WriteString( fmt.Sprintf("%f", Average(data) ) )
                    case "c", "count":
                        out.WriteString( fmt.Sprintf("%d", len(data)) )
                    case "h", "har", "harmonic":
                        out.WriteString( fmt.Sprintf("%f", Harmonic(data)) )
                    case "mx", "max":
                        out.WriteString( fmt.Sprintf("%f", Max(data)) )
                    case "m", "medium":
                        out.WriteString( fmt.Sprintf("%f", Median(data)) )
                    case "md", "mode":
                        out.WriteString( fmt.Sprintf("%f", Mode(data)) )
                    case "mn", "min":
                        out.WriteString( fmt.Sprintf("%f", Min(data)) )
                    case "", "n", "none":
                        out.WriteString("")
                    case "s", "sum":
                        out.WriteString( fmt.Sprintf("%f", Sum(data)) )
                    case "sd", "sdev":
                        sd := StandardDeviation(data)
                        out.WriteString( fmt.Sprintf("%f", sd) )
                }
            }
        }
        fmt.Printf( "%v\n", string(out.Bytes()) )
    }
}

func Average(data []interface{}) float64 {
    total := 0.0
    count := 0
    average := 0.0
    for _, value := range data {
        total = total + interface_to_float(value)
        count = count + 1
    }
    average = total / float64(count)
    return average
}

func Harmonic(data []interface{}) float64 {
    total := 0.0
    count := 0
    harmonic := 0.0
    for _, value := range data {
        total = total + ( 1.0 / interface_to_float(value) )
        count = count + 1
    }
    harmonic = float64(count) / total
    return harmonic
}


func StandardDeviation (data []interface{}) float64 {
    var sum, mean, sd float64 = 0, 0, 0
    count_i := len(data)
    count_f := float64(count_i)

    for i:=0 ; i<count_i; i++ {
        sum += interface_to_float(data[i])
    }
    mean = sum / count_f
    for i:=0 ; i<count_i; i++ {
        sd += math.Pow( interface_to_float(data[i])-mean, 2)
    }
    sd = math.Sqrt( sd / count_f)
    return sd
}

func Sum(data []interface{}) float64 {
    total := 0.0
    for _,value := range data {
        total = total + interface_to_float(value)
    }
    return total
}

func Max(data []interface{}) float64 {
    max := math.SmallestNonzeroFloat64
    for _,value := range data {
        max = math.Max(max, interface_to_float(value))
    }
    return max
}

func Min(data []interface{}) float64 {
    min := math.MaxFloat64
    for _,value := range data {
        min = math.Min(min, interface_to_float(value))
    }
    return min
}

func Median(data []interface{}) float64 {

    sort.Slice(data, func(i, j int) bool {
        return interface_to_float(data[i]) < interface_to_float(data[j])
    })

    len_of_data := float64(len(data))
    ret := 0.0
    if math.Mod(len_of_data, 2) == 0.0 { //even number
        index := int((math.Floor(len_of_data) / 2) - 1.0)
        left := interface_to_float(data[index])
        right := interface_to_float(data[index+1])
        ret = ( left + right ) / 2.0
    } else { //odd number
        index := int(math.Floor(len_of_data / 2))
        ret = interface_to_float(data[index])
    }
    return ret
}

func Mode(data []interface{}) float64 {
    sort.Slice(data, func(i, j int) bool {
        return interface_to_float(data[i]) < interface_to_float(data[j])
    })
    hash := make( map[float64]int )
    //collect counts
    for _, v := range data {
        value := interface_to_float(v)
        existing := hash[value]
        hash[value] = existing+1
    }
    selected := 0.0
    count := 1  //assume at least two values to bump off the default
    for k, v := range hash {
        if count<v {
            selected = k
            count = v
        }
    }
    return selected
}

/*
outputs a number of characters to visually separate out the output
@param arg 1 if empty, output '----'
@param arg 1 if not empty, output that character
@param arg 2 if number, output arg[1] this many times
*/
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

/* do nothing */
func Nop() {
    fmt.Printf("not implemented yet\n")
}

/* print out a help method */
func Help() {
    fmt.Printf("Database by thomas.cherry@gmail.com\n")
    fmt.Printf("Manage table data with optional form display.\n")
    fmt.Printf("\nNote: Arguments with ? are optional\n\n")
    format := "%4s %-7s %-12s %-40s\n"
    forty := strings.Repeat("-",40)
    fmt.Printf(format, "Flag", "Long", "Arguments", "Description")
    fmt.Printf(format,"----","-------","------------",forty)
    fmt.Printf(format, "c", "create", "name?",
        "create a row in each column, or a new named column")
    fmt.Printf(format, "r", "read", "col row", "read a column row")
    fmt.Printf(format, "u", "update", "col row val", "update a column row")
    fmt.Printf(format, "d", "delete", "row", "delete a row from each column")
    fmt.Printf("\n")

    fmt.Printf(format, "t", "table", "form?",
        "display a table, optionally as a form")
    fmt.Printf(format, "sum", "summary", "form list",
        "sumarize a form with function list: avg,count,max,min,medium,mode,min,sum,sdev")
    fmt.Printf(format, "l", "ls list", "", "")
    fmt.Printf("\n")

    fmt.Printf(format, "s", "save", "", "save database to file")
    fmt.Printf(format, "q", "quit", "", "quit interactive mode")
    fmt.Printf(format, "", "exit", "", "quit interactive mode")
    fmt.Printf(format, "h", "help", "", "this output")
    fmt.Printf(format, "e", "echo", "string", "echo out something")
    fmt.Printf(format, "-", "----", "sep count", "print out a separator")
    fmt.Printf(format, "", "file", "name?", "set or print current file name")
    fmt.Printf(format, "", "rpn", "path?", "set or print current rpn command")
    fmt.Printf(format, "", "verbose", "", "toggle verbose mode")
}

/** used by Sub only */
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
func Initialize(file_name string) {
    data := DataBase{}
    
    data.Columns = make( map[string][]interface{} )
    data.Columns["foo"] = make( []interface{}, 2 )
    data.Columns["foo"] = []interface{}{0.0,1.0,2.0}
    data.Columns["bar"] = make( []interface{}, 2 )
    data.Columns["bar"] = []interface{}{3.0,4.0,3.0}
    data.Columns["rab"] = make( []interface{}, 2 )
    data.Columns["rab"] = []interface{}{5.0,6.0,6.0}

    data.Forms = make( map[string][]string )
    data.Forms["main"] = []string{"foo","bar","foobar"}
    data.Forms["alt"] = []string{"bar","rab","foobar"}

    data.Calculations = make ( map[string]string )
    data.Calculations["foobar"] = "$foo $bar +"

    data.Settings = make ( map[string]string )
    data.Settings["author"] = "thomas.cherry@gmail.com"
    data.Settings["main.summary"] = "avg,sum,avg"
    data.Settings["alt.summary"] = "sum,avg,sum"

    fmt.Printf("the database is %+v\n", data)

    //file := "data.json"
    file := file_name
    if len(file_name)<1 {
        file = app_data.active_file
    }

    var json_text []byte
    var err error
    if app_data.indent_file {
        json_text, err = json.MarshalIndent(data, "", "    ")
    } else {
        json_text, err = json.Marshal(data)
    }

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
        case "verbose":
            app_data.verbose = !app_data.verbose
            v("Verbose is %s\n", "on")
        case "file":
            if 0<len(args) && 0<len(args[0]) {
                //set mode
                app_data.active_file = args[0]
            } else {
                fmt.Printf("Active file: '%s'.\n", app_data.active_file) 
            }
        case "rpn":
            if 0<len(args) && 0<len(args[0]) {
                app_data.rpn = args[0]
            } else {
                fmt.Printf("RPN command: %s\n", app_data.rpn)
            }

        /**************************************************************/
        /* CRUD */
        case "c", "create":     //create ; add row to all columns
            if 1<len(args) {
                fmt.Fprintf(os.Stderr, ERR_MSG_CREATE_ARGS)
            } else {
                if len(args[0])==0 {
                    Create()
                } else {            //create column
                    CreateColumn(args[0])
                }
            }
        case "r", "read":       //read column row
            if len(args)!=2 {
                fmt.Fprintf(os.Stderr, ERR_MSG_READ_ARGS)
            } else {
                column := args[0]
                row, err := strconv.Atoi(args[1])
                if err==nil {
                    Read( column, row )
                }
            }
        case "u", "update":     //update column row value
            if len(args)!=3 {
                fmt.Fprintf(os.Stderr, ERR_MSG_UPDATE_ARGS)
            } else {
                column := args[0]
                row, row_err := strconv.Atoi(args[1])
                value := args[2]
                if row_err==nil {
                    Update( column, row, value)
                }
            }
        case "d", "delete":     //delete row
            if len(args)!=1 || len(args[0])<1{
                fmt.Fprintf(os.Stderr, ERR_MSG_DELETE_ARGS)
            } else {
                row, err := strconv.Atoi(args[0])
                if err==nil {
                    Delete(row)
                } else {            //delete column
                    DeleteColumn(args[0]) //TODO: add way to delete column
                }
            }
        case "ff", "form":
            FormFiller(args[0], "create")
        case "t", "table":
            Table(args[0])
        case "sum", "summary":
            if len(args)>=2 {
                Summary(args[0], args[1])
            } else {
                fmt.Printf("here with %s.\n", args)
            }
        case "calc", "calculate":
            Calculate()
        case "init", "initialize":
            file := app_data.active_file
            if len(args)==1 || 0<len(args[0]) {
                file = args[0]
            }
            Initialize(file)
        case "l", "ls", "list":
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
            file := app_data.active_file
            if len(args)==1 || 0<len(args[0]) {
                file = args[0]
            }
            Save(data, file)
    }
}

// #mark
func main() {
    verbose := flag.Bool("verbose", false, "verbose")
    file_name := flag.String("file", "data.json", "data file")
    init_command := flag.String("command", "", "Run one command and exit")
    rpn_command := flag.String("rpn", "rpn", "command to process calculations")
    flag.Parse()

    app_data.verbose = *verbose
    app_data.active_file = *file_name
    app_data.rpn = *rpn_command
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
