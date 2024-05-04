package main

/* ****************************************************************************
Read in content markdown file
find template file
merge template and content
convert markdown to HTML

**************************************************************************** */

// go get gitlab.com/golang-commonmark/markdown
// 'printf "# Title\n* list item" | go run md2html.go'

import (
    "bufio"
    "flag"
    "fmt"
    "gitlab.com/golang-commonmark/markdown"
    "io"
    "io/ioutil"
    "log"
    "math/rand"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "sync"
    "text/template"
    "time"
    //"reflect"
    )

/* ************************************************************************** */
// MARK: - Constants

/* Default template to use if no other template can be found */
const FILE_TEMPLATE = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8" />
	<title>{{if .Title }} {{ .Title }} {{end}}</title>
	<meta name="generator" content="md2html" />
</head>
<body>
    <section id="header">{{if .Title }} {{ .Title }} {{end}}</header>
	<section id="content">
		{{if .Content}} {{ .Content }} {{end}}
	</section>
	<hr>
	<section id="footer">
		{{if .DateWithTime}}
		<span id="generated">{{ .DateWithTime }}</span>
		{{end}}
	</section>
</body>
</html>
`

const EVENT_ARTICLE = `
<!-- article start -->
<article class="blog_item">
    <div class="content">
        %s
    </div>
    <div class="tools">
        <span>%s</span>
        <img class="qrcode" src="%s" alt="qrcode" title="QR Code">
        <a class="direct" href="%s"><i class="fas fa-link"></i></a>
    </div>
