/*
This program uses the Logrus library for logging and the Watcher package for monitoring file system changes. It accepts the following command-line arguments:

    dir: Directory to monitor (default is the current directory)
    log: Log file path (default is monitor.log)
    events: Events to monitor separated by comma (default is create,write,remove,rename)
    extensions: File containing allowed extensions (default is allowed_extensions.txt)

The program reads the allowed extensions from the specified file and sets up a watcher to monitor the specified directory recursively. It filters the events based on the specified events and file extensions.

When an event occurs, the program checks if the file extension is allowed. If not, it logs an error message and sends a system notification. For allowed events, it logs the event details using Logrus.

*/


package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/gen2brain/beeep"
	"github.com/sirupsen/logrus"
)

// Define a custom type for the allowed extensions set
type AllowedExtensions map[string]struct{}

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Usage: go run filewatcher-v8.go <directory> <log file> <events> <allowed extensions file>")
		fmt.Println("OR")
		fmt.Println("Usage: go build -o birdeye.exe filewatcher-v8.go ")
		fmt.Println("birdeye.exe <directory> <log file> <events> <allowed extensions file>")
		os.Exit(1)
	}

	directory := os.Args[1]
	logFilePath := os.Args[2]
	events := strings.Split(os.Args[3], ",")
	allowedExtFilePath := os.Args[4]

	// Setup logrus
	logger := logrus.New()
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Could not open log file, logging to console instead:", err)
		logger.SetOutput(os.Stdout)
	} else {
		logger.SetOutput(logFile)
		defer logFile.Close()
	}

	// Load allowed extensions
	allowedExtensions, err := loadAllowedExtensions(allowedExtFilePath)
	if err != nil {
		logger.Fatalf("Failed to load allowed extensions: %v", err)
	}

	// Create a new watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatalf("Failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	// Add directory and subdirectories to watcher
	err = addWatchRecursive(watcher, directory)
	if err != nil {
		logger.Fatalf("Failed to add directory to watcher: %v", err)
	}

	// Start event processing
	go processEvents(watcher, logger, events, allowedExtensions)

	logger.Infof("Starting to monitor directory: %s", directory)
	select {} // Block forever
}

// Load allowed file extensions from a file
func loadAllowedExtensions(filePath string) (AllowedExtensions, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open allowed extensions file: %w", err)
	}
	defer file.Close()

	allowedExtensions := make(AllowedExtensions)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ext := strings.TrimSpace(scanner.Text())
		if ext != "" {
			allowedExtensions[ext] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading allowed extensions file: %w", err)
	}

	return allowedExtensions, nil
}

// Add a directory and its subdirectories to the watcher
func addWatchRecursive(watcher *fsnotify.Watcher, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				return fmt.Errorf("could not add directory to watcher: %w", err)
			}
		}
		return nil
	})
}

// Process file system events
func processEvents(watcher *fsnotify.Watcher, logger *logrus.Logger, events []string, allowedExtensions AllowedExtensions) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Check if the event type is in the specified list of events to monitor
			if containsEvent(events, event) {
				logger.Infof("Event: %s on file: %s", event.Op.String(), event.Name)
				
				if event.Op&fsnotify.Create == fsnotify.Create {
					// Check for allowed extensions
					if !isAllowedExtension(event.Name, allowedExtensions) {
						logger.Warnf("Disallowed file extension: %s", event.Name)
						err := handleDisallowedFile(event.Name)
						if err != nil {
							logger.Errorf("Failed to handle disallowed file: %v", err)
						}
					}
				}
			}

			// Add new directories to the watcher
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					err = watcher.Add(event.Name)
					if err != nil {
						logger.Errorf("Failed to add new directory to watcher: %v", err)
					}
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Errorf("Watcher error: %v", err)
		}
	}
}

// Check if the event type is in the list of events to monitor
func containsEvent(events []string, event fsnotify.Event) bool {
	for _, e := range events {
		switch strings.ToLower(e) {
		case "create":
			if event.Op&fsnotify.Create == fsnotify.Create {
				return true
			}
		case "write":
			if event.Op&fsnotify.Write == fsnotify.Write {
				return true
			}
		case "remove":
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				return true
			}
		case "rename":
			if event.Op&fsnotify.Rename == fsnotify.Rename {
				return true
			}
		case "chmod":
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				return true
			}
		}
	}
	return false
}

// Check if the file extension is allowed
func isAllowedExtension(fileName string, allowedExtensions AllowedExtensions) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	_, allowed := allowedExtensions[ext]
	return allowed
}

// Handle disallowed file by sending a notification and removing the file
func handleDisallowedFile(fileName string) error {
	err := os.Remove(fileName)
	if err != nil {
		if errors.Is(err, syscall.ENOENT) {
			return nil // File already removed or does not exist
		}
		return fmt.Errorf("could not remove disallowed file: %w", err)
	}

	err = beeep.Notify("Disallowed File Alert", fmt.Sprintf("A file with disallowed extension was created and removed: %s", fileName), "")
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

