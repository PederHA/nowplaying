/*
Based off of this Playground example: https://play.golang.org/p/YfGDtIuuBw

Basic "Now Playing" application for Windows that can be used to log the currently
playing Spotify track without using the web API.
Useful for overlaying "Now Playing" data on a streaming broadcast.

TODO:
	1. Use init() to set up global flag variables
	2. Add prefix and suffix flags for file output (ex. -pre [ -suf ])
	3. Add attempts flag, so program terminates if window cannot be found after N tries
*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32             = syscall.MustLoadDLL("user32.dll")
	procEnumWindows    = user32.MustFindProc("EnumWindows")
	procGetWindowTextW = user32.MustFindProc("GetWindowTextW")
)

func enumWindows(enumFunc uintptr, lparam uintptr) (err error) {
	r1, _, e1 := syscall.Syscall(procEnumWindows.Addr(), 2, uintptr(enumFunc), uintptr(lparam), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func getWindowText(hwnd syscall.Handle, str *uint16, maxCount int32) (len int32, err error) {
	r0, _, e1 := syscall.Syscall(procGetWindowTextW.Addr(), 3, uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	len = int32(r0)
	if len == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func findWindow(title string, callback uintptr, handle *syscall.Handle) (syscall.Handle, error) {
	enumWindows(callback, 0)
	if *handle == 0 {
		return 0, fmt.Errorf("No window with title '%s' found", title)
	}
	return *handle, nil
}

func windowTextToString(handle syscall.Handle, str *uint16, maxCount int32) (string, error) {
	b := make([]uint16, 200)
	_, err := getWindowText(handle, &b[0], int32(len(b)))
	if err != nil {
		// ignore the error
		return "", fmt.Errorf("Unable to identify window title. Did you close the window?")
	}
	windowTitle := syscall.UTF16ToString(b)
	return windowTitle, nil
}

/* Continually polls the identified Spotify window until an error occurs */
func pollSpotifyWindow(handle syscall.Handle, sleepDurationSec int, npPath string) error {
	var prevTitle string
	var b = make([]uint16, 200) // Holds window title byte array (or w/e it's called in Go)

	for true {
		windowTitle, err := windowTextToString(handle, &b[0], int32(len(b)))
		if err != nil {
			return err
		}
		// Only write artist & song to file if new window title is different
		if len(prevTitle) == 0 || prevTitle != windowTitle {
			logSpotifyWindow(windowTitle, npPath)
			prevTitle = windowTitle
		}
		time.Sleep(time.Duration(sleepDurationSec) * time.Second)
	}
	return nil
}

func logSpotifyWindow(windowTitle string, npPath string) {
	data := []byte(windowTitle)
	err := ioutil.WriteFile(npPath, data, 0644)
	if err != nil {
		log.Fatal(err) // Terminate if we are unable to write to the file
	}
}

/*
Creates a callback function using a pointer to the handle defined in main().
There are probably 100 ways to do this in more readable and efficient ways, but I am awful at Go.
*/
func makeCallback(handle *syscall.Handle) uintptr {
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		b := make([]uint16, 200)
		_, err := getWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			// ignore the error
			return 1 // continue enumeration
		}
		windowTitle := syscall.UTF16ToString(b)
		if strings.HasPrefix(windowTitle, "Spotify") {
			// note the window
			*handle = h
			return 0 // stop enumeration
		}
		return 1 // continue enumeration
	})
	return cb
}

func main() {
	// NOTE: This should probably all be wrapped in log.Fatal() or os.Exit()?

	// Parse flags
	title := flag.String("t", "Spotify", "Window title to look for") // Maybe cba this?
	pollInterval := flag.Int("n", 1, "Polling interval (sec)")
	flag.Parse()

	// Parse non-flag arguments
	npPath := "nowplaying.txt"
	args := flag.Args()
	if len(args) > 0 {
		npPath = args[0]
	}

	var handle syscall.Handle // Our window handle that we pass around (Probably REALLY bad)
	callback := makeCallback(&handle)

	// Main program flow
	for true {
		handle, err := findWindow(*title, callback, &handle)
		if err != nil {
			continue
		}
		// pollSpotifyWindow() loops forever until it hits an error
		// In which case the error is logged and we start from the beginning
		// of this loop again, and try to find a new Spotify window handle.
		err = pollSpotifyWindow(handle, *pollInterval, npPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}
