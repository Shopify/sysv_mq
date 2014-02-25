package sysv_mq

import (
	"bytes"
	"sync"
	"syscall"
	"testing"
)

var path string = ""

const SysVIPCKey = 0x12345

func GetTestQueueId(t *testing.T) int {
	mode := IPC_CREAT | syscall.O_RDWR | syscall.S_IRUSR | syscall.S_IWUSR

	id, err := msgget(SysVIPCKey, mode)

	if err != nil {
		t.Error(err)
	}

	return id
}

func Test_FtokWithNonExistantPath(t *testing.T) {
	_, err := ftok("/i dont exist", 1)

	if err == nil {
		t.Error("expected error when calling ftok on non-existant path")
	}
}

func Test_MsggetWithBadKeyErrors(t *testing.T) {
	_, err := msgget(42, 0)

	if err == nil {
		t.Error(err)
	}
}

func Test_SendAndReceiveMessage(t *testing.T) {
	id := GetTestQueueId(t)

	wired := "hello world"

	buffer, err := allocateBuffer(len(wired))

	if err != nil {
		t.Error(err)
	}

	err = msgsnd(id, []byte("hello world"), buffer, len(wired), 1, 0)

	if err != nil {
		t.Error(err)
	}

	msg, _, err := msgrcv(id, 0, buffer, len(wired), 0)

	if err != nil {
		t.Error(err)
	}

	if string(msg) != "hello world" {
		t.Error("expected hello world, got: ", msg)
	}
}

func Test_Async(t *testing.T) {
	wired := "walrus tea party"

	id := GetTestQueueId(t)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		buffer, err := allocateBuffer(len(wired))

		if err != nil {
			t.Error(err)
		}

		msg, _, err := msgrcv(id, 0, buffer, len(wired), 0)

		if err != nil {
			t.Error(err)
			return
		}

		if string(msg) != wired {
			t.Errorf("expected %s, got: %s", wired, msg)
		}

		wg.Done()
	}()

	buffer, err := allocateBuffer(len(wired))

	if err != nil {
		t.Error(err)
	}

	msgsnd(id, []byte(wired), buffer, len(wired), 1, 0)
	wg.Wait()
}

func Test_MassAsync(t *testing.T) {
	id := GetTestQueueId(t)

	var wg sync.WaitGroup

	wired := "walrusser and unicorns"

	N := 50000

	wg.Add(2)

	go func() {
		buffer, err := allocateBuffer(len(wired))

		if err != nil {
			t.Error(err)
		}

		for i := 0; i < N; i++ {
			msg, _, err := msgrcv(id, 0, buffer, len(wired), 0)

			if err != nil {
				t.Error(err)
				return
			}

			if string(msg) != wired {
				t.Errorf("expected %s, got: %s\n", wired, msg)
			}
		}

		wg.Done()
	}()

	go func() {
		buffer, err := allocateBuffer(len(wired))

		if err != nil {
			t.Error(err)
		}

		for i := 0; i < N; i++ {
			err := msgsnd(id, []byte(wired), buffer, len(wired), 1, 0)

			if err != nil {
				t.Error(err)
				return
			}
		}

		wg.Done()
	}()

	wg.Wait()
}

func Test_SendingBinary(t *testing.T) {
	id := GetTestQueueId(t)

	wired := []byte{0x00, 0x01, 0x02, 0x03, 0x04}

	buffer, err := allocateBuffer(len(wired))

	if err != nil {
		t.Error(err)
	}

	err = msgsnd(id, wired, buffer, len(wired), 1, 0)

	if err != nil {
		t.Error(err)
	}

	msg, _, err := msgrcv(id, 0, buffer, len(wired), 0)

	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(msg, wired) != 0 {
		t.Errorf("expected %q, got: %q", wired, msg)
	}
}

func Test_SendingEmpty(t *testing.T) {
	id := GetTestQueueId(t)

	wired := []byte{}

	buffer, err := allocateBuffer(len(wired))

	if err != nil {
		t.Error(err)
	}

	err = msgsnd(id, wired, buffer, len(wired), 1, 0)

	if err != nil {
		t.Error(err)
	}

	msg, _, err := msgrcv(id, 0, buffer, len(wired), 0)

	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(msg, wired) != 0 {
		t.Errorf("expected %q, got: %q", wired, msg)
	}
}

func Test_SendingUTF8(t *testing.T) {
	id := GetTestQueueId(t)

	wired := "åø…假借字"

	buffer, err := allocateBuffer(len(wired))

	if err != nil {
		t.Error(err)
	}

	err = msgsnd(id, []byte(wired), buffer, len(wired), 1, 0)

	if err != nil {
		t.Error(err)
	}

	msg, _, err := msgrcv(id, 0, buffer, len(wired), 0)

	if err != nil {
		t.Error(err)
	}

	if string(msg) != wired {
		t.Errorf("expected %s, got: %s", wired, msg)
	}
}

func Test_ReceiveBounds(t *testing.T) {
	id := GetTestQueueId(t)

	wired := "123456789"

	buffer, err := allocateBuffer(len(wired))

	if err != nil {
		t.Error(err)
	}

	err = msgsnd(id, []byte(wired), buffer, len(wired), 1, 0)

	if err != nil {
		t.Error(err)
	}

	msg, _, err := msgrcv(id, 0, buffer, len(wired), 0)

	if err != nil {
		t.Error(err)
	}

	if string(msg) != "123456789" {
		t.Errorf("expected %s, got: %s", wired, msg)
	}
}

func Test_GivesE2BIGOnSmallBufferSize(t *testing.T) {
	id := GetTestQueueId(t)

	wired := []byte("123456789")

	buffer, err := allocateBuffer(len(wired) - 1)

	if err != nil {
		t.Error(err)
	}

	err = msgsnd(id, wired, buffer, len(wired), 1, 0)

	if err != nil {
		t.Error(err)
	}

	_, _, err = msgrcv(id, 0, buffer, len(wired)-1, 0)

	if err == nil {
		t.Error(err)
	}

	buffer, err = allocateBuffer(len(wired))

	if err != nil {
		t.Error(err)
	}

	_, _, err = msgrcv(id, 0, buffer, len(wired), 0)

	if err != nil {
		t.Error(err)
	}
}

func Test_RemoveQueue(t *testing.T) {
	id := GetTestQueueId(t)

	wired := []byte("something")

	buffer, err := allocateBuffer(len(wired))

	if err != nil {
		t.Error(err)
	}

	_, err = msgctl(id, IPC_RMID)

	err = msgsnd(id, wired, buffer, len(wired), 1, 0)

	if err == nil {
		t.Error("expected error from msgsnd with removed queue")
	}
}

func Test_ErrorsOnBufferSmallerThanMsg(t *testing.T) {
	id := GetTestQueueId(t)

	wired := []byte("something")

	buffer, err := allocateBuffer(len(wired) - 1)

	if err != nil {
		t.Error(err)
	}

	_, err = msgctl(id, IPC_RMID)

	err = msgsnd(id, wired, buffer, len(wired)-1, 1, 0)

	if err == nil {
		t.Error("expected error because buffer is too small to fit the message")
	}
}
