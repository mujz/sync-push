package util

import (
  "reflect"
)

// idea from https://github.com/eapache/channels/blob/master/channels.go#L234
// return a receive channel from an interface
func getReceiveChannel(i interface{}) chan interface{} {
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

