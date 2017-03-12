package util
import (
  "time"
)

// idea from https://nathanleclaire.com/blog/2014/08/03/write-a-function-similar-to-underscore-dot-jss-debounce-in-golang/
func Debounce(delay time.Duration, event interface{}, eventEmitter interface{}, handler func(interface{})) {
  eventEmitterChan := getReceiveChannel(eventEmitter)
  select {
  case event = <-eventEmitterChan: /* do nothing */
  case <-time.After(delay):
    handler(event)
  }
}
