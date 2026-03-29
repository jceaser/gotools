package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"io"

	"golang.org/x/term"
	"github.com/creack/pty"
)

func main() {
	// Save original terminal state
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Create PTY and child process
	cmd := exec.Command("vim")
	ptmx, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = ptmx.Close()
	}()

	// Handle resize signals
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			//pty.InheritSize(os.Stdin, ptmx)
			resizePTY(ptmx, 1)
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize

    top := false
	// Show a simple status bar
	go func() {
		for {
			time.Sleep(1 * time.Second)
			if top {
    			fmt.Fprintf(os.Stdout,
    			    "\x1b7\x1b[1;1HStatus: %s\n\x1b[K\x1b8",
    			    time.Now().Format("15:04:05"))
            } else {
	    		/*fmt.Fprintf(os.Stdout,
	    		    //"\x1b7\x1b[%d;1HStatus: %s\x1b8",
		    	    "\x1b7\x1b[%d;1H\x1b[30;42mStatus: %s\x1b[0m\x1b8",
			        getTerminalHeight(),
			        time.Now().Format("15:04:05"))*/

			    /*msg := fmt.Sprintf("\x1b[30;42mStatus: \x1b[31;42m%s\x1b[0m",
			        time.Now().Format("15:04:05"))*/

			    msg := fmt.Sprintf("Status: %s", time.Now().Format("15:04:05"))
			    drawStatusBar(getTerminalHeight(), msg)
			}
		}
	}()

	// Relay input/output
	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
	}()
	_, _ = io.Copy(os.Stdout, ptmx)
}

func drawStatusBar(row int, message string) {
	width, _, _ := term.GetSize(int(os.Stdout.Fd()))
	padding := width - len(message)
	if padding < 0 {
		padding = 0
	}
	fmt.Fprintf(os.Stdout,
		"\x1b7\x1b[%d;1H\x1b[30;42m%s%s\x1b[0m\x1b8",
		row, message, strings.Repeat(" ", padding))
}

func getTerminalHeight() int {
	_, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 24 // fallback
	}
	return h
}

func resizePTY(ptmx *os.File, reserveBottomLines int) {
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err == nil && h > reserveBottomLines {
		pty.Setsize(ptmx, &pty.Winsize{
			Cols: uint16(w),
			Rows: uint16(h - reserveBottomLines),
		})
	}
}
