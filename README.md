# gotools
A collection of command line tools written in Go

* console - get the size of a console
* counter - count up or down with big numbers
* date-math - find difference between two dates
* db - simple name/value storage
* db2 - complex data storage with forms ; in progress
* filter - filter out matching lines from regexp
* iif - inline if for scripting
* isodate - current date in iso format because date stinks
* md2html - template based HTML generator from markdown source
* params - a tool for processing url parameters and parts
* roll - rotate files like logs ; in progress
* rpn - an RPN calculator
* scale - console line chart scroler ; experimental
* smatrix - like cmatrix but a little basic ; experimental
* striper - remove spaces from lines or words
* thru - Marshal-Unmarshal example

## Details ##

### iif ###
Inline If
returns a value if the left and right operands meet a condition

Usage:
    iff -left 10 -test '==' -right 10

### rpn ###
A Reverse Polish Notation calculator

    rpn --interactive
    rpn --formula 'formula'
    echo 'formula' | rpn
        --pop return only last value
        --verbose print out some extra details

Use: `2 3 + print`

### scale ###

Usage:

echo 10 20 30 40 50 60 50 40 30 20 10 | scale -wait 10

    ↑⎺↑⎺↑⎺↑⎺↑⎺↑⎺⎻──⎼⎽
    floor: 0.000000, avg: 32.727273, ceil: 60.000000
    w=18, c=11

Arrows show a scale change, you can avoid them by setting a ceil or floor value through command line flags.

