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
    "fmt"
    "bufio"
    //"encoding/json"
    "flag"
    "gitlab.com/golang-commonmark/markdown"
    "io"
    "io/ioutil"
    "os"
    //"strconv"
    "strings"
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
    Verbose bool
    Limit int
}

/**
Data to be passed to the template engine to be swapped out with template
variables
*/
type TemplateData struct {
    Title string
    Content string
    Date string
}

func (td TemplateData) Random(path string) string {
    return fmt.Sprintf("%s = %s", path, "from the random function")
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

/* ************************************************************************** */
// MARK: - App functions

/**
Take a reader and the current application configuration and process a markdown
file using a template as a wrapper.
*/
func work(reader io.Reader, app_data AppData) string {
    content_markdown := ReaderToString(reader)
    content_title := ""
    
    //look for an H1 title, break it out for use in the HTML title tag
    first_newline := strings.Index(content_markdown, "\n")
    if first_newline > -1{
        if strings.HasPrefix(content_markdown[0:first_newline], "# ") {
            content_title = content_markdown[2:first_newline]
            content_markdown = content_markdown[first_newline:]
        }
    }

    //get the content
    var content_html string
    if app_data.Markdown {
        content_html = MarkdownToHTML(content_markdown)
    } else {
        content_html = content_markdown
    }

    //set up reporting time/date
    now := time.Now()
    formated_date := now.Format("2006-01-02 03:04 PM")

    //expose data to template
    data := TemplateData{
        Title: content_title,
        Content: content_html,
        Date: formated_date,
    }
    
    //get template file
    template_file_path := FindTemplate(app_data.Template, app_data.Limit)
    var template_content string
    if template_file_path == "" {
        template_content = FILE_TEMPLATE
    } else {
        template_content = readFile(template_file_path)
    }
    
    return render(template_content, data)
}

/** Assign defaults to an AppData structure */
func InitApp(now time.Time) AppData {
    year := now.Year()
    month := int(now.Month())
    day := now.Day()

    date := fmt.Sprintf("%d-%d-%d", year, month, day)

    app_data := AppData{
        TimeZone: -4,
        Date: date,
        Template: "index.thtml",
        Markdown: true,
        Verbose: false,
        Limit: 3,
    }
    return app_data
}

/** Call back function for the flag API defining the help output */
func HelpMessageCallback() {
    fmt.Fprintf(flag.CommandLine.Output(),
        "ReadIcal by thomas.cherry@gmail.com\n\n")
    fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
    flag.PrintDefaults()
}

/** command line interface */
func main() {
	//overwrite the usage function
    flag.Usage = HelpMessageCallback

    //process command line arguments
    verbose := flag.Bool("verbose", false, "send more text to Standard Error")
    template := flag.String("template", "index.thtml", "look for Template File")
    search_limit := flag.Int("limit", 3,
        "Limit the number of parent directories that can be searched")
    no_html := flag.Bool("no-html", false, "Do not convert markdown to HTML")
    write_template := flag.Bool("write-template", false,
        "save the template and exit")
    markdown_only := flag.Bool("markdown", false, "Only convert input")

    flag.Parse()

    if *write_template {
        writeFile("template.html", FILE_TEMPLATE)
        os.Exit(0)
    }

    app_data := InitApp(time.Now())

    if len(*template)>0 { app_data.Template = *template }
	if *verbose {app_data.Verbose = *verbose}
	if *no_html {app_data.Markdown = false}
	if 0 < *search_limit && * search_limit < 99 {app_data.Limit = *search_limit}

    reader := bufio.NewReader(os.Stdin)
    if *markdown_only {
        markdown := ReaderToString(reader)
        MarkdownToHTML(markdown)
    } else {
        fmt.Println(work (reader, app_data))
    }

}
