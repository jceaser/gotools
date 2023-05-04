package main

import ("fmt"
    "math/rand"
    "os"
    "sort"
    "strings"
    "time"
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
    for _, key := range Keys(data) {
        v := data[key]
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

type DeltaContext struct {
    ParentNode interface{}
    ParentName string
    ArrayIndex int
}

func (cxt *DeltaContext) Set (node interface{}, name string) {
    (*cxt).SetArray(node, name, -1)
}

func (cxt *DeltaContext) SetArray (node interface{}, name string, index int) {
    (*cxt).ParentNode = node
    (*cxt).ParentName = name
    (*cxt).ArrayIndex = index
}

func Set(context DeltaContext, value interface{}) {
    if asmap, okay := context.ParentNode.(map[string]interface{}) ; okay {
        asmap[ context.ParentName ] = value
    } else if asarray, okay := context.ParentNode.([]interface{}) ; okay {
        asarray[ context.ArrayIndex] = value
    }
}

/* turn a JSON list into an actual array */
func list(data interface{}) []interface{} {
    var ret []interface{}
    if as_array, okay := data.([]interface{}) ; okay {
        return as_array
    }
    return ret
}

func Keys(data interface{}) []string {
    list := []string{}
    
    if as_map, okay := data.(map[string]interface{}) ; okay {
        for key := range as_map {
            list = append(list, key)
        }
    }
    sort.Strings(list)
    return list
}

func Find (data interface{}, path []string) interface{} {
    context := DeltaContext{nil, "", -1}
    return find (context, data, path)
}

//*map[string]interface{}
/*
- parent JSON parent node
- pname parent node name
- data current node
- path list
*/
func find (context DeltaContext, data interface{}, path []string) interface{} {
    var ret []interface{}
    if asmap, okay := data.(map[string]interface{}) ; okay {
        if sub, okay := asmap[path[0]] ; okay {
            context.Set(data, path[0])
            found := find(context, sub, path[1:])
            ret = append(ret, list(found)...)
        }
    } else if asarray, okay := data.([]interface{}) ; okay {
        for i, item := range asarray {
            context.SetArray(data, path[0], i)
            found := find(context, item, path)
            ret = append(ret, list(found)...)
        }
    } else if asnum, okay := data.(int) ; okay {
        ret = append(ret, asnum)
    } else if asnum, okay := data.(float64) ; okay {
        ret = append(ret, asnum)
    } else if asstr, okay := data.(string) ; okay {
        ret = append(ret, asstr)
    } else {
        fmt.Printf ("not known %s\n", data)
    }
    return ret
}

func Update (data interface{}, path []string, value interface{}) {
    context := DeltaContext{nil, "", -1}
    update (context, data, path, value)
}

//*map[string]interface{}
/*
- parent JSON parent node
- pname parent node name
- data current node
- path list
*/
func update (context DeltaContext,
        data interface{},
        path []string,
        value interface{}) {
    //fmt.Println (context)
    if asmap, okay := data.(map[string]interface{}) ; okay {
        if sub, okay := asmap[path[0]] ; okay {
            context.Set(data, path[0])
            update(context, sub, path[1:], value)
        }
    } else if asarray, okay := data.([]interface{}) ; okay {
        for i, item := range asarray {
            context.SetArray(data, path[0], i)
            update (context, item, path, value)
        }
    } else if asnum, okay := data.(int) ; okay {
        fmt.Printf ("num %f of %s %s\n", asnum, context.ParentName, value)
        Set(context, value)
    } else if asnum, okay := data.(float64) ; okay {
        fmt.Printf ("num %f of %s %s\n", asnum, context.ParentName, value)
        Set(context, value)
        
    } else if asstr, okay := data.(string) ; okay {
        fmt.Printf ("str %s of %s %s\n", asstr, context.ParentName, value)
        Set(context, value)
        
    } else {
        fmt.Printf ("not known %s\n", data)
    }
}

/******************************************************************************/
// MARK: - Command Line Functions

func init() {
    rand.Seed(time.Now().UnixNano())
}

/* Command line interface*/
func main() {
    data := Load(os.Args[1])
    
    data = Change(data)
    
    fmt.Println (Keys(data))
    
    fmt.Println ("\n/****/")
    Update(data, []string{"some-list","some"}, "test")
    Update(data, []string{"some-list","val"}, 3.14)
    Update(data, []string{"random"}, rand.Intn(100) )
    fmt.Println ("/****/\n")
    
    fmt.Println(Find(data, []string{"some-list","some"}))
    fmt.Println(Find(data, []string{"some-list","val"}))
    fmt.Println(Find(data, []string{"random"}))
    
    fmt.Println ("/****/\n")
    
    //parent := Find(data, []string{"some-object"})
    //obj := Find(data, []string{"some-object", "a"})
    //fmt.Println (parent, obj)
    
    Dump(data)
        
    Save(data, os.Args[2])
}