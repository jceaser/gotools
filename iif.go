package main

import ("os"
	"fmt"
    "flag"
	"strconv"
    )

/*
iff left cmd right value other
*/
func main() {
    left := flag.String("left", "true", "left hand")
	cmd := flag.String("test", "==", "actions: [== != <= < >= > && ||]")
	right := flag.String("right", "true", "right hand")
	value := flag.String("value", "true", "return value on success")
	alt := flag.String("alt", "false", "return value on fail")
    
    flag.Parse()
    
    state := "unknown"
    
    switch *cmd {
        case "==":
            state = iff (*left==*right, *value, *alt)
        case "!=":
            state = iff (*left!=*right, *value, *alt)
        case "<=":
            state = iff (*left<=*right, *value, *alt)
        case "<":
            state = iff (*left<*right, *value, *alt)
        case ">=":
            state = iff (*left>=*right, *value, *alt)
        case ">":
            state = iff (*left>*right, *value, *alt)
		case "&&":
			l, _ := strconv.ParseBool(*left)
			r, _ := strconv.ParseBool(*right)
			state = iff (l && r, *value, *alt)
		case "||":
			l, _ := strconv.ParseBool(*left)
			r, _ := strconv.ParseBool(*right)
			state = iff (l || r, *value, *alt)
    }
    fmt.Printf("%s\n", state)
	if (state==*value) {
		os.Exit(0);
	} else {
		os.Exit(1);	
	}
}

func iff(test bool, good string, bad string) string {
    if test==true {
        return good
    } else {
        return bad
    }
}