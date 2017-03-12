package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mujz/sync-push/util"
)

const ENTER_REMOTE_PROMPT = "Remote Location (eg. user@host:/path/to/remote/dir): "
const LOCATIONS_FILE string = "/locations.ini"
const IGNORE_FILE_NAME string = ".syncignore"
const PUSH_CMD string = "rsync"

var PUSH_OPTS map[string][]string = map[string][]string{
	"--delete": {"--delete", "--force"},
}
var GENERAL_OPTS map[string]func() = map[string]func(){
	"help":      printHelp,
	"--version": printVersion,
}

func printVersion() {
	fmt.Println("sync-watch version 0.0.1")
	os.Exit(0)
}

func printHelp() {
	usage := "Usage: sync-push [options]"
	description := "Watch and sync files from current directory to a remote directory"
	options := "help\t\t\tprint this message\n"
	options += "--delete\t\tdelete extraneous files from destination dirs\n"
	options += "--version\t\tprint version number"
	// TODO check -C, --cvs-exclude
	fmt.Printf("%s\n%s\n\nOptions:\n%s\n", usage, description, options)
	os.Exit(0)
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getDirs() (local, remote string) {
	// get the current working directory
	local, err := os.Getwd()
	panicIfErr(err)

	// TODO find a better place for locations
	// open locations.ini
	locationsAbsPath := os.Getenv("HOME") + LOCATIONS_FILE
	file, isNew, err := util.OpenOrCreate(locationsAbsPath, os.O_APPEND|os.O_RDWR, os.ModeAppend)
	panicIfErr(err)

	defer file.Close()
	if !isNew {
		// read the file
		data, err := ioutil.ReadAll(file)
		panicIfErr(err)
		locations := string(data)

		// get the current directory's remote match either from the locations file or from stdin
		localIndex := strings.Index(locations, local+" ")
		if localIndex >= 0 {
			//remote = util.ReadRemoteFromLocations(locations, local, localIndex);
			remote = locations[localIndex+len(local)+2:]
			remote = remote[:strings.Index(remote, "\n")]
			return local, remote
		}
	}

	remote, err = util.ReadFromStdin(ENTER_REMOTE_PROMPT)
	panicIfErr(err)
	_, err = file.WriteString(local + " " + remote)
	panicIfErr(err)

	return local, remote
}

func push(local, remote string, options []string) func(interface{}) {

	// read the command options into cmdOptions
	cmdOptions := []string{"-aiz", local, remote}
	if len(options) > 0 {
		cmdOptions = append(cmdOptions, options...)
	}
	// if the ignore file exists, append it to the command options
	if _, err := os.Stat(local + "/" + IGNORE_FILE_NAME); !os.IsNotExist(err) {
		cmdOptions = append(cmdOptions, "--exclude-from", IGNORE_FILE_NAME)
	}

	return func(e interface{}) {
		out, err := exec.Command(PUSH_CMD, cmdOptions...).CombinedOutput()
		if len(out) > 0 {
			fmt.Println(string(string(out)))
		}
		panicIfErr(err)
	}
}

func watch(path string, handler func(interface{})) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// if a new dir is created, watch it
				if event.Op&fsnotify.Create == fsnotify.Create {
					info, err := os.Lstat(event.Name)
					panicIfErr(err)
					if info.IsDir() {
						err := watcher.Add(event.Name)
						panicIfErr(err)
					}
				}

				// debounce the handler
				util.Debounce(10*time.Millisecond, event, watcher.Events, handler)
			case err := <-watcher.Errors:
				fmt.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	panicIfErr(err)
	err = filepath.Walk(path, util.WalkDirs(watcher.Add))
	panicIfErr(err)

	wg.Wait()
	return nil
}

func checkPushCmd() {
	if err := exec.Command(PUSH_CMD, "--version").Run(); err != nil {
		fmt.Printf("%s is not installed. For information on how to install it on your OS, ask Google\n", PUSH_CMD)
		os.Exit(1)
	}
}

func main() {
	// If rsync is not installed, exit immediately.
	checkPushCmd()

	// TODO read command line arguments: remote -v, remote set, remote remove
	pushOpts := make([]string, 0)
	for _, opt := range os.Args[1:] {
		if pushOpt, ok := PUSH_OPTS[opt]; ok {
			pushOpts = append(pushOpts, pushOpt...)
		} else if handler, ok := GENERAL_OPTS[opt]; ok {
			handler()
		} else {
			fmt.Printf("Option %s is not recognized. call \"sync-push help\" for a list of valid options\n", opt)
		}
	}

	local, remote := getDirs()

	push(local, remote, pushOpts)(nil)

	watch(local, push(local, remote, pushOpts))
}
