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

type Format struct {
    markdown bool
    divider string
    divider_pipe string
    template_float string
    template_string string
    template_decimal string
}

type App_Data struct {
    backlog_command string
    worker_command string
    backlog_list []string
    rpn string
    column_cache map[string][]interface{}

    data DataBase
    verbose bool
    indent_file bool
    active_file string
    running bool
    format Format
    sort bool
}

var (
    history_fn = filepath.Join(os.TempDir(), ".rpn_history")    //used by liner
    names      = []string{"Create", "Read", "Update", "Delete"} //used by liner
)

var buffers = screen_buffers{left_hud: "", right_hud: "", content: ""}
var app_data = App_Data{backlog_command:"",
    worker_command:"",
    indent_file:true,
    verbose:false,
    format: Format{template_float:"%10.3f",
        template_string:"%10s",
        template_decimal:"%10d",
        markdown:false,
        divider:"│",
        divider_pipe:"|"},
    sort:true}

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
    ERR_MSG_FORM_REQUIRED = "A form name is required\n"
    ERR_MSG_FORM_EXISTS = "There already exists a form named '%s'.\n"
    ERR_MSG_FORM_NOT_EXISTS = "There is no form named '%s'.\n"
    ERR_MSG_FORM_create = "form-create <name> [list]\n"
    ERR_MSG_FORM_UPDATE = "form-update <name> [list]\n"
    ERR_MSG_FORM_RENAME = "form-rename <name> <src> <dest>\n"
)

// #mark - utility functions

/** print only in verbose mode */
func v(format string, args ...string) {
    if app_data.verbose {
        fmt.Printf(format, args)
    }
}

/** print to error */
func e(format string, args ...string) {
    fmt.Fprintf(os.Stderr, format, args)
}

/** print to error but only in verbose mode */
func ev(format string, args ...string) {
    if app_data.verbose {
        fmt.Fprintf(os.Stderr, format, args)
    }
}

/** print to a specific terminal screen location */
func PrintStrAt(msg string, y, x int) {
    fmt.Printf("\033[%d;%dH%s", y, x, msg)
}