</article>
<!-- article end -->
`

type ExitCode int

const (
    ExitOkay int = iota
    ExitBadTemplate
    ExitBadLimit
    ExitErrorWorkingDir
    ExitUnkownTimezone
)

/* ************************************************************************** */
// MARK: - Data Types

//loggers
type LogType struct {
    Error *log.Logger
    Warn *log.Logger
    Info *log.Logger
    Debug *log.Logger
    Stats *log.Logger
}

var Log LogType

func init() {
    file := os.Stderr
    settings := log.Ldate|log.Ltime|log.Lshortfile
    Log = LogType{
        Error: log.New(file, "ERROR: ", settings),
        Warn: log.New(file, "WARNING: ", settings),
        Info: log.New(ioutil.Discard, "INFO: ", settings),
        Debug: log.New(ioutil.Discard, "DEBUG: ", settings),
        Stats: log.New(ioutil.Discard, "", settings)}
}

func (self *LogType) EnableInfo() {
    self.Info.SetOutput(os.Stderr)
}

func (self *LogType) EnableDebug() {
    self.Debug.SetOutput(os.Stderr)
}

func (self *LogType) EnableStats(file *os.File) {
    self.Stats.SetOutput(file)
}

/* ********************************** */
// MARK: -

/** App Data holding the current state of the application */
type AppData struct {
    MarkdownOff *bool
    Location *string
    TimeZone int
    //Date string
    Template *string
    Verbose *bool
    Limit *int
    UserMessageOne *string
    UserMessageTwo *string
    Name *string

}

func (self AppData) Now() time.Time {
    loc, e := time.LoadLocation(*self.Location)
    if e!=nil {
        fmt.Println(e)
        os.Exit(ExitUnkownTimezone)
    }
    now := time.Now().In(loc)
    return now
}

/* ********************************** */
// MARK: -

/**
Data to be passed to the template engine to be swapped out with template
variables
*/
type TemplateData struct {
    Title string
    SafeTitle string
    Content string
    Date string
    Time string
    DateWithTime string
    Path string
    UserMessageOne string
    UserMessageTwo string
    burned BurnList
    FileName string
    QrCode string
    Url string
}

func (self TemplateData) HasStyle() bool {
    l := len(self.FileName)
    if 0<l {
        withoutExtention := self.FileName[0:l-3]
        styleSheet := withoutExtention + ".css"
        return FileExists(styleSheet)
    }
    return false
}

func (self TemplateData) Style() string {
    l := len(self.FileName)
    if 0<l {
        withoutExtention := self.FileName[0:l-3]
        styleSheet := withoutExtention + ".css"
        return styleSheet
    }
    return ""
}


func (self TemplateData) Has(sub string) bool {
    dir := filepath.Dir(fmt.Sprintf("%s/%s", self.Path, self.FileName))
    return FileExists(fmt.Sprintf("%s/%s", dir, sub))
}

func (self TemplateData) Dump() string {
    return fmt.Sprintf("%v", self)
}

// tests if a path exists
func (td TemplateData) HasTitle(title string) bool {
    return td.Title == title
}

func (td TemplateData) NameHasSuffix(name string) bool {
    return strings.HasSuffix(td.FileName,  name)
}

func (td TemplateData) NameIs(name string) bool {
    return td.FileName == name
}

// tests if a path exists
func (td TemplateData) Exists(path string) bool {
    if FileExists(td.Path + "/" + path) {
        return true
    }
    return false
}

// Return a random resource
func (td TemplateData) Random(fileName string) string {

    //construct full path to resources
    filePath := fileName
    if len(td.Path) > 0 {
        filePath = fmt.Sprintf("%s/%s", td.Path, fileName)
    }
    if len(filePath) < 1 {
        return ""
    }

    //get all the objects in a directory
    rawEntries, err := os.ReadDir(filePath)
    if err != nil {
        Log.Warn.Printf("%s - %s: %v", td.Title,  filePath, err)
        return ""
    }

    //just consider the viable items
    entries := []os.DirEntry{}
    for _, ent := range rawEntries {
        if !strings.HasPrefix(ent.Name(), ".") && !ent.IsDir() {
            //not hidden, not a directory
            entries = append(entries, ent)
        }
    }

    //find resource to return
    index := td.burned.NotRepeated(len(entries))
    if index > len(entries) || index < 0 {
        return ""
    }
    entry := entries[index]

    /*
    //remove base directory
    base := filepath.Dir(filePath)
    if strings.HasPrefix(base, "/") {
        //absolute path, strip it out
        base = strings.Replace(base, td.Path, "", -1)
    }
    */
    path := fmt.Sprintf("%s/%s", fileName, entry.Name())

    return path
}

/*
Allow templates to load other templates
fileName is the directory to look into
*/
func (td TemplateData) Import(fileName string) string {
    //td.path is full base path
    //td.Name is relative calling page
    //fileName is requested path to include

    dir := filepath.Dir(td.FileName)
    //base := filepath.Base(td.FileName)

    dirPath := fmt.Sprintf("%s/%s", dir, fileName)

    if DirectoryExists(dirPath) {
        entries, err := os.ReadDir(dirPath)
        if err != nil {
            Log.Error.Println(err)
            return "error in directory"
        }
        out := ""
        for _, e := range entries {
            if e.Name()[len(e.Name())-3:] == ".md" {
                justName := e.Name()[0:len(e.Name())-3]
                srcPath := fmt.Sprintf("%s/%s.md", fileName, justName)
                destPath := fmt.Sprintf("/%s/%s.html", dirPath, justName)
                localPath := fmt.Sprintf("%s/%s.html", fileName, justName)

                subPath := fmt.Sprintf("%s/%s/%s", td.Path, dir, srcPath)

                td.Url = localPath
                td.QrCode = fmt.Sprintf("/cgi-bin/qrc.cgi?size=100&path=%s",
                    destPath)

                content := readFile(subPath)
                content = Render(content, td)
                result := MarkdownToHTML(content)
                out = out + "\n" + fmt.Sprintf(EVENT_ARTICLE,
                    result,
                    "",
                    td.QrCode,
                    localPath)
            }
        }
        return out
    }
    filePath := fileName
    if len(td.Path) > 0 {
        filePath = fmt.Sprintf("%s/%s", td.Path, fileName)
    }
    if len(filePath) < 1 {
        return ""
    }

    content := readFile(filePath)
    content_title := ""
    //look for an H1 title, break it out for use in the HTML title tag
    first_newline := strings.Index(content, "\n")
    if first_newline > -1{
        if strings.HasPrefix(content[0:first_newline], "# ") {
            content_title = content[2:first_newline]
            content = content[first_newline:]
        }
    }
    td.Title = content_title

    content = Render(content, td)//allow for template execution
    result := MarkdownToHTML(content)
    return result
}

/* ********************************** */
// MARK: -

/*
Burn List maintains a list of numbers which have been used up. For cases where
random numbers are to be used just once. A Map is used instead of a List for
faster lookup.
*/
type BurnList map[int]int

// Record a number as being used
func (self *BurnList) Burn(num int) {
    (*self)[num] = len(*self)
}

// Number has not been used yet
func (self BurnList) Available(num int) bool {
    if _, okay := self[num]; okay {
        return false
    }
    return true
}

// Number has already been used
func (self BurnList) Burned(num int) bool {
    if _, okay := self[num]; okay {
        return true
    }
    return false
}

// Try numbers till one of them is found to have not been used
func (self *BurnList) NotRepeated(max int) int {
    if len(*self) >= max {
        //everything burned, allow repeats
        //index = int(rand.Int31n(int32(max)))
        return -1
    } else {
        //try randomly, but don't try forever
        limit := max * 2
        for limit > 0 {
            limit = limit - 1
            posible := int(rand.Int31n(int32(max)))
            if self.Available(posible) {
                self.Burn(posible)
                return posible
            }
        }
        //try the hard way
        for i := 0 ; i < max; i++ {
            if self.Available(i) {
                self.Burn(i)
                return i
            }
        }
    }
    return -1
}

/* ************************************************************************** */
// MARK: - Util functions

/**
Util Function Write a file
Not tested!
*/
func writeFile(file string, content string) {
    d1 := []byte(content)
    err := ioutil.WriteFile(file, d1, 0644)
    if err!=nil {
        Log.Warn.Printf("%v", err.Error())
    }
}

/**
Util Function Read a file
Not tested!
@param full path to read
@return empty string on error, file contents otherwise
*/
func readFile(file_path string) string {
    content, err := ioutil.ReadFile(file_path)
    if err != nil {
        Log.Warn.Printf("%v", err.Error())
        return ""
    }
    return string(content)
}

/**
Test if a file exists
@param full path to file to be tested
@return true if no error
*/
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err==nil
}

func DirectoryExists(path string) bool {
    if stat, err := os.Stat(path); err == nil && stat.IsDir() {
        return true
    }
    return false
}

/** convert a reader to a string */
func ReaderToString(reader io.Reader) string {
    buf := new(strings.Builder)
    _, err := io.Copy(buf, reader)
    if err!=nil {
        Log.Warn.Printf("%v", err.Error())
    }
    output := buf.String()
    return output
}

func TimeStart() int64 {
    return time.Now().UnixMicro()
}

func TimeLog(who string, start int64) {
    stop := TimeStart()
    Log.Stats.Printf("%s took %d μs", who, stop-start)
}

/* ************************************************************************** */
// MARK: - task functions

/** wrapper function to the markdown package
not tested
@param input markdown
@return html
*/
func MarkdownToHTML(input string) string {
    start := TimeStart()
    md := markdown.New(markdown.HTML(true), markdown.Typographer(false))
    output := md.RenderToString([]byte(input))
    TimeLog("MarkdownToHTML", start)
    return output
}

/**
Get the current working directory, return it along with the directory count to
root. Working directory is the directory from where the app is launched, not
from where the app lives
@return (path, count)
*/
func WorkingDirectory() (string, int) {
    working_directory, err := os.Getwd()
    if err != nil {
        Log.Error.Printf("%v", err.Error())
        os.Exit(ExitErrorWorkingDir)
    }
    directories := strings.Split(working_directory, "/")
    return working_directory, len(directories)-1
}

/**
Search for a template file, first in the current working directory, then in parent directories till one is found
@param name what the template is called
@return path to found template, empty string otherwise
*/
func FindTemplate(name string, limit int) string {
    const NOT_FOUND = ""
    path, count := WorkingDirectory()
    path_back := "" //start in current directory
    var max int
    if count < limit {
        max = count
    } else {
        max = limit
    }
    for i:=0 ; i<max; i++ {
        new_path := path + "/" + path_back + name
        if FileExists(new_path) {
            return new_path
        }
        path_back = path_back + "../" //look back one directory
    }
    return NOT_FOUND
}

/**
Populate a template with data
@param template_content The template
@param data values to populate in template
@return resulting output of executing the template
*/
func Render(template_content string, data TemplateData) string {
    start := TimeStart()
    defer TimeLog("Render", start)

    temp, err := template.New("html").Parse(template_content)
    if err!=nil {
        Log.Warn.Printf("%v", err.Error())
    } else {
        buf := new(strings.Builder)
        err = temp.Execute(buf, data)
        if err!=nil {
            Log.Warn.Printf("%v", err.Error())
        }
        output := buf.String()
        return output
    }
    return ""
}

//walk a tree of directories converting files from md to html
func WalkTree(appData AppData) {

    current, err := os.Getwd()
    if err != nil {
        Log.Error.Printf("%v", err.Error())
        return
    }

    var wg sync.WaitGroup
    wg.Add(1)
    WalkTreeFromPath(&wg, appData, current, "")
    wg.Wait()
}

//recursive function to process a directory
func WalkTreeFromPath(wg *sync.WaitGroup,
        appData AppData,
        path,
        template string) {
    defer wg.Done()
    entries, err := os.ReadDir(path)
    if err != nil {
        Log.Error.Printf("%v", err.Error())
        return
    }
    markdownFiles := []string{}
    dirList := []string{}
    for _, e := range entries {
        if !strings.HasPrefix(e.Name(), ".") {
            if e.IsDir() {
                dirPath := fmt.Sprintf("%s/%s", path, e.Name())
                dirList = append(dirList, dirPath)
            } else if strings.HasSuffix(e.Name(), ".md") {
                filePath := fmt.Sprintf("%s/%s", path, e.Name())
                markdownFiles = append(markdownFiles, filePath)
            } else if e.Name() == *appData.Template {
                //if a template is found, overwrite the current template
                template = fmt.Sprintf("%s/%s", path, e.Name())
            }
        }
    }

    if len(template) < 1 {
        Log.Warn.Printf("No template found")
        return
    }
    //sort.Strings(dirList)
    for _, dir := range dirList {
        wg.Add(1)
        go WalkTreeFromPath(wg, appData, dir, template)
    }
    //sort.Strings(markdownFiles)
    for _, mdFile := range markdownFiles {
        wg.Add(1)
        go ProcessMarkdownThread(wg, appData, mdFile, template)
    }
}

//process a single file
func ProcessMarkdownThread(wg *sync.WaitGroup, appData AppData, src, template string) {
    start := TimeStart()
    defer TimeLog("ProcessMarkdownThread", start)
    dest := src[0:len(src)-3] + ".html"
    srcReader := strings.NewReader(readFile(src))
    appData.Template = &template
    content := work(srcReader, &appData)
    writeFile(dest, content)
    defer wg.Done()
}

/* ************************************************************************** */
// MARK: - App functions

/**
Take a reader and the current application configuration and process a markdown
file using a template as a wrapper.
*/
func work(reader io.Reader, app_data *AppData) string {
    start := TimeStart()
    defer TimeLog("work", start)
    content_markdown := ReaderToString(reader)
    content_title := ""

    //setup basic template data
    data := TemplateData{}
    data.burned = BurnList{}
    data.Content = ""

    if 0<len(*app_data.Name) {
        data.FileName = *app_data.Name
    }

    data.Url = *app_data.Name

    data.UserMessageOne = *app_data.UserMessageOne
    data.UserMessageTwo = *app_data.UserMessageTwo

    //set up reporting time/date
    now := app_data.Now()
    data.Date = now.Format("2006-01-02")
    data.Time = now.Format("03:04 PM")
    data.DateWithTime = now.Format("2006-01-02 03:04 PM")

    //get template file
    template_file_path := ""
    if strings.HasPrefix(*app_data.Template, "/") {
        template_file_path = *app_data.Template
    } else {
        template_file_path = FindTemplate(*app_data.Template, *app_data.Limit)
    }
    var template_content string
    if template_file_path == "" {
        template_content = FILE_TEMPLATE
        data.Path, _ = os.Getwd()
    } else {
        template_content = readFile(template_file_path)
        data.Path = filepath.Dir(template_file_path)
    }

    //look for an H1 title, break it out for use in the HTML title tag
    first_newline := strings.Index(content_markdown, "\n")
    if first_newline > -1{
        if strings.HasPrefix(content_markdown[0:first_newline], "# ") {
            content_title = content_markdown[2:first_newline]
            content_markdown = content_markdown[first_newline:]
        }
    }
    data.Title = content_title

    re := regexp.MustCompile(`[^\w]`)
    cleanTitle := re.ReplaceAll(
        []byte(strings.ToLower(content_title)),
        []byte("_"))
    data.SafeTitle = string(cleanTitle)

    //get the content
    var content_html string
    if *app_data.MarkdownOff {
        content_html = content_markdown
    } else {
        content_markdown = Render(content_markdown, data)
        content_html = MarkdownToHTML(content_markdown)
    }

    data.Content = content_html

    return Render(template_content, data)
}

/** Assign defaults to an AppData structure */
func InitApp() AppData {
    defTemplate := "index.thtml"
    defLimit := 3
    defMdOff := false
    loc := "UTC"
    appData := AppData{
        TimeZone: 0,
        Location: &loc,
        Template: &defTemplate,
        MarkdownOff: &defMdOff,
        Limit: &defLimit,
    }
    return appData
}

/** Call back function for the flag API defining the help output */
func HelpMessageCallback() {
    fmt.Fprintf(flag.CommandLine.Output(),
        "md2html by thomas.cherry@gmail.com\n\n")
    fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
    flag.PrintDefaults()
}

// Validate the flag values
func processFlags(appData *AppData) {
    if len(*appData.Template) < 1 {
        fmt.Println("invalid template name")
        os.Exit(ExitBadTemplate)
    }
    if *appData.Limit < 0 || 99 < *appData.Limit {
        fmt.Println("invalid search limit")
        os.Exit(ExitBadLimit)
    }
    if len(*appData.Location) < 1 {
        *appData.Location = "UTC"
    }
}

func SetStatsLog(path string) os.File {
    var stream os.File
    if len(path) > 0 {
        stream, err := os.OpenFile(path,
            os.O_APPEND | os.O_WRONLY | os.O_CREATE,
            0644)
        if err != nil {
            Log.Error.Printf("Error opening stat log: %v", err)
        }
        Log.EnableStats(stream)
    }
    return stream
}

/** command line interface */
func main() {
    appData := InitApp()

	//overwrite the usage function
    flag.Usage = HelpMessageCallback

    //process command line arguments
    appData.Verbose = flag.Bool("verbose", false,
        "send more text to Standard Error")
    appData.Template = flag.String("template", "index.thtml",
        "look for Template File")
    appData.Limit = flag.Int("limit", 3,
        "Limit the number of parent directories that can be searched")
    appData.MarkdownOff = flag.Bool("no-markdown", false,
        "Do not convert markdown to HTML")
    appData.UserMessageOne = flag.String("user-message-one", "",
        "A user defined message which can be included in templates")
    appData.UserMessageTwo = flag.String("user-message-two", "",
        "A user defined message which can be included in templates")
    appData.Name = flag.String("name", "", "Name of the file being streamed in")

    appData.Location = flag.String("time-zone", "UTC",
        "EST, " +
        "America/Chicago, " +
        "America/Denver, " +
        "America/Los_Angeles, " +
        "UTC")

    statLog := flag.String("stats-file", "",
        "Path to the file where stats are to be saved")

    modeWalkTree := flag.Bool("walk-tree", false,
        "MODE: Walk a tree applying conversions and exit")
    modeMarkdownOnly := flag.Bool("markdown", false,
        "MODE: Only convert input and exit")
    modeWriteTemplate := flag.Bool("write-template", false,
        "MODE: save the template and exit")

    flag.Parse()

    //setup the stats log
    if len(*statLog) > 0 {
        switch *statLog {
        case "out":
            Log.EnableStats(os.Stdout)
        case "err":
            Log.EnableStats(os.Stderr)
        default:
            statFile := SetStatsLog(*statLog)
            defer statFile.Close()
        }
    }

    processFlags(&appData)

    //run things
    if *modeWriteTemplate {
        writeFile("template.html", FILE_TEMPLATE)
        return
    }

    if *modeWalkTree {
        WalkTree(appData)
        return
    }

    reader := bufio.NewReader(os.Stdin)
    if *modeMarkdownOnly {
        markdown := ReaderToString(reader)
        out := MarkdownToHTML(markdown)
        fmt.Println(out)
        return
    }

    //default action
    {
        out := work(reader, &appData)
        fmt.Println(out)
        return
    }
}
