/* Filename: filewatcher-v1.go
Steps followed:
1) Program takes `path to directory` as an argument.
2) Then it checks if directory exist at the given path or not.Then it checks if it is a directory.
3) Converts the given path to absolute path
4) Sets filter operations for notifying file changes such as creation, deletion, modification & renaming.
5) It will recursively watch any sub-directories.
	\-> It will give an issue: If you create any new file inside sub-directory it will give you a result of 
	modification instead of creation.
*/
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/radovskyb/watcher"
)

func main() {
    // Define a command-line flag for the directory path.
    // `flag` package is used to accespt a directory path from the command line.
    // If no path is given, it defaults to the current directory (.)
    dirPath := flag.String("path", ".", "Directory to watch")
    flag.Parse()

    // Check if the provided path is a directory.
    info, err := os.Stat(*dirPath)
    if os.IsNotExist(err) {
        log.Fatalf("Directory does not exist: %s", *dirPath)
    }
    if !info.IsDir() {
        log.Fatalf("Provided path is not a directory: %s", *dirPath)
    }

    // Convert given directory path to an absolute path.
    absPath, err := filepath.Abs(*dirPath)
    if err != nil {
        log.Fatalf("Could not determine absolute path: %v", err)
    }

    // Create a new watcher.
    w := watcher.New()

    // Set max events. This is take care of maximum number of events to be processes at once to avoid flooding.
    w.SetMaxEvents(1)

    // Only notify for file changes.
    w.FilterOps(watcher.Create, watcher.Remove, watcher.Rename, watcher.Move, watcher.Write)

    // Start the watching process with help of goroutines
    go func() {
        for {
            select {
            case event := <-w.Event:
                switch event.Op {
                case watcher.Create:
                    fmt.Println("File created:", event.Path)
                case watcher.Remove:
                    fmt.Println("File removed:", event.Path)
                case watcher.Rename:
                    fmt.Println("File renamed:", event.Path)
                case watcher.Move:
                    fmt.Println("File moved:", event.Path)
                case watcher.Write:
                    fmt.Println("File modified:", event.Path)
                }
            case err := <-w.Error:
                log.Println("Error:", err)
            case <-w.Closed:
                return
            }
        }
    }()

    // Add the directory to the watcher.
    // This will take care of specified directory and its sub-directories are added to the watcher.
    if err := w.AddRecursive(absPath); err != nil {
        log.Fatalf("Error adding directory to watcher: %v", err)
    }

    fmt.Printf("Watching directory: %s\n", absPath)

    // Start the watcher with a polling interval.
    if err := w.Start(time.Millisecond * 100); err != nil {
        log.Fatalf("Error starting watcher: %v", err)
    }
}

