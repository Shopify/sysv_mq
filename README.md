# Go wrapper for SysV Message Queues

sysv_mq is a Go wrapper for [SysV Message
Queues](). It's important you
read the [manual for SysV Message Queues][overview], [msgrcv(2)][rcvsnd] and
[msgsnd(2)][rcvsnd] before using this library. `sysv_mq` is a very light
wrapper, and will not hide any errors from you.

Documentation for the public API can be viewed at [Godoc][godoc].

sysv_mq is tested on Linux and OS X. To run the tests run `make test`. This
makes sure that any messages queues currently on your system are deleted before
running the tests.

## Example

Example which sends a message to the queue with key: `0xDEADBEEF` (or creates it
if it doesn't exist).

```go
package main

import (
	"fmt"
	"github.com/Shopify/sysv_mq"
)

func main() {
	mq, err := sysv_mq.NewMessageQueue(&sysv_mq.QueueConfig{
		Key:     0xDEADBEEF,               // SysV IPC key
		MaxSize: 1024,                     // Max size of a message
		Mode:    sysv_mq.IPC_CREAT | 0600, // Creates if it doesn't exist, 0600 permissions
	})
	if err != nil {
		fmt.Println(err)
	}

	// Send a message to the queue
	err = mq.Send("Hello World", 1)
	if err != nil {
		fmt.Println(err)
	}

	// Receive a message from the queue, 0 gives you the top message regardless of
	// message type passed to send().
	response, err := mq.Receive(0)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)
	// Output:
	// Hello World
}
```

## Caveats

* `Send()` and `Receive()` are (by default) blocking. Right now it doesn't
  support non-blocking calls (it's supported in the C -> Go wrapper,
  `wrapper.go`, but not currently exposed in the public interface.)
* `Send()`, `Receive()` and `NewMessageQueue()` all do syscalls, and these could
  be interrupted by a signal (very common if you do a blocking `Receive()`). The
  error will be `EAGAIN` in that case. It's not wrapped here, because `EAGAIN`
  is also the error if the call would block or the queue is full. Consult the
  manual for more information.

[overview]: http://man7.org/linux/man-pages/man7/svipc.7.html
[rcvsnd]: http://man7.org/linux/man-pages/man2/msgrcv.2.html
[godoc]: http://godoc.org/github.com/Shopify/sysv_mq
