package main
// TODO FIX watcher firing after every other event, not after every event

import (
  "bufio"
  "io/ioutil"
  "os"
  "os/exec"
  "path/filepath"
  "log"
  "fmt"
  "reflect"
  "strings"
  "time"

  "github.com/fsnotify/fsnotify"
)

const LOCATIONS_FILE string = "/locations.ini"
const PUSH_CMD string = "rsync"
var PUSH_OPTS map[string]bool = map[string]bool {
  "--delete": true,
}
var GENERAL_OPTS map[string]func() = map[string]func() {
  "help": printHelp,
  "--version": printVersion,
}
var IGNORE_FLAG map[bool][]string = map[bool][]string{
  true: {"--exclude-from", ".syncignore"},
  false: {},
}

func printVersion() {
  log.Println("sync-watch version 0.0.1")
}

func printHelp() {
  usage := "Usage: sync-push [options]\n"
  description := "Watch and sync files from current directory to a remote directory"
  options := "help\tprint this message\n"
  options += "--delete\tdelete extraneous files from destination dirs\n"
  options += "--version\tprint version number\n"
  // TODO check -C, --cvs-exclude
  log.Println(usage + description + options)
}

func panicIfErr(err error) {
  if err != nil {
    panic(err)
  }
}

func readRemoteFromStdin() string {
  reader := bufio.NewReader(os.Stdin)
  fmt.Print("Remote Location (eg. user@host:/path/to/remote/dir): ")
  text, err := reader.ReadString('\n')
  panicIfErr(err)
  return text
}

func readRemoteFromLocations(locations, local string, localIndex int) string {
  remote := locations[localIndex + len(local) + 1:]
  remote = remote[:strings.Index(remote, "\n")]
  return remote
}

func getDirs() (local, remote string) {
  // TODO find a better place for locations
  // open locations.ini
  locationsAbsPath := os.Getenv("HOME") + LOCATIONS_FILE
  file, err := os.OpenFile(locationsAbsPath, os.O_APPEND|os.O_RDWR, os.ModeAppend)
  // if there was an error, create the file
  if err != nil {
    file, err = os.Create(locationsAbsPath)
    panicIfErr(err)
  }

  // read the file
  data, err := ioutil.ReadAll(file)
  panicIfErr(err);
  locations := string(data)

  // get the current working directory and append it with a space
  local, err = os.Getwd()
  panicIfErr(err);
  local += " "

  // get the current directory's remote match either from the locations file or from stdin
  defer file.Close()
  localIndex := strings.Index(locations, local);
  if localIndex == -1 {
    remote = readRemoteFromStdin()
    _, err = file.WriteString(local + " " + remote);
    panicIfErr(err)
  } else {
    remote = readRemoteFromLocations(locations, local, localIndex);
  }

  return local[:len(local) - 1], remote
}

// TODO watch when a new dir is created to add a watcher for it
func dirWalker(handler func(path string) error) filepath.WalkFunc {
  return func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }
    if info.IsDir() {
      if err := handler(path); err != nil {
        fmt.Print("Error: ")
        fmt.Println(err)
      }
    }
    return nil
  }
}

// idea from https://github.com/eapache/channels/blob/master/channels.go#L234
func getChan(i interface{}) chan interface{} {
  ch := make(chan interface{})
  go func() {
    val, ok := reflect.ValueOf(i).Recv()
    if !ok {
      close(ch)
      return
    }
    ch <- val.Interface()
  }()
  return ch
}

// idea from here: https://nathanleclaire.com/blog/2014/08/03/write-a-function-similar-to-underscore-dot-jss-debounce-in-golang/
func debounce(delay time.Duration, event interface{}, eventEmitter interface{}, handler func(interface{})) {
  eventEmitterChan := getChan(eventEmitter)
  select {
  case event = <-eventEmitterChan: /* do nothing */
  case <-time.After(delay):
    handler(event)
  }
}

func push(local, remote string, shouldIgnore bool, options []string) func(interface{}) {
  return func(e interface{}) {
    if (e != nil) {
      fmt.Println(e.(fsnotify.Event))
    }
    cmdOptions := []string{"-aiz", local, remote}
    if len(options) > 0 {
      cmdOptions = append(cmdOptions, options...)
    }
    if len(IGNORE_FLAG) > 0 {
      cmdOptions = append(cmdOptions, IGNORE_FLAG[shouldIgnore]...)
    }

    out, err := exec.Command(PUSH_CMD, cmdOptions...).Output()
    panicIfErr(err)
    fmt.Println(string(string(out)))
  }
}

func watch(path string, handler func(interface{})) error {
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    return err
  }
  defer watcher.Close()

  done := make(chan bool)
  go func() {
    for {
      select {
      case event := <-watcher.Events:
        debounce(100 * time.Millisecond, event, watcher.Events, handler)
      case err := <-watcher.Errors:
        fmt.Println("error:", err)
      }
    }
  }()

  err = watcher.Add(path)
  panicIfErr(err)
  err = filepath.Walk(path, dirWalker(watcher.Add))
  panicIfErr(err)

  <-done
  return nil
}


func checkPushCmd() {
  if err := exec.Command(PUSH_CMD, "--version").Run(); err != nil {
    log.Fatal("rsync is not installed. For information on how to install it on your OS, ask Google")
  }
}

func main() {
  // If rsync is not installed, exit immediately.
  checkPushCmd()

  // TODO read command line arguments: remote -v, remote set, remote remove
  pushOpts := make([]string, 0)
  for i, opt := range os.Args[1:] {
    if _, ok := PUSH_OPTS[opt]; ok {
      pushOpts = append(pushOpts, os.Args[i+1])
    } else if handler, ok := GENERAL_OPTS[opt]; ok {
      handler()
    } else {
      fmt.Printf("Option %s is not recognized. Please use one of the following options\n", opt)
      printHelp()
      os.Exit(1)
    }
  }

  local, remote := getDirs()

  shouldIgnore := true
  if _, err := os.Stat(local + "/.syncignore"); os.IsNotExist(err) {
    shouldIgnore = false
  }

  push(local, remote, shouldIgnore, pushOpts)(nil)

  watch(local, push(local, remote, shouldIgnore, pushOpts))
}
