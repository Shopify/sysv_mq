package sysv_mq

import (
	"fmt"
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

	err = mq.Send("Hello World", 1)
	if err != nil {
		fmt.Println(err)
	}

	response, err := mq.Receive(0)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response)
	// Output:
	// Hello World
}

func Test_SendMessage(t *testing.T) {
	mq := SampleMessageQueue(t)

	wired := "Narwhals and ice cream"

	err := mq.Send(wired, 1)

	if err != nil {
		t.Error(err)
	}

	response, err := mq.Receive(0)

	if err != nil {
		t.Error(err)
	}

	if wired != response {
		t.Error("expected %s, got: %s", wired, response)
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

	if err := mq.Send(wired, 1); err != nil {
		t.Error(err)
	}

	messages, err = mq.Count()

	if messages != 1 {
		t.Errorf("expected empty queue, messages: %d\n", messages)
	}

	if err != nil {
		t.Error(messages)
	}

	response, err := mq.Receive(0)

	if err != nil {
		t.Error(err)
	}

	if wired != response {
		t.Error("expected %s, got: %s", wired, response)
	}

	messages, err = mq.Count()

	if messages != 0 {
		t.Errorf("expected empty queue, messages: %d\n", messages)
	}

	if err != nil {
		t.Error(messages)
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

	if err = mq.Send(wired, 1); err != nil {
		t.Error(err)
	}

	size, err = mq.Size()

	if err != nil {
		t.Error(err)
	}

	if size != uint64(len(wired)) {
		t.Errorf("expected queue of len %d queue, got: %d\n", len(wired), size)
	}

	response, err := mq.Receive(0)

	if err != nil {
		t.Error(err)
	}

	if wired != response {
		t.Error("expected %s, got: %s", wired, response)
	}

	size, err = mq.Size()

	if err != nil {
		t.Error(err)
	}

	if size != 0 {
		t.Error("expected empty queue")
	}
}
