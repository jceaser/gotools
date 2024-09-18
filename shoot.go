package main

import (
    "fmt"
    "math"
)

func create_buffer(x,y int) [][]rune {
    buffer := make([][]rune, x)
    for i := range buffer {
        buffer[i] = make([]rune, y)
    }
    return buffer
}

func init(){
    fmt.Println(len(create_buffer(80,24)[79]))
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
        buffer[x][y] = object
    }
}

/* convert human degrees to math radians */
func deg_to_rad(deg float64) float64 {
    return deg * (math.Pi/180)
}

func bar() string {
    bar := ""
    for x := 0; x < 80; x++ {
        bar = bar + "─"
    }
    return bar
}

/* draw the buffer onto the screen */
func draw(buffer [][]rune) {
    fmt.Printf("  ┌%s┐\n", bar())
    for y :=0; y < 24; y++ {
        fmt.Printf("%02d│", 24-y)
        for x := 0; x < 80; x++ {
            block := buffer[x][23-y]
            if block == 0 {
                block = ' '
            }
            fmt.Printf("%c", block)
        }
        fmt.Printf("│\n")
    }
    fmt.Printf("  └%s┘\n", bar())
}

func main() {
    //buffer := [80][24]rune{} //offscreen buffer to plot on
    buffer := create_buffer(80,24)

    velocity := 28.0    //how hard to shoot
    angle := 35.0       //angle to shoot at
    rads := deg_to_rad(angle)

    fmt.Printf("Fire: %.1f˚ with %.1f m/s:\n", angle, velocity)

    // plot all the projectile points
    for x := 0; x < 80; x++ {
        y := Shoot(float64(x), 9.8, velocity, rads, 0.0)
        if y<0.0 {
            break
        }
        Plot(buffer, int(x), int(y), '*')
    }
    draw(buffer)
}
