package main

import ("fmt"
    "os"
    "strings"
    "io/ioutil"
    "encoding/json"
    )

//Enum of exit codes
const (
    Success int = iota // 0 is good
    FailOpenFile
    FailReadFile
    FailMarshal
    FailWrite
)

/*
Load a JSON file and return a map containing the data
*/
func Load(file string) map[string]interface{} {
    json_raw, err := os.Open(file)
    if err!=nil {
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
        os.Exit(FailOpenFile)
    }
    defer json_raw.Close()
    
    var json_data map[string]interface{}
    bytes, err := ioutil.ReadAll(json_raw)
    if err!=nil {
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
        os.Exit(FailReadFile)
    } else {
        json.Unmarshal([]byte(bytes), &json_data)
        return json_data
    }
    return nil
}

/*
Save map data as a JSON file
data - map of interfaces using strings for keys
file - path and name of file to save to
*/
func Save(data map[string]interface{}, file string) {
    json_text, err := json.MarshalIndent(data, "", "    ")
    if err!=nil {
        fmt.Fprintf(os.Stderr, "error: %s\n", err)
        os.Exit(FailMarshal)
    }
    err = ioutil.WriteFile(file, json_text, 0644)
    if err!=nil {
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
        os.Exit(FailWrite)
    }
}

/*
Take data, make a change to it and return that changed data. Changes are not
made in place
*/
func Change(data map[string]interface{}) map[string]interface{} {
    for key, _ := range data {
        if key=="change-me" {
            value := data["change-me"]
            switch value.(type) {
                case string:
                    broken := fmt.Sprintf("%v", value)
                    data["change-me"] = strings.ToUpper(broken)
                default:
            }
        }
    }
    return data
}

/*
Walk the tree of objects and process each item
maps made from JSON can have: maps, arrays, strings, or numbers.
*/
func Dump(data map[string]interface{}) {
    for key, v := range data {
        dump(key, v)
    }
}

/*
Print out some interesting details of a data node, can be recursive
*/
func dump(key string, v interface{}) {
    switch i := v.(type) {
        case map[string]interface{}:
            Dump(i)
        case []interface{}:
            for _, inner := range i {
                dump(key, inner)
            }
        case string:
            fmt.Printf("%s is string: %s\n", key, i)
        case float64:
            fmt.Printf ("%s is number: %f\n", key, float64(i))
        default:
            fmt.Printf ("%s is unknown: %v (%T)\n", key, v, i)
    }
}

/* Command line interface*/
func main() {
    data := Load(os.Args[1])
    
    data = Change(data)
    
    Dump(data)
        
    Save(data, os.Args[2])
}