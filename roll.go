package main

/*
find files in the pattern of name.ext and roll them back to name.num.ext. Any name.num.ext should also be rolled back by one
*/

import ("os"
    "fmt"
    "flag"
    //"testing"
    "bufio"
    //"strings"
    "path/filepath"
    )

type App_Data struct {
    file_name string
    verbose bool
}

const (
    ERR_MSG_01 = "01: File '%s' does not exist\n"
    ERR_MSG_02 = "02: File '%s' could not be moved to %s: %s\n"
)

var app_data App_Data

func handleFlags() {
    
    raw_verbose := flag.Bool("verbose", false, "debug info")
    raw_help := flag.Bool("help", false, "help")
    raw_name := flag.String("name", "log.txt", "name of file to roll over")
    
    flag.Parse()
    
    if *raw_help {
        fmt.Printf("cmd --help --verbose --name\n")
        fmt.Printf("\n")
        fmt.Printf("Roll files down to make room or a new file. Rolled files have a number in the file name. If file is 'foo.txt' then it will be rolled over to foo.0.txt. If there was a foo.0.txt already then it will be rolled down to foo.1.text and so on.\n")
        os.Exit(-1)
    }
    app_data.file_name = *raw_name
    app_data.verbose = *raw_verbose
}

func Move (folder string, source string, destination string) bool {
    result := true
    src_file := folder+source
    dest_file := folder+destination
    
    vprintf("%s->%s\n", src_file, dest_file);
    err := os.Rename(src_file, dest_file)
    if err!=nil {
        fmt.Printf(ERR_MSG_02, src_file, dest_file, err)
        result = false
    } else {
        result = true
    }
    return result
}

func exists(path string) bool {
    var e = false
    if _, err := os.Stat(path); err == nil {
        e = true
    }
    return e;
}

func vprintf(format string, options ...interface{}) {
    if (app_data.verbose) {
        fmt.Printf(format, options...)
    }
}

func ReadFromStdin(name string) {
    f, _ := os.Create(name)
    defer f.Close()
    w := bufio.NewWriter(f)
    
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        //fmt.Println(scanner.Text())
        fmt.Fprintf(w, "%s\n", scanner.Text())
    }
    w.Flush()

    if err := scanner.Err(); err != nil {
        //log.Println(err)
    }
}

/** primary entry point to task at hand */
func work() {
    if exists(app_data.file_name) {
        //something to do
        var dir, full_name = filepath.Split(app_data.file_name)
        var ext = filepath.Ext(full_name)
        var name = full_name[0:len(ext)]
        
        vprintf("in: '%s', full: '%s', name: '%s', ext: '%s'\n",
            dir, full_name, name, ext)

        //now look for older copies
        var looking = true  //still looking for the last file to roll
        var found = false   //found files to roll
        var last_found = "" //name of the last one found
        var i = 0           //index number in file
        var first = 0       //first index number 0 or 1
        var not_found_count = 0        
        
        for looking {
            var test = dir
            //if 0<len(dir) { test = "/"}
            test = test + fmt.Sprintf("%s.%d%s", name, i, ext)
            vprintf("looking for file '%s' in '%s'.\n", test, dir)
            if exists(test) {//found file, is it the first?
                found = true
                last_found = test
                i = i+1
                not_found_count = 0
            } else if found==false {//don't assume the first one is 0 or 1
                first = first + 1
                i = i+1
                not_found_count = not_found_count + 1
            } else {
                //i = i+1
                looking = false
            }
            if not_found_count>10 {
                break
            }
        }

        vprintf("Last file found is '%s', and first index is %d.\n", last_found, first)
        
        if found && 0<len(last_found) {
            //something was found
            for c:=i-1; (first-1)<c ; c-- {
                //work backwards
                src_file := fmt.Sprintf("%s.%d%s", name, c, ext);
                dest_file:= fmt.Sprintf("%s.%d%s", name, c+1, ext)
                
                Move(dir, src_file, dest_file)
            }
        }
        orig_file := fmt.Sprintf("%s%s", name, ext);
        dest_file := fmt.Sprintf("%s.%d%s", name, first, ext);

        Move(dir, orig_file, dest_file)
        
    } else {//nothing to do
        fmt.Printf(ERR_MSG_01, app_data.file_name)
    }
}

/*
*/
func main() {
    handleFlags()
    
    vprintf("test this %s %s\n", "output", "text");
    
    work()
}

/******************************************************************************/

/*func TestVprintf(*testing.T) {
    app_data.verbose = false
    vprintf("")
}*/