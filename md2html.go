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
    "math/rand"
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "sync"
    "text/template"
    "time"
    )

/* ************************************************************************** */
// MARK: - Constants

/* Default template to use if no other template can be found */
const FILE_TEMPLATE = `
<!DOCTYPE html>
<html>
<head>
	<title>{{if .Title }} {{ .Title }} {{end}}</title>
	<meta name="generator" content="md2html" />
</head>
<body>
	<div id="content">
		{{if .Content}} {{ .Content }} {{end}}
	</div>
	<hr>
	<div id="footer">
		{{if .Date}}<span id="generated">{{ .Date }}</span>{{end}}
	</div>
</body>
</html>
`

/* ************************************************************************** */
// MARK: - Data Types

/** App Data holding the current state of the application */
type AppData struct {
    Markdown bool
    TimeZone int
    Date string
    Template string
    Verbose *bool
    Limit int
    WalkTree *bool
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
    Path string
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
        fmt.Println(os.Stderr, err)
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
    index := burned.NotRepeated(len(entries))
    entry := entries[index]

    //solve for base directory
    base := filepath.Dir(filePath)
    if strings.HasPrefix(base, "/") {
        //absolute path, strip it out
        base = strings.Replace(base, td.Path, "", -1)
    }
    path := fmt.Sprintf("%s/%s", fileName, entry.Name())

    return path
}

// Allow templates to load other templates
func (td TemplateData) Import(fileName string) string {
    filePath := fileName
    if len(td.Path) > 0 {
        filePath = fmt.Sprintf("%s/%s", td.Path, fileName)
    }
    if len(filePath) < 1 {
        return ""
    }
    content := readFile(filePath)
    content = render(content, td)//allow for template execution
    result := MarkdownToHTML(content)
    return result
}

/* ********************************** */
// MARK: -

type BurnList map[int]bool

