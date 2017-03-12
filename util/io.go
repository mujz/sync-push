package util

import (
  "os"
  "bufio"
  "fmt"
  "path/filepath"
)

func ReadFromStdin(prompt string) (string, error) {
  reader := bufio.NewReader(os.Stdin)
  fmt.Print(prompt)
  text, err := reader.ReadString('\n')
  if err != nil {
    return "", err
  }
  return text, nil
}

func OpenOrCreate(path string, flag int, perm os.FileMode) (*os.File, bool, error) {
  isNew := false
  file, err := os.OpenFile(path, flag, perm)
  // if there was an error, create the file
  if err != nil {
    if os.IsNotExist(err) {
      file, err = os.Create(path)
      isNew = true
      if err != nil {
        return nil, false, err
      }
    } else {
      return nil, false, err
    }
  }
  return file, isNew, nil
}

func WalkDirs(handler func(path string) error) filepath.WalkFunc {
  return func(path string, info os.FileInfo, err error) error {
    if err != nil {
      return err
    }
    if info.IsDir() {
      if err := handler(path); err != nil {
        return err
      }
    }
    return nil
  }
}
