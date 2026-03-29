package main

import (
    "flag"
    "fmt"
    "math"
    "os"
)

/* ************************************************************************** */
// MARK: - Functions

/* convert human degrees to math radians */
func deg_to_rad(deg float64) float64 {
    return deg * (math.Pi/180)
}

func create_buffer(y, x int) [][]rune {
    buffer := make([][]rune, y)
    for i := range buffer {
        buffer[i] = make([]rune, x)
    }
    return buffer
}

/* Plot the y value for a given x */
func Shoot(x, gravity, velocity, angle, height float64) float64 {
    t := x / (velocity * math.Cos(angle))
    y := velocity * math.Sin(angle) * t - (0.5 * gravity * math.Pow(t, 2))
    y += height
    return y
}

/* Place a rune on the buffer plot */
func Plot(buffer [][]rune, x, y int, object rune) {
    if 0 <= y && y<24 && 0 <= x && x < 80 {
        buffer[y][x] = object
    }
}

/* create a border for the table */
func bar(width int) string {
    bar := ""
    for x := 0; x < width; x++ {
        bar = bar + "─"
    }
    return bar
}

/* draw the buffer onto the screen */
func draw(buffer [][]rune) {
    fmt.Printf("  ┌%s┐\n", bar(len(buffer[y])))
    for y:=len(buffer)-1; y >= 0; y-- {
        fmt.Printf("%02d│", y+1) //printing for humans is 1 off

        for x := 0; x < len(buffer[y]); x++ {
            block := buffer[y][x]
            if block == 0 {
                block = ' '
            }
            fmt.Printf("%c", block)
        }
        fmt.Printf("│\n")
    }
    fmt.Printf("  └%s┘\n", bar())
}

/* ************************************************************************** */
// MARK: - App functions

func HelpMessageCallback() {
    fmt.Fprintf(flag.CommandLine.Output(),
        "shoot by thomas.cherry@gmail.com\n\n")
    fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
    flag.PrintDefaults()
}

type AppData struct {
    Angle *float64
    Velocity *float64
    Gravity *float64
    Columns *int
    Rows *int
    X *int
    Y *int
}

func GenerateAppData() AppData {
    flag.Usage = HelpMessageCallback

    appData := AppData{}

    appData.Angle = flag.Float64("angle", 45.0, "Projectile Angle in deg")
    appData.Velocity = flag.Float64("velocity", 32.0, "Projectile Velocity in m/s")
    appData.Gravity = flag.Float64("gravity", 9.8, "Gravity in m/s")
    appData.Columns = flag.Int("col", 80, "Screen Columns")
    appData.Rows = flag.Int("row", 24, "Screen Rows")
    appData.X = flag.Int("x", 0, "Starting X")
    appData.Y = flag.Int("y", 0, "Starting y")

    flag.Parse()
    return appData
}

func main() {
    appData := GenerateAppData()

    buffer := create_buffer(*appData.Rows, *appData.Columns)

    velocity := *appData.Velocity    //how hard to shoot
    angle := *appData.Angle       //angle to shoot at
    rads := deg_to_rad(angle)

    fmt.Printf("Fire: %.1f˚ with %.1f m/s:\n", angle, velocity)

    // plot all the projectile points
    for x := *appData.X; x < *appData.Columns; x++ {
        y := Shoot(float64(x), *appData.Gravity, velocity, rads, float64(*appData.Y))
        if y<0.0 {
            break
        }
        Plot(buffer, int(x), int(y), '*')
    }
    draw(buffer)
}
