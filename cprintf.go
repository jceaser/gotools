package main

/*
A simple wrapper around printf to allow commands to format text with color and
styles.

cprintf --color red --style underline "text to format"

*/

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type ColorCode int
type StyleCode int
type StyleMap map[string]StyleCode
type ColorMap map[string]ColorCode

// Create styles the terminal can display
const (
    Normal StyleCode = iota + 0
    Bold
    Faint
    Italix      // not working
    Underline
    Blink       //5
    Fast        // not working
    Reverse
    Conceal
    Strike      //9
    Framed = iota + 41 //51
    Encircled
    Overlined
    IdeogramUnder = iota + 47 //60
    IdeogramUnderDouble
    IdeogramOver
    IdeogramOverDouble
    IdeogramStress
)

// Create colors to support RGB, CMYK, and Traffic lights
const (
	Black ColorCode = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	DefaultForground = 39
	DefaultBackground = 49
)

var (
    styles = StyleMap{
        //"normal": Normal,
        "bold": Bold,
        "faint": Faint,
        "italix": Italix,
        "underline": Underline,
        "blink": Blink,
        "fast": Fast,
        "reverse": Reverse,
        "conceal": Conceal,
        "strike": Strike,
        "framed": Framed,
        "encircled": Encircled,
        "overlined": Overlined,
        "ideogram-under": IdeogramUnder,
        "ideogram-under-double": IdeogramUnderDouble,
        "ideogram-over": IdeogramOver,
        "ideogram-over-double": IdeogramOverDouble,
        "ideogram-stress": IdeogramStress,
    }

    colors = ColorMap {
        "black": Black,
        "red": Red,
        "green": Green,
        "yellow": Yellow,
        "blue": Blue,
        "magenta": Magenta,
        "cyan": Cyan,
        "white": White,
        "default-forground": DefaultForground,
        "default-background": DefaultBackground,
    }
)

/*
Populate an additional key using the code value as a string for the same color
*/
func init() {
    //add to colors
    for _, value := range colors {
        colors[strconv.Itoa(int(value))] = value
    }

    //add to styles
    for _, value := range styles {
        styles[strconv.Itoa(int(value))] = value
    }
}

func (self StyleMap) reverse(code StyleCode) string {
    for name, value := range self {
        if value == code {
            return name
        }
    }
    return ""
}

func (self StyleMap) lookup(name string) StyleCode {
    style := Normal
    if len(name) > 0 {
        if code, okay := self[name]; okay {
            style = code
        } else {
            style = Normal
        }
    }
    return style
}

func (self ColorMap) reverse(code ColorCode) string {
    for name, value := range self {
        if value == code {
            return name
        }
    }
    return ""
}

func (c ColorMap) lookup(name string) ColorCode {
    color := White
    if len(name) > 0 {
        if code, okay := c[name]; okay {
            color = code
        } else {
            color = White
        }
    }
    return color
}

func (self ColorMap) asBackground(name string) int {
    forground := self.lookup(name)
    return int(forground) + 10
}

type AppConf struct {
    Color string
    Background string
    Style string
    Demo bool
}

/******************************************************************************/

/* Take a slice made up of anything and return a slice of 'any' values */
func unpackArray[SliceType ~[]Anything, Anything any](s SliceType) []any {
    r := make([]any, len(s))
    for i, e := range s {
        r[i] = e
    }
    return r
}

func sprintf(style StyleCode, color ColorCode, background int, line string) string {
    // \x1b or \033 is the ASCII escape number
    encoded := fmt.Sprintf("\033[%d;%d;%dm%s\033[0m", style, color, background, line)
    return encoded
}

func sortKeys[MapType ~map[string]Anything, Anything any](data MapType) []string {
    var keys []string
    for k := range data {

        if _, err := strconv.Atoi(k); err == nil {
            continue
        }

        keys = append(keys, k)
    }
    sort.Strings(keys)
    return keys
}

func RunDemo() {
    fmt.Println("Use these values for specifing --colors and --styles")

    background := colors.asBackground("default-background")

    fmt.Println(sprintf(Underline, Green, background, "Colors"))
    sortedColors := sortKeys(colors)

    for _, key := range sortedColors {
        if strings.HasPrefix(key, "default") {
            continue
        }
        value := colors[key]
        color := key
        fmt.Println(sprintf(Normal, value, background, "\uEE03\uEE04\uEE05 " + color))
    }

    fmt.Println(sprintf(Underline, Green, background, "\nStyles"))
    sortedStyles := sortKeys(styles)
    for _, key := range sortedStyles {
        value := styles[key]
        style := key
        fmt.Println(sprintf(value, White, background, "\uEE03\uEE04\uEE05 " + style))
    }

    //just to make sure, clean up
    fmt.Print("\033[0m")
    os.Exit(0)
}

func main() {
    //set up the application and it's parameters
    app := AppConf{}
    flag.StringVar(&app.Color, "color", "default-forground", "Forground/text color")
    flag.StringVar(&app.Background, "background", "default-background", "Background color")
    flag.StringVar(&app.Style, "style", "", "Text style")
    flag.BoolVar(&app.Demo, "demo", false, "Print out a demo")
    flag.Parse()

    if app.Demo {
        RunDemo()
    }

    remainingArgs := flag.Args()

    //translate the inputs
    color := colors.lookup(strings.ToLower(app.Color))
    background := colors.asBackground(strings.ToLower(app.Background))
    style := styles.lookup(strings.ToLower(app.Style))

    //process the format and the remaining parameters together
    format := remainingArgs[0]
    others := remainingArgs[1:]
    line := fmt.Sprintf(format, unpackArray(others)...)

    //fix new lines
    line = strings.ReplaceAll(line, "\\n", "\n")

    //apply the color and styles
    encoded := sprintf(style, color, background, line)
    //encoded := fmt.Sprintf("\033[%d;%dm%s\033[0m", style, color, line)

    fmt.Print(encoded)
}