// Record a number as being used
func (self BurnList) Burn(num int) {
    self[num] = true
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
func (self BurnList) NotRepeated(max int) int {
    var index int
    if len(self) >= max {
        //everything burned, allow repeats
        index = int(rand.Int31n(int32(max)))
    } else {
        for {
            posible := rand.Int31n(int32(max))
            if burned.Available(int(posible)) {
                index = int(posible)
                burned.Burn(index)
                break
            }
        }
    }
    return int(index)
}

var (
    burned = BurnList{}
)

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
        os.Stderr.WriteString(err.Error() + "\n")
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
        os.Stderr.WriteString(err.Error() + "\n")
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

/** convert a reader to a string */
func ReaderToString(reader io.Reader) string {
    buf := new(strings.Builder)
    _, err := io.Copy(buf, reader)
    if err!=nil {
        os.Stderr.WriteString(err.Error() + "\n")
    }
    output := buf.String()
    return output
}

/* ************************************************************************** */
// MARK: - task functions

/** wrapper function to the markdown package
not tested
@param input markdown
@return html
*/
func MarkdownToHTML(input string) string {
    md := markdown.New(markdown.HTML(true), markdown.Typographer(false))
    output := md.RenderToString([]byte(input))
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
        os.Stderr.WriteString(err.Error() + "\n")
        os.Exit(1)
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
func render(template_content string, data TemplateData) string {
    temp, err := template.New("html").Parse(template_content)
    if err!=nil {
        os.Stderr.WriteString(err.Error() + "\n")
    } else {
        buf := new(strings.Builder)
        err = temp.Execute(buf, data)
        if err!=nil {
            os.Stderr.WriteString(err.Error() + "\n")
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
        fmt.Println(os.Stderr, err)
        return
    }

    var wg sync.WaitGroup
    wg.Add(1)
    WalkTreeFromPath(&wg, appData, current, "")
    wg.Wait()
}

func WalkTreeFromPath(wg *sync.WaitGroup,
        appData AppData,
        path,
        template string) {
    defer wg.Done()
    entries, err := os.ReadDir(path)
    if err != nil {
        fmt.Println(os.Stderr, err)
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
            } else if e.Name() == appData.Template {
                //if a template is found, overwrite the current template
                template = fmt.Sprintf("%s/%s", path, e.Name())
            }
        }
    }

    if len(template) < 1 {
        fmt.Println(os.Stderr, "No template found")
        return
    }
    //sort.Strings(markdownFiles)
    for _, mdFile := range markdownFiles {
        wg.Add(1)
        go ProcessMarkdownThread(mdFile, template, wg)
    }
    //sort.Strings(dirList)
    for _, dir := range dirList {
        wg.Add(1)
        go WalkTreeFromPath(wg, appData, dir, template)
    }
}

func ProcessMarkdownThread(src, template string, wg *sync.WaitGroup) {
    dest := src[0:len(src)-3] + ".html"
    srcReader := strings.NewReader(readFile(src))
    appData := AppData{
        TimeZone: -4,
        Template: template,
        Markdown: true,
        Limit: 3,
    }
    content := work(srcReader, appData)
    writeFile(dest, content)
    defer wg.Done()
}

/* ************************************************************************** */
// MARK: - App functions

/**
Take a reader and the current application configuration and process a markdown
file using a template as a wrapper.
*/
func work(reader io.Reader, app_data AppData) string {
    content_markdown := ReaderToString(reader)
    content_title := ""

    data := TemplateData{}

    //set up reporting time/date
    now := time.Now()
    formated_date := now.Format("2006-01-02 03:04 PM")

    //get template file
    template_file_path := ""
    if strings.HasPrefix(app_data.Template, "/") {
        template_file_path = app_data.Template
    } else {
        template_file_path = FindTemplate(app_data.Template, app_data.Limit)
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

    re := regexp.MustCompile(`[^\w]`)
    cleanTitle := re.ReplaceAll(
        []byte(strings.ToLower(content_title)),
        []byte("_"))

    //get the content
    var content_html string
    if app_data.Markdown {
        td := TemplateData{}
        td.Path = data.Path
        td.Title = content_title
        td.SafeTitle = string(cleanTitle)
        td.Content = ""
        td.Date = formated_date
        content_markdown = render(content_markdown, td)
        content_html = MarkdownToHTML(content_markdown)
    } else {
        content_html = content_markdown
    }

    //expose data to template
    data.Title = content_title
    data.SafeTitle = string(cleanTitle)
    data.Content = content_html
    data.Date = formated_date

    return render(template_content, data)
}

/** Assign defaults to an AppData structure */
func InitApp(now time.Time) AppData {
    year := now.Year()
    month := int(now.Month())
    day := now.Day()

    date := fmt.Sprintf("%d-%d-%d", year, month, day)

    appData := AppData{
        TimeZone: -4,
        Date: date,
        Template: "index.thtml",
        Markdown: true,
        Limit: 3,
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

/** command line interface */
func main() {
    appData := InitApp(time.Now())

	//overwrite the usage function
    flag.Usage = HelpMessageCallback

    //process command line arguments
    appData.Verbose = flag.Bool("verbose", false,
        "send more text to Standard Error")
    template := flag.String("template", "index.thtml", "look for Template File")
    search_limit := flag.Int("limit", 3,
        "Limit the number of parent directories that can be searched")
    no_html := flag.Bool("no-html", false, "Do not convert markdown to HTML")
    write_template := flag.Bool("write-template", false,
        "save the template and exit")
    markdown_only := flag.Bool("markdown", false, "Only convert input")
    appData.WalkTree = flag.Bool("walk-tree", false,
        "Walk a tree applying conversions")

    flag.Parse()

    if *write_template {
        writeFile("template.html", FILE_TEMPLATE)
        os.Exit(0)
    }

    if len(*template) > 0 { appData.Template = *template }
	if *no_html {appData.Markdown = false}
	if 0 < *search_limit && * search_limit < 99 {appData.Limit = *search_limit}

    //run things
    if *appData.WalkTree {
        WalkTree(appData)
        return
    }

    reader := bufio.NewReader(os.Stdin)
    if *markdown_only {
        markdown := ReaderToString(reader)
        MarkdownToHTML(markdown)
    } else {
        fmt.Println(work (reader, appData))
    }
}
