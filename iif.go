package main

import ("fmt"
    "flag"
    )

/*
iff left cmd right value other
*/
func main() {
    //args := os.Args
    
    left := flag.String("left", "true", "line mode, trim each line")
    right := flag.String("right", "true", "edge mode, trim just edges")
    //allMode := flag.String("all", false, "all mode, trim everything")
    
    flag.Parse()
    
    //fmt.Printf("%s\n%s\ndone\n", *left, *right)
    
    /*for _, a := range flag.Args() {
        fmt.Printf("'%s'\n", a)
    }*/
    
    state := "unknown"
    var cmd string
    var value string
    var alt string
    
    if 0<len(flag.Args()){
        *left = flag.Arg(0)
    }
    
    if 1<len(flag.Args()){
        cmd = flag.Arg(1)
    } else {
        cmd = "=="
    }
    if 2<len(flag.Args()){
        *right = flag.Arg(2)
    } else {
        *right = "true"
    }
    
    if 3<len(flag.Args()){
        value = flag.Arg(3)
    } else {
        value = "true"
    }
    
    if 4<len(flag.Args()){
        alt = flag.Arg(4)
    } else {
        alt = "false"
    }
    
    /*****/
    
    switch cmd {
        case "==":
            state = iff (*left==*right, value, alt)
        case "!=":
            state = iff (*left!=*right, value, alt)
        case "<=":
            state = iff (*left<=*right, value, alt)
        case "<":
            state = iff (*left<*right, value, alt)
        case ">=":
            state = iff (*left>=*right, value, alt)
        case ">":
            state = iff (*left>*right, value, alt)
    }
    fmt.Printf("%s\n", state)

}

func iff(test bool, good string, bad string) string {
    if test==true {
        return good
    } else {
        return bad
    }
}