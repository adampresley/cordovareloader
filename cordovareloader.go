/*
Copyright (c) 2015 Adam Presley
Licensed under the MIT license
*/

package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/adampresley/sigint"
)

func main() {
	log.SetPrefix("Cordova Reloader - ")
	log.Println("Setting up...")

	setupShutdownHandler()

	var cordovaHandle *exec.Cmd
	var err error
	var startTime time.Time

	/*
	 * Loop forever. First we get the start time so we can compare
	 * file modify times against it. Then we run 'cordova prepare'
	 * and 'cordova serve'. We start a directory walker than, when a
	 * date doesn't match, kills the serve process, and the whole
	 * thing starts over.
	 */
	for {
		startTime = time.Now()

		err = runCordovaPrepare()
		if err != nil {
			log.Println("There was an error calling 'cordova prepare':", err.Error())
			os.Exit(1)
		}

		cordovaHandle, err = startCordovaServe()
		if err != nil {
			log.Println("There was an error calling 'cordova serve':", err.Error())
			os.Exit(2)
		}

		scanDirectory(startTime, cordovaHandle)
		cordovaHandle.Wait()

		log.Println("File changed. Reloading...")
	}
}

func runCordovaPrepare() error {
	log.Println("Running 'prepare'...")

	cmd := exec.Command("cordova", "prepare")
	return cmd.Run()
}

func scanDirectory(startTime time.Time, handle *exec.Cmd) {
	log.Println("Listening for directory changes...")

	/*
	 * Walk the directory tree and scan for changes
	 */
	go func() {
		for {
			filepath.Walk("./www", func(path string, info os.FileInfo, err error) error {
				if info.ModTime().After(startTime) {
					handle.Process.Signal(os.Interrupt)
					return errors.New("done")
				}

				return nil
			})

			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func setupShutdownHandler() {
	sigint.Listen(func() {
		log.Println("Shutting down...")
		os.Exit(0)
	})
}

func startCordovaServe() (*exec.Cmd, error) {
	log.Println("Starting 'serve'...")

	cmd := exec.Command("cordova", "serve")
	err := cmd.Start()

	return cmd, err
}
