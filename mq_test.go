package sysv_mq

import (
	"fmt"
	"syscall"
	"testing"
)

func SampleMessageQueue(t *testing.T) *MessageQueue {
	config := new(QueueConfig)
	config.Key = SysVIPCKey
	config.MaxSize = 1024
	config.ProjId = 1
	config.Mode = IPC_CREAT | 0600

	mq, err := NewMessageQueue(config)

	if err != nil {
		t.Error(err)
	}

	return mq
}

func ExampleMessageQueue() {
	mq, err := NewMessageQueue(&QueueConfig{
		Key:     0xDEADBEEF,
		MaxSize: 1024,
		Mode:    IPC_CREAT | 0600,
	})
	if err != nil {
		fmt.Println(err)
	}

	err = mq.SendBytes([]byte("Hello World"), 1, IPC_NOWAIT)
	if err != nil {
		fmt.Println(err)
	}

	response, _, err := mq.ReceiveBytes(0, IPC_NOWAIT)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(response))
	// Output:
	// Hello World
}

func Test_SendMessage(t *testing.T) {
	mq := SampleMessageQueue(t)

	wired := "Narwhals and ice cream"

	err := mq.SendBytes([]byte(wired), 4, IPC_NOWAIT)

	if err != nil {
		t.Error(err)
	}

	response, mtype, err := mq.ReceiveBytes(0, IPC_NOWAIT)

	if err != nil {
		t.Error(err)
	}

	if mtype != 4 {
		t.Errorf("expected mtype 4, got: %d", mtype)
	}

	if wired != string(response) {
		t.Errorf("expected %s, got: %s", wired, response)
	}

	if 4 != mtype {
		t.Errorf("expected mtype 4, got: %d", mtype)
	}
}

func Test_CountMessages(t *testing.T) {
	mq := SampleMessageQueue(t)

	wired := "kitkat"

	messages, err := mq.Count()

	if messages != 0 {
		t.Errorf("expected empty queue, messages: %d\n", messages)
	}

	if err != nil {
		t.Error(messages)
	}

	if err := mq.SendString(wired, 1, IPC_NOWAIT); err != nil {
		t.Error(err)
	}

	messages, err = mq.Count()

	if messages != 1 {
		t.Errorf("expected empty queue, messages: %d\n", messages)
	}

	if err != nil {
		t.Error(messages)
	}

	response, _, err := mq.ReceiveString(0, IPC_NOWAIT)

	if err != nil {
		t.Error(err)
	}

	if wired != response {
		t.Errorf("expected %s, got: %s", wired, response)
	}

	messages, err = mq.Count()

	if messages != 0 {
		t.Errorf("expected empty queue, messages: %d\n", messages)
	}

	if err != nil {
		t.Error(messages)
	}
}

func Test_QueueClose(t *testing.T) {
	mq := SampleMessageQueue(t)
	mq.Close()
	mq.Close()
	mq.Close()
}

func Test_QueueDestroy(t *testing.T) {
	mq := SampleMessageQueue(t)

	if mq2, err := NewMessageQueue(&QueueConfig{Key: SysVIPCKey}); err != nil {
		t.Fatal(err)
	} else {
		mq2.Close()
	}

	mq.Destroy()

	mq3, err := NewMessageQueue(&QueueConfig{Key: SysVIPCKey})
	switch err {
	case nil:
		mq3.Close()
		t.Fatal("Expected opening of MQ to fail after it has been destroyed, but it succeeded.")
	case syscall.ENOENT:
		t.Log("SUCCESS: failed to open MQ as expected")
	default:
		t.Fatal(err)
	}
}

func Test_QueueSize(t *testing.T) {
	mq := SampleMessageQueue(t)

	wired := "koalas"

	size, err := mq.Size()

	if err != nil {
		t.Error(err)
	}

	if size != 0 {
		t.Error("expected empty queue")
	}

	if err = mq.SendString(wired, 1, 0); err != nil {
		t.Error(err)
	}

	size, err = mq.Size()

	if err != nil {
		t.Error(err)
	}

	if size != uint64(len(wired)) {
		t.Errorf("expected queue of len %d queue, got: %d\n", len(wired), size)
	}

	response, _, err := mq.ReceiveString(0, 0)

	if err != nil {
		t.Error(err)
	}

	if wired != response {
		t.Errorf("expected %s, got: %s", wired, response)
	}

	size, err = mq.Size()

	if err != nil {
		t.Error(err)
	}

	if size != 0 {
		t.Error("expected empty queue")
	}
}
