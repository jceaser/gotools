package main

import (
    "fmt"
    "os"
    "bufio"
    "flag"
    "strconv"
    "strings"
    )

type App_Data struct {
    verbose bool
}

var app_data = App_Data{verbose:false,}

func AsColor(color [3]byte, text string) string {
    return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[00m",
        color[0], color[1], color[2], text)
    return text
} 

func ExpandColorText(color [3]byte) string {
    box := AsColor(color, "\uEE03\uEE04\uEE05")
    red := color[0]
    green := color[1]
    blue := color[2]
    return fmt.Sprintf("%s%00x%0x%0x", box, red, green, blue)
}

/******************************************************************************/
/* Application Tasks */

/* Make sure that a supplied value in the byte range of 0-255 */
func ColorLimits(value int) byte {
    var ret byte
    
    if value>255 {
        ret = 255
    } else if value < 0 {
        ret = 0
    } else {
        ret = byte(value)
    }
    return ret
}

func ValueToNumber(reader *bufio.Reader, fallback byte) byte {
    var ret byte
    text, read_err := reader.ReadString('\n')
    if read_err != nil {
        fmt.Fprintf(os.Stderr, "ReadString: %s", read_err)
    }
    text = strings.Trim(text, "\t \n")
    value, err := strconv.Atoi(text)
    if err == nil {
        ret = ColorLimits(value)
    } else {
        fmt.Fprintf(os.Stderr, "Bad value [%s] using fallback: %d", text,
            fallback)
        ret = fallback
    }
    return ret
}

func AskForNumbers() [3]byte {
    reader := bufio.NewReader(os.Stdin)
    
    fmt.Println ("Enter in a value between 0 and 255 and press enter:")
    
    fmt.Print ("\nRed> ")
    red := ValueToNumber(reader, 255)
    
    fmt.Print ("\nGreen> ")
    green :=  ValueToNumber(reader, 255)
    
    fmt.Print ("\nBlue> ")
    blue :=  ValueToNumber(reader, 255)
    fmt.Println()
    
    return [3]byte{red, green, blue}
}

func FindCompliment(color [3]byte) [3]byte{
    color[0] = 255 - color[0]
    color[1] = 255 - color[1]
    color[2] = 255 - color[2]
    return color
}

func ExampleLine(color [3]byte) string {
    compliment := FindCompliment(color)
    color_text := ExpandColorText(color)
    compliment_text := ExpandColorText(compliment)
    
    return fmt.Sprintf ("%s->%s\n", color_text, compliment_text)
}

/******************************************************************************/
/* Command Tasks */

func main() {
    verbose := flag.Bool("verbose", false, "verbose")
    flag.Parse()

    app_data.verbose = *verbose
    
    color := AskForNumbers()
    fmt.Printf(ExampleLine(color))
}