/** print to a terminal code */
func PrintCtrAt(esc string, y, x int) {
    fmt.Printf("\033[%d;%dH\033[%s", y, x, esc)
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

/** Save the screen setup at the start of the app */
func ScrSave() {
    PrintCtrOnErr(ESC_SAVE_SCREEN)
    PrintCtrOnErr(ESC_SAVE_CURSOR)
    //PrintCtrOnErr(ESC_CURSOR_OFF)
    PrintCtrOnErr(ESC_CLEAR_SCREEN)
}

/** print a string in a color */
func ColorText(text string, color int) string {
    encoded := fmt.Sprintf("\033[0;%dm%s\033[0m", color, text)
    return encoded
}

func Red(text string) string {
    return ColorText(text, 31)
}

func Green(text string) string {
    return ColorText(text, 32)
}

func Blue(text string) string {
    return ColorText(text, 34)
}

/** Restore the screen setup from SrcSave() */
func ScrRestore() {
    //PrintCtrOnErr(ESC_CURSOR_ON)
    PrintCtrOnErr(ESC_RESTORE_CURSOR)
    PrintCtrOnErr(ESC_RESTORE_SCREEN)
}

/**
Run the external 'rpn' command
rpn -formula '2 3 +' -pop
*/
func run(formula string) string {
    ev("Calling the command: '%s'.\n", formula)
    out, err := exec.Command(app_data.rpn, "-formula", formula, "-pop").Output()
    if err != nil {
        fmt.Printf("%s", err)
    }
    output := strings.TrimSpace(string(out[:]))
    ret := output
    return ret
}

/**************************************/
/* manage load and unload database functions */

/** set the internal data value, use this to setup tests */
func SetData(data DataBase) {
    app_data.data = data
}

/** load a json data base file */
func load(file string) *os.File {
    json_raw, err := os.Open(file)
    if err!=nil {
        if os.IsNotExist(err) {
            //create the file because it does not exist
            v("Creating data file %s\n", file)
            //todo: this looks like the wrong default
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

/** load a database from a file */
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

/** save the database to a file */
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

/** print out json */
func DumpJson() {
    var json_text []byte
    var err error
    if app_data.indent_file {
        json_text, err = json.MarshalIndent(app_data.data, "", "    ")
    } else {
        json_text, err = json.Marshal(app_data.data)
    }
    if err==nil {
        fmt.Printf("%s\n", json_text)
    } else {
        fmt.Printf("Error: %s\n", err)
    }
}

/* database functions
/**************************************/
/* helpers */

/** look for an option in an array and if not found use the fallback */
func arg (args []string, index int, fallback string) string {
    //[a, b, c, d] ; len=4
    //i==3
    ret := fallback
    if index<len(args) {    //request in range
        raw := args[index]
        if 0<len(raw) {
            ret = raw
        }
    }
    return ret
}

/** check if a string is contained in a list of strings */
func contains(arr []string, str string) bool {
   for _, a := range arr {
      if a == str {
         return true
      }
   }
   return false
}

/** return sorted keys from a map of interfaces */
func sorted_keys(data map[string][]interface{}) []string {
    keys := make([]string, len(data))
    i := 0
    for k := range data {
        keys[i] = k
        i++
    }
    sort.Strings(keys)
    return keys
}

/* util method to find the length of the 'first' column */
func data_length() int {
    return DataLength(app_data.data)
}

/* find the length of the 'first' column */
func DataLength(data DataBase) int {
    length := -1
    for _ , v := range data.Columns {
        length = len(v)
        break
    }
    return length
}

func FirstForm(data DataBase) string{
    name := "def"
    for k, _ := range data.Forms {
        name = k
        break
    }
    return name
}

func is_interface_a_string(raw interface{}) bool {
    ret := false
    switch raw.(type) {
        case string:
            ret = true
        default:
            ret = false
    }
    return ret
}

func is_interface_a_number(raw interface{}) bool {
    ret := false
    switch raw.(type) {
        case string:
            ret = false
        case float64:
            ret = true
        case float32:
            ret = true
        case int64:
            ret = true
        case int32:
            ret = true
        case int:
            ret = true
        default:
            ret = false
    }
    return ret
}

func interface_to_string(raw interface{}) string {
    ret := ""
    switch i := raw.(type) {
        case string:
            ret = i
        case float64:
            ret = fmt.Sprintf("%f", i)
        case float32:
            ret = fmt.Sprintf("%f", i)
        case int64:
            ret = fmt.Sprintf("%0.0d", i)
        case int32:
            ret = fmt.Sprintf("%0.0d", i)
        case int:
            ret = fmt.Sprintf("%0.0d", i)
        default:
            fmt.Printf("got here")
    }
    return ret
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

/** cache the calculated results */
func put_cache(key string, data []interface{}) {
    if app_data.column_cache==nil {
        app_data.column_cache = make(map[string][]interface{})
    }
    app_data.column_cache[key] = data
}

/** get cached calculated results */
func get_cache(key string) []interface{} {
    if app_data.column_cache==nil {
        app_data.column_cache = make(map[string][]interface{})
    }
    data := app_data.column_cache[key]
    
    return data
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
        //this allows for the row number to be inserted in as as column
        if strings.HasPrefix(v, "#row") {
            words[i] = fmt.Sprintf("%d",row)
        }
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

func Row(args []string, data DataBase) {
    var header bytes.Buffer
    var body bytes.Buffer

    //row form? row?
    row := 0
    form := FirstForm(data)
    if 0<len(args) {
        form = arg(args, 0, FirstForm(data))
    }
    if 1<len(args) {
        //just a row number
        raw_row, err := strconv.Atoi(arg(args, 1, "0"))
        if err!=nil {
            fmt.Printf("error: %v\n", err)
            return
        } else {
            row = raw_row
        }
    }

    keys := data.Forms[form]

    for i,v := range keys {
        if value, exists := data.Columns[v] ; exists {
            if i!=0 {
                header.WriteString(" ")
                body.WriteString(" ")
            }
            header.WriteString(v)
            body.WriteString(fmt.Sprintf("%f", value[row]))
        }
    }
    fmt.Printf("%s\n%s\n", string(header.Bytes()), string(body.Bytes()))
}

func CreateColumn(data DataBase, column string) {
    size := data_length()
    data.Columns[column] = make( []interface{}, 0)
    for i:=0 ; i<size; i++ {
        data.Columns[column]=append(data.Columns[column],0.0)
    }
}

func Create(data DataBase) {
    column_data := data.Columns
    for k,v := range column_data {
        data.Columns[k] = append(v, 0.0)
    }
}

//read a specific value from the column table ; called with read command
func Read(data DataBase, key string, row int) {
    if data.Columns[key]==nil {
        fmt.Fprintf(os.Stderr, ERR_MSG_COL_NOT_FOUND, key)
    } else {
        max := len(data.Columns[key])-1
        if max<row || row<0 {
            fmt.Fprintf(os.Stderr, ERR_MSG_ROW_BETWEEN, max)
        } else {
            data := data.Columns[key][row]
            fmt.Printf("%s[%d]=%+v\n", key, row, data)
        }
    }
}

//update a specific value from the column table
func Update(data DataBase, key string, row int, value string) {
    column := data.Columns[key]
    if column == nil {
        fmt.Fprintf(os.Stderr, ERR_MSG_COL_NOT_FOUND, key)
    } else {
        max := len(column)-1
        if row<0 || max<row {
            fmt.Fprintf(os.Stderr, ERR_MSG_ROW_BETWEEN, max)
        } else {
            //if value can be turned into a number, then stuff it as a number
            if number, err := strconv.ParseFloat(value, 64) ; err==nil {
                //no error, value is a number
                data.Columns[key][row] = number
            } else {
                data.Columns[key][row] = value
            }
        }
    }
}

//delete a row from all columns
func Delete(data DataBase, row int) {
    for k,v := range data.Columns {
        max := len(v)-1
        //while we have the first column, check the length before going on
        if max<row || row<0 {
            fmt.Fprintf(os.Stderr, ERR_MSG_ROW_BETWEEN, max)
            break
        } else {
            copy( v[row:], v[row+1:] )
            v[len(v)-1] = ""
            v = v[:len(v)-1]
            data.Columns[k] = v
        }
    }
}

func DeleteColumn(data DataBase, column string) {
    delete(data.Columns, column)
}

/** append a new row to the data, and populate the named rows */
func AppendTableByName(data DataBase, args []string) {
    /* format: column_values */
    arg_count := len(args)
    if arg_count<1 {
        return
    }
    Create(data)
    row := DataLength(data)-1
    for i:=0; i<len(args); i++ {
        raw := args[i]
        parts := strings.Split(raw, ":")
        column := parts[0]
        value := parts[1]
        if _, ok := data.Columns[column]; ok {
            data.Columns[column][row] = value
        }
    }
}

/** populate a new row with provided data */
func AppendTable(data DataBase, args []string) {
    /* format: column_values */
    arg_count := len(args)
    column_count := len(data.Columns)
    if arg_count<1 {
        return
    }
    Create(data)
    row := DataLength(data)-1
    index := 0
    for _, column := range sorted_keys(data.Columns) {
        if value, err := strconv.ParseFloat( args[index], 64 ) ; err==nil {
            data.Columns[column][row] = value
        } else {
            data.Columns[column][row] = args[index]
        }

        //prep for next round
        index = index + 1
        if arg_count<=index || column_count<=index {
            break
        }
    }
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
            header.WriteString( app_data.format.divider )
        }
        header_title := fmt.Sprintf("%s='%v'", key, formula )
        header.WriteString( header_title )

        var calc_values []interface{}
        for i,_ := range rows {
            if !first {
                rows[i].WriteString( app_data.format.divider )
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
        Create(app_data.data) //new row
    }
    row := data_length() - 1
    for _,column := range app_data.data.Forms[form] {
        if 0<len(app_data.data.Calculations[column]) {
            continue // this is a calculation, skip it
        }
        asking := true
        answer := 0.0
        for asking {
            fmt.Printf("Enter in a number for column '%s'.\n", column)
            raw_response, _ := line.Prompt("#")
            quiters := []string{"stop","exit","quit"}
            if contains(quiters, raw_response) {
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
        if !dry_run {
            app_data.data.Columns[column][row] = answer
        }
    }
}

/* Use control codes to draw a form the user fills out */
func FormFillerVisual(form string, action string) {
ScrSave()

    //Lines
    LINE_QUESTION := 1
    LINE_INPUT := 2
    LINE_MESSAGE := 3
    LINE_FORM_START := 5
    LINE_FORM_HEAD := LINE_FORM_START - 1

    dry_run := false
    if action=="dry-run" {
        dry_run = true
    }
    line := liner.NewLiner()
    defer line.Close()
    /*if !dry_run {
        Create() //new row
    }
    row := data_length() - 1*/

    temp_data := make(map[string]interface{})

    //setup
    PrintStrAt("", LINE_FORM_HEAD, 1)
    fmt.Print(strings.Repeat(fmt.Sprintf("%c", RuneS3), 80))
    for c_count,column := range app_data.data.Forms[form] {
        if _, ok := app_data.data.Calculations[column] ; ok {
            continue // this is a calculation, skip it
        }
        //answer := app_data.data.Columns[column][row]
        answer := temp_data[column]
        PrintStrAt(fmt.Sprintf("%d: %s = %f.\n", c_count, column, answer), c_count+LINE_FORM_START, 1)
    }
reviewing:=true
for reviewing {

    for c_count,column := range app_data.data.Forms[form] {
        if 0<len(app_data.data.Calculations[column]) {
            continue // this is a calculation, skip it
        }
        asking := true
        answer := 0.0
        for asking {
            PrintStrAt(fmt.Sprintf("Enter in a number for column '%s'.\n", Green(column)), LINE_QUESTION, 1)
            
            PrintCtrAt(ESC_CLEAR_LINE, LINE_INPUT,1)
            PrintStrAt(fmt.Sprintf(""), LINE_INPUT, 1)
            raw_response, _ := line.Prompt("#")
            PrintCtrAt(ESC_CLEAR_LINE, LINE_INPUT,1)
            
            quiters := []string{"stop","exit","quit"}
            if contains(quiters, raw_response) {
ScrRestore()
                return
            }
            number, err := strconv.ParseFloat(raw_response, 64)
            if err!=nil {
                PrintStrAt(fmt.Sprintf(Red("that was not a number. Try again.\n")), LINE_MESSAGE,1)
            } else {
                PrintCtrAt(ESC_CLEAR_LINE, LINE_MESSAGE,1)
                answer = number
                asking = false
            }
        }
        if !dry_run {
            msg := fmt.Sprintf("%d: %s = %f\n", c_count, Green(column), answer)
            PrintStrAt(msg, c_count+LINE_FORM_START, 1)
            var a interface{}
            a = answer
            temp_data[column] = a
            //app_data.data.Columns[column][row] = answer
        }
    }
    
    PrintCtrAt(ESC_CLEAR_LINE, LINE_QUESTION,1)
    raw_response, _ := line.Prompt("done? yes or no: ")
    PrintCtrAt(ESC_CLEAR_LINE, LINE_QUESTION,1)
    
    quiters := []string{"stop","exit","e","done","d","yes", "y", "save"}
    if contains(quiters, raw_response) {
        reviewing = false
        if !dry_run {
            Create(app_data.data) //new row
            row := data_length() - 1
            for k, v := range temp_data {
                //should also create the new row here
                app_data.data.Columns[k][row] = v
            }
        }
    }
}

ScrRestore()

}

/**
level - 0=top, 1=middle, 2=bottom
*/
func table_divider (markdown bool, columns int, level int) string {
    var sbuf bytes.Buffer
    if !markdown {
        if level==0 {
            sbuf.WriteRune(RuneULCorner)
        } else if level==1 {
            sbuf.WriteRune(RuneLTee)
        } else if level==2 {
            sbuf.WriteRune(RuneLLCorner)
        }
    }
    for i:=0 ; i<=columns ; i++ {
        if markdown {
            sbuf.WriteString(app_data.format.divider_pipe)
            sbuf.WriteString(fmt.Sprintf(app_data.format.template_string, "---------"))
        } else {
            if 0<i && i<=columns {
                if level==0 {
                    sbuf.WriteRune(RuneTTee) //'┬'
                } else if level==1 {
                    sbuf.WriteRune(RunePlus) //"┼"
                } else if level==2 {
                    sbuf.WriteRune(RuneBTee) //┴
                }
            }
            sbuf.WriteString(fmt.Sprintf(app_data.format.template_string, "──────────"))
        }
    }
    if markdown {
        sbuf.WriteString(app_data.format.divider_pipe)
    } else {
        if level==0 {
            sbuf.WriteRune(RuneURCorner)
        } else if level==1 {
            sbuf.WriteRune(RuneRTee)
        } else if level==2 {
            sbuf.WriteRune(RuneLRCorner)
        }
    }
    return sbuf.String()
}

//Dump table of all columns
//* @param form name of the form to dump out, empty for entire table
func Table(form string) {
    markdown := app_data.format.markdown
    var divider string
    if markdown {
        divider = app_data.format.divider_pipe
    } else {
        divider = app_data.format.divider
    }
    header, rows, keys := table_worker(form, divider)
    
    //print out the table
    if !markdown {
        fmt.Printf("%s\n", table_divider(markdown, len(keys), 0))
    }
    fmt_lined := (app_data.format.template_string + "\n")
    fmt.Printf( fmt_lined , string(header.Bytes()))
    fmt.Printf("%s\n", table_divider(markdown, len(keys), 1))
    for i := range rows {
        fmt.Printf("%v\n", string(rows[i].Bytes()))
    }
    if !markdown {
        fmt.Printf("%s\n", table_divider(markdown, len(keys), 2))
    }
}

func table_worker(form string, divider string) (bytes.Buffer, []bytes.Buffer, []string) {
    var header bytes.Buffer
    var rows []bytes.Buffer
    var keys []string
    
    for i:=0 ; i<data_length() ; i++ {
       rows = append(rows, bytes.Buffer{})
    }

    first := true

    //figure out which fields need to be displayed
    if 0<len(form) {
        //use the form list
        keys = app_data.data.Forms[form]
    } else {
        //always sort because map is not order consistent
        keys = sorted_keys(app_data.data.Columns)
    }
    
    //loop throug all the column and calculation keys
    max := len(keys)-1
    for index,k := range keys {
        last := false
        if max<=index {
            last = true
        }
        var formula = ""
        values := app_data.data.Columns[k] //return a list of strings

        // if values is nil, then not a column, search calculations
        if values==nil {
            formula = app_data.data.Calculations[k]
            if formula=="" {
                continue //key is blank, skip it
            }
            var calc_values []interface{}
            for i,_ := range rows {
                result := formula_for_row(formula, i)
                result_as_float, _ := strconv.ParseFloat(result, 64)
                calc_values = append(calc_values, result_as_float)
            }
            put_cache(k, calc_values)
            values = calc_values
        }
        //if !first {
            header.WriteString( divider )
        //}
        header.WriteString( fmt.Sprintf(app_data.format.template_string, k) )
        if last {
            header.WriteString(fmt.Sprintf("%s%10s%s", divider, "row", divider ))
        }
        for i := range values {
            //should this section be outside of the loop
            if !first {
            }
                rows[i].WriteString( divider )
            //}
            column := ""
            if i<len(values) {
                format := app_data.format.template_float
                if is_interface_a_string(values[i]) {
                    format = app_data.format.template_string
                }
                column = fmt.Sprintf(format, values[i])
            }
            rows[i].WriteString( column )
            if last {
                rows[i].WriteString(fmt.Sprintf("%s%10d%s", divider, i, divider))
            }
        }
        first = false
    }

    if app_data.sort {
        sort.Slice(rows, func(i, j int) bool {
            return rows[i].String() < rows[j].String()
        })
    }
    return header, rows, keys
}

const (
    SUMMARY_FUNCS = "avg, count, har, max, medium, mode, min, nop, sum, sdev"
)
// Summaries a form by printing out a table, first row is header, last row is
// summary row. Each column is represented on the summary row based on data
// example: sum main avg,avg
// * @param form name of form to summarize
// * @param args dash delimitated list of summarize functions
func Summary(form string, args string) {
    var out bytes.Buffer
    if 0<len(form) {
        v("sumarize form %s with %s\n", form, args)
        if app_data.data.Forms[form]==nil {
            fmt.Printf("Could not find form '%s'.\n", form)
            return
        }
        Table(form)
        out.WriteString(" ")
        first_form := app_data.data.Forms[form][0]
        var alist []string
        if len(args)<1 {
            form_summary := app_data.data.Settings[form+".summary"]
            if 0<len(form_summary) {
                alist = strings.Split(form_summary, ",")
            }
        } else {
            alist = strings.Split(args, ",")
        }
        for i,value := range alist {
            if i<len(app_data.data.Forms[form]) {
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
                    for i,cached_value := range raw {
                        data[i] = cached_value
                    }
                }
                if 0<i {
                    out.WriteString( app_data.format.divider )
                }
                result := ""
                switch value {
                    case "a", "avg":
                        result = fmt.Sprintf(app_data.format.template_float, Average(data) )
                    case "c", "cnt", "count":
                        result = fmt.Sprintf(app_data.format.template_decimal, len(data))
                    case "h", "har", "harmonic":
                        result = fmt.Sprintf(app_data.format.template_float, Harmonic(data))
                    case "mx", "max":
                        result = fmt.Sprintf(app_data.format.template_float, Max(data))
                    case "m", "med", "medium":
                        result = fmt.Sprintf(app_data.format.template_float, Median(data))
                    case "md", "mod", "mode":
                        result = fmt.Sprintf(app_data.format.template_float, Mode(data))
                    case "mn", "min":
                        result = fmt.Sprintf(app_data.format.template_float, Min(data))
                    case "n", "nop":
                        result = fmt.Sprintf(app_data.format.template_string, "")
                    case "s", "sum":
                        result = fmt.Sprintf(app_data.format.template_float, Sum(data))
                    case "sd", "dev", "sdev":
                        sd := StandardDeviation(data)
                        result = fmt.Sprintf(app_data.format.template_float, sd)
                }
                out.WriteString ( result )
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
        if is_interface_a_number(value) {
           total = total + interface_to_float(value)
            count = count + 1
        }
    }
    average = total / float64(count)
    return average
}

func Harmonic(data []interface{}) float64 {
    total := 0.0
    count := 0
    harmonic := 0.0
    for _, value := range data {
        if is_interface_a_number(value) {
            total = total + ( 1.0 / interface_to_float(value) )
            count = count + 1
        }
    }
    harmonic = float64(count) / total
    return harmonic
}

func StandardDeviation (data []interface{}) float64 {
    var sum, mean, sd float64 = 0, 0, 0
    count_i := 0;//len(data)
    count_f := 0.0//float64(count_i)

    for i:=0 ; i<count_i; i++ {
        if is_interface_a_number(data[i]) {
            sum += interface_to_float(data[i])
            count_i = count_i + 1
        }
    }
    count_f = float64(count_i)
    mean = sum / count_f
    for i:=0 ; i<count_i; i++ {
        if is_interface_a_number(data[i]) {
            sd += math.Pow( interface_to_float(data[i])-mean, 2)
        }
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
        if is_interface_a_number(value) {
            max = math.Max(max, interface_to_float(value))
        }
    }
    return max
}

func Min(data []interface{}) float64 {
    min := math.MaxFloat64
    for _,value := range data {
        if is_interface_a_number(value) {
            min = math.Min(min, interface_to_float(value))
        }
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
    
    format := "%4s %-14s %-14s %-40s\n"
    
    forty := strings.Repeat("-",40)
    fmt.Printf(format, "Flag", "Long", "Arguments", "Description")
    fmt.Printf(format,"----","------------","------------",forty)
    fmt.Printf(format, "c", "create", "name?",
        "create a row in each column, or a new named column")
    fmt.Printf(format, "r", "read", "col row", "read a column row")
    fmt.Printf(format, "u", "update", "col row val", "update a column row")
    fmt.Printf(format, "d", "delete", "index", "delete a row by number")
    fmt.Printf(format, "", "", "name", "delete a column by name")
    fmt.Printf(format, "a", "append", "<value list>", "append a table")
    fmt.Printf(format, "A", "append-by-name", "name:value...", "append a table with named columns")

    fmt.Printf(format, "n", "rename", "src dest", "rename a column from src to dest")
    fmt.Printf("\n")

    fmt.Printf(format, "fc", "form-create", "name list", "Create a form")
    fmt.Printf(format, "fr", "form-read", "name?", "list forms, all if name is not given")
    fmt.Printf(format, "fu", "form-update", "name formula", "update a form")
    fmt.Printf(format, "fd", "form-delete", "name", "delete a form")
    fmt.Printf(format, "fn", "form-rename", "src dest", "rename a form from src to dest")
    fmt.Printf("\n")

    fmt.Printf(format, "cc", "calc-create", "name formula", "Create a calculation")
    fmt.Printf(format, "cr", "calc-read", "name?", "list calculations, all if name is not given")
    fmt.Printf(format, "cu", "calc-update", "name formula", "update a calculation")
    fmt.Printf(format, "cd", "calc-delete", "name", "delete a calculation")
    fmt.Printf(format, "cn", "calc-rename", "src dest", "rename a calculation from src to dest")
    fmt.Printf("\n")

    fmt.Printf(format, "t", "table", "form?",
        "display a table, optionally as a form")
    fmt.Printf(format, "sum", "summary", "form list",
        "summarize a form with function list:")
    fmt.Printf(format, "", "", "", "avg,count,max,min,medium,mode,min,nop,sum,sdev")
    fmt.Printf(format, "l", "ls list", "", "")
    fmt.Printf(format, "", "row", "form row?", "row from form")
    fmt.Printf("\n")

    fmt.Printf(format, "s", "save", "", "save database to file")
    fmt.Printf(format, "", "dump", "", "output the current data")
    fmt.Printf(format, "q", "quit", "", "quit interactive mode")
    fmt.Printf(format, "", "exit", "", "quit interactive mode")
    fmt.Printf(format, "h", "help", "", "this output")
    fmt.Printf(format, "e", "echo", "string", "echo out something")
    fmt.Printf(format, "-", "----", "sep count", "print out a separator")
    fmt.Printf(format, "", "file", "name?", "set or print current file name")
    fmt.Printf(format, "", "rpn", "path?", "set or print current rpn command")
    fmt.Printf(format, "", "verbose", "", "toggle verbose mode")
    fmt.Printf(format, "", "sort?", "", "output current sorting state")
    fmt.Printf(format, "", "sort", "", "toggle the current sort mode")
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

func FormCreate(args []string) {
    if len(args)<2 {
        e(ERR_MSG_FORM_create)
    } else {
        name := arg(args, 0, "")
        if len(name)<1 {
            e(ERR_MSG_FORM_REQUIRED, name)
        } else {
            if app_data.data.Forms[name]!=nil {
                e(ERR_MSG_FORM_EXISTS, name)
            } else {
                items := args[1:]
                app_data.data.Forms[name] = items
            }
        }
    }
}

func FormRead(args []string) {
    name := arg(args, 0, "")
    if len(name)<1 {
        fmt.Printf("%+v\n", app_data.data.Forms)//TODO: make this pretty
    } else {
        fmt.Printf("%+v\n", app_data.data.Forms[name])//TODO: make this pretty
    }
}

func FormUpdate(args []string) {
    if len(args)<2 {
        e(ERR_MSG_FORM_UPDATE)
    } else {
        name := arg(args, 0, "")
        if len(name)<1 {
            e(ERR_MSG_FORM_REQUIRED)
        } else {
            items := args[1:]
            if app_data.data.Forms[name] == nil {
                e(ERR_MSG_FORM_NOT_EXISTS, name)
            } else {
                app_data.data.Forms[name] = items
            }
        }
    }
}

func FormDelete(args []string) {
    name := arg(args, 0, "")
    if len(name)<1 {
        e(ERR_MSG_FORM_REQUIRED)
        return
    }
    delete (app_data.data.Forms, name)
}

func FormRename(args []string) {
    if len(args)<2 {
        e(ERR_MSG_FORM_RENAME);
    } else {
        src_name := arg(args, 0, "")
        dest_name := arg(args, 1, "")
        if 0<len(src_name) && 0<len(dest_name) {
            app_data.data.Forms[dest_name] =
                app_data.data.Forms[src_name]
            delete(app_data.data.Forms, src_name)
        }
    }
}

func CalculationCreate (args []string) {
    name := arg(args, 0, "")
    formula := ""
    for i:=1 ; i<len(args) ; i++ {
        formula = formula + " " + args[i]
    }
    if 0<len(name) && 0<len(formula) {
        app_data.data.Calculations[name] = formula
    }
}

func CalculationRead (args []string) {
    name := arg(args, 0, "")
    if 0<len(name) {
        fmt.Printf("%s\n", app_data.data.Calculations[name])
    } else {
        fmt.Printf("%v\n", app_data.data.Calculations)
    }
}

func CalculationUpdate (args []string) {
    name := arg(args, 0, "")
    formula := ""
    for i:=1 ; i<len(args) ; i++ {
        formula = formula + " " + args[i]
    }
    if 0<len(name) && 0<len(formula) {
        if _, ok := app_data.data.Calculations[name] ; ok {
            fmt.Printf("no calculation")
        } else {
            app_data.data.Calculations[name] = formula
        }
    }
}

func CalculationDelete (args []string) {
    name := arg(args, 0, "")
    if 0<len(name) {
        delete(app_data.data.Calculations, name)
    }
}

func CalculationRename (args []string) {
    src_name := arg(args, 0, "")
    dest_name := arg(args, 1, "")
    if 0<len(src_name) && 0<len(dest_name) {
        app_data.data.Calculations[dest_name] =
            app_data.data.Calculations[src_name]
        delete(app_data.data.Calculations, src_name)
    }
}

//create a sample database with 3x2 columns and rows, 2 forms, one setting
func InitDataBase() DataBase {
    data := DataBase{}
    
    data.Columns = make( map[string][]interface{} )
    data.Columns["foo"] = make( []interface{}, 2 )
    data.Columns["foo"] = []interface{}{0.0,1.0,2.0}
    data.Columns["bar"] = make( []interface{}, 2 )
    data.Columns["bar"] = []interface{}{3.0,4.0,3.0}
    data.Columns["rab"] = make( []interface{}, 2 )
    data.Columns["rab"] = []interface{}{5.0,6.0,6.0}

    data.Forms = make( map[string][]string )
    data.Forms["main"] = []string{"foo","bar","foobar", "row"}
    data.Forms["alt"] = []string{"bar","rab","foobar", "row"}

    data.Calculations = make ( map[string]string )
    data.Calculations["foobar"] = "$foo $bar +"
    data.Calculations["row"] = "#row"

    data.Settings = make ( map[string]string )
    data.Settings["author"] = "thomas.cherry@gmail.com"
    data.Settings["main.summary"] = "avg,sum,avg"
    data.Settings["alt.summary"] = "sum,avg,sum"
    
    return data
}

func Initialize(file_name string) {
    data := InitDataBase()
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

/** setup the prompt reader */
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
library store history, take each line and send it to ProcessLine()
*/
func InteractiveAdvance(line *liner.State, data DataBase) {
    fmt.Printf("Database by thomas.cherry@gmail.com\n")
    app_data.running = true
    for app_data.running==true {
        if name, err := line.Prompt(">"); err == nil {
            input := strings.Trim(name, " ")    //clean it
            line.AppendHistory(name)            //save it
            ProcessManyLines(input, data)
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

func ProcessManyLines(raw_line string, data DataBase) {
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
                    Create(app_data.data)
                } else {            //create column
                    CreateColumn(app_data.data, args[0])
                }
            }
        case "r", "read":       //read column row
            if len(args)!=2 {
                fmt.Fprintf(os.Stderr, ERR_MSG_READ_ARGS)
            } else {
                column := args[0]
                row, err := strconv.Atoi(args[1])
                if err==nil {
                    Read(app_data.data, column, row )
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
                    Update(app_data.data, column, row, value)
                }
            }
        case "d", "delete":     //delete row
            if len(args)!=1 || len(args[0])<1{
                fmt.Fprintf(os.Stderr, ERR_MSG_DELETE_ARGS)
            } else {
                row, err := strconv.Atoi(args[0])
                if err==nil {
                    Delete(app_data.data, row)
                } else {            //delete column
                    DeleteColumn(app_data.data, args[0]) //TODO: add way to delete column
                }
            }
        case "n", "rename":     //rename a row
            src_name := arg(args, 0, "")
            dest_name := arg(args, 1, "")
            if 0<len(src_name) && 0<len(dest_name) {
                app_data.data.Columns[dest_name] =
                    app_data.data.Columns[src_name]
                delete(app_data.data.Columns, src_name)
            }
        case "a", "append":
            AppendTable(app_data.data, args)
        case "A", "append-by-name":
            AppendTableByName(app_data.data, args)
        /**************************************************************/
        /* Form CRUD */
        
        case "fc", "form-create":
            FormCreate(args)
        case "fr", "form-read":
            FormRead(args)
        case "fu", "form-update":
            FormUpdate(args)
        case "fd", "form-delete":
            FormDelete(args)
        case "fn", "form-rename":
            FormRename(args)
        
        /**************************************************************/
        /* Calculation CRUD */

        case "cc", "calc-create":
            CalculationCreate(args)
        case "cr", "calc-read":
            CalculationRead(args)
        case "cu", "calc-update":
            CalculationUpdate(args)
        case "cd", "calc-delete":
            CalculationDelete(args)
        case "cn", "calc-rename":
            CalculationRename (args)

        /**************************************************************/
        /* Other actions */

        case "FF", "Form":
            form := ""
            action := "create"
            if 0<len(args) {
                form = args[0]
            }
            if 1<len(args) {
                action = args[1]
            }
            FormFiller(form, action)
        case "ff", "form":
            form := ""
            action := "create"
            if 0<len(args) {
                form = args[0]
            }
            if 1<len(args) {
                action = args[1]
            }
            FormFillerVisual(form, action)
        case "markdown?":
            fmt.Printf("markdown is %t.\n", app_data.format.markdown)
        case "markdown":
            app_data.format.markdown = !app_data.format.markdown
        case "sort?":
            fmt.Printf("sort is %t.\n", app_data.sort)
        case "sort":
            app_data.sort = !app_data.sort
        case "t", "table":
            Table(args[0])
        case "sum", "summary":
            form := arg(args, 0, "main")
            options := arg(args, 1, app_data.data.Settings[args[0]+".summary"])
            Summary(form, options)
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
        case "row":
            Row(args, app_data.data)
        case "-dev":
            Sub(args[0]) //- test function
        case "dump":
            DumpJson()
        case "f", "forms":
            fmt.Printf("%+v\n", app_data.data.Forms)//TODO: make this pretty
        
        /*case "cs", "calcs":
            Nop()*/
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
    //dry_run := flag.String("rpn", "rpn", "command to process calculations")
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
        ProcessManyLines(*init_command, app_data.data)
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
