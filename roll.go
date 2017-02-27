package main

/*
find files in the pattern of name.ext and roll them back to name.num.ext. Any name.num.ext should also be rolled back by one
*/

import ("os"
    "fmt"
    "flag"
    "path/filepath"
    )

type App_Data struct {
    file_name string
}

const (
    ERR_MSG_01 = "01: File '%s' does not exist"
)

var app_data App_Data

func handleFlags() {
    
    raw_help := flag.Bool("help", false, "help")
    raw_name := flag.String("name", "log.txt", "name of file to roll over")
    
    flag.Parse()
    
    if *raw_help {
        fmt.Printf("cmd --help --name\n")
        fmt.Printf("\n")
        fmt.Printf("Roll files down to make room or a new file. Rolled files have a number in the file name. If file is 'foo.txt' then it will be rolled over to foo.0.txt. If there was a foo.0.txt already then it will be rolled down to foo.1.text and so on.\n")
        os.Exit(-1)
    }
    app_data.file_name = *raw_name
}

func exists(path string) bool {
    var e = false
    if _, err := os.Stat(path); err == nil {
        e = true
    }
    return e;
}

/** primary entry point to task at hand */
func work() {
    if exists(app_data.file_name) {
        //something to do
        var dir,full_name = filepath.Split(app_data.file_name)
        var ext = filepath.Ext(full_name)
        
        var name = full_name[0:len(ext)-1]
        
        //now look for older copies
        var looking = true
        var found = false
        var last_found = ""
        var i = 0;
        for looking {
            var test = ""
            if 0<len(dir) {test = dir + "/"}
            test = test + fmt.Sprintf("%s.%d%s", name, i, ext)
            if exists(test) {
                found = true
                last_found = test
                i = i+1
            } else {
                looking = false
            }
        }
        
        if found && 0<len(last_found) {
            //something was found
            for c:=i-1; -1<c ; c-- {
                //work backwards
                src_file := fmt.Sprintf("%s.%d%s", name, c, ext);
                dest_file:= fmt.Sprintf("%s.%d%s", name, c+1, ext)
                
                os.Rename(src_file, dest_file)
            }
        }
        orig_file := fmt.Sprintf("%s%s", name, ext);
        dest_file := fmt.Sprintf("%s.%d%s", name, 0, ext);
        
        os.Rename(orig_file, dest_file)
    } else {//nothing to do
        fmt.Printf(ERR_MSG_01, app_data.file_name)
    }
}

/*
*/
func main() {

    handleFlags()
    
    work()

}
