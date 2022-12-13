package main

import ("fmt"
    "os"
    "io/ioutil"
    "encoding/json"
    )

func Load(file string) map[string]interface{} {
    json_raw, err := os.Open(file)
    if err!=nil {
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
        os.Exit(1)
    }
    defer json_raw.Close()
    
    var json_data map[string]interface{}
    bytes, err := ioutil.ReadAll(json_raw)
    if err!=nil {
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
        os.Exit(2)
    } else {
        json.Unmarshal([]byte(bytes), &json_data)
        return json_data
    }
    return nil
}

func Save(data map[string]interface{}, file string) {
    json_text, err := json.MarshalIndent(data, "", "    ")
    if err!=nil {
        fmt.Fprintf(os.Stderr, "error: %s\n", err)
        os.Exit(3)
    }
    err = ioutil.WriteFile(file, json_text, 0644)
    if err!=nil {
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
        os.Exit(4)
    }
}

func main() {
    Save(Load(os.Args[1]), os.Args[2])
}