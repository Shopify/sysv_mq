package sysv_mq

/*
#include <stdlib.h>
typedef struct _sysv_msg {
  long mtype;
  char mtext[1];
} sysv_msg;
*/
import "C"
import (
	"errors"
	"runtime"
)

// Represents the message queue
type MessageQueue struct {
	id     int
	config *QueueConfig
	buffer *C.sysv_msg
	stats  *QueueStats
}

// Wraps the C structure "struct msgid_ds" (see msgctl(2))
type QueueStats struct {
	Perm  QueuePermissions
	Stime int64 // signed long, according to bits/types.h
	// Rtime  int64  // https://github.com/Shopify/sysv_mq/issues/10
	Ctime  int64  //
	Cbytes uint64 // unsigned long, according to msgctl(2)
	Qnum   uint64 // unsigned long, according to bits/msq.h
	Qbytes uint64 // unsigned long, according to bits/msg.h
	Lspid  int32  // signed int32, according to bits/typesizes.h
	Lrpid  int32  //
}

// Wraps the C structure "struct ipc_perm" (see msgctl(2))
type QueuePermissions struct {
	Uid  uint32 // unsigned int32, according to bits/typesizes.h
	Gid  uint32 //
	Cuid uint32 //
	Cgid uint32 //
	Mode uint16 // unsigned short, according to msgctl(2)
}

// QueueConfig is used to configure an instance of the message queue.
type QueueConfig struct {
	Mode    int // The mode of the message queue, e.g. 0600
	MaxSize int // Size of the largest message to retrieve or send, allocates a buffer of this size

	Key int // SysV IPC key

	Path   string // The path to a file to obtain a SysV IPC key if Key is not set
	ProjId int    // ProjId for ftok to generate a SysV IPC key if Key is not set
}

// QueueSet is used to modify an instance of the message queue.
type QueueSet struct {
	Uid    uint32 // unsigned int32, according to bits/typesizes.h
	Gid    uint32 //
	Mode   uint16 // unsigned short, according to msgctl(2)
	Qbytes uint64 // unsigned long, according to bits/msg.h
}

// NewMessageQueue returns an instance of the message queue given a QueueConfig.
func NewMessageQueue(config *QueueConfig) (*MessageQueue, error) {
	mq := new(MessageQueue)
	mq.id = -1
	mq.config = config
	err := mq.connect()

	if err != nil {
		return mq, err
	}

	mq.buffer, err = allocateBuffer(mq.config.MaxSize)

	if err != nil {
		return mq, err
	}

	mq.stats, err = mq.Stat()

	if err != nil {
		return mq, err
	}

	runtime.SetFinalizer(mq, func(mq *MessageQueue) {
		mq.Close()
	})

	return mq, err
}

// Sends a []byte message to the queue of the type passed as the second argument.
func (mq *MessageQueue) SendBytes(message []byte, msgType int, flags int) error {
	return msgsnd(mq.id, message, mq.buffer, mq.config.MaxSize, msgType, flags)
}

// Sends a string message to the queue of the type passed as the second argument.
func (mq *MessageQueue) SendString(message string, msgType int, flags int) error {
	return mq.SendBytes([]byte(message), msgType, flags)
}

// Receive a []byte message with the type specified by the integer argument.
// Pass 0 to retrieve the message at the top of the queue, regardless of type.
func (mq *MessageQueue) ReceiveBytes(msgType int, flags int) ([]byte, int, error) {
	mq.buffer.mtype = C.long(msgType)
	return msgrcv(mq.id, msgType, mq.buffer, mq.config.MaxSize, flags)
}

// Receive a string message with the type specified by the integer argument.
// Pass 0 to retrieve the message at the top of the queue, regardless of type.
func (mq *MessageQueue) ReceiveString(msgType int, flags int) (string, int, error) {
	mq.buffer.mtype = C.long(msgType)
	bytes, mtype, err := mq.ReceiveBytes(msgType, flags)
	if err != nil {
		return "", 0, err
	} else {
		return string(bytes), mtype, nil
	}
}

// Get statistics about the message queue.
func (mq *MessageQueue) Stat() (*QueueStats, error) {
	return ipcStat(mq.id)
}

// Get statistics about the message queue.
func (mq *MessageQueue) Destroy() error {
	defer mq.Close()
	return ipcDestroy(mq.id)
}

// Number of messages currently in the queue.
func (mq *MessageQueue) Count() (uint64, error) {
	info, err := mq.Stat()
	return info.Qnum, err
}

// Size of the queue in bytes.
func (mq *MessageQueue) Size() (uint64, error) {
	info, err := mq.Stat()
	return info.Cbytes, err
}

// Frees up the resources associated with the message queue,
// but it will leave the message wueue itself in place.
func (mq *MessageQueue) Close() {
	if mq.id > -1 {
		freeBuffer(mq.buffer)
		mq.id = -1
		mq.config = nil
	}
}

// connect the the message queue
func (mq *MessageQueue) connect() (err error) {
	if mq.config.Key == 0 {
		mq.config.Key, err = ftok(mq.config.Path, mq.config.ProjId)

		if err != nil {
			return errors.New("Error obtaining key with ftok: " + err.Error())
		}
	}

	mq.id, err = msgget(mq.config.Key, mq.config.Mode)
	return err
}

// Set modify the message queue
func (mq *MessageQueue) Set(queueSet *QueueSet) error {
	return ipcSet(mq.id, queueSet)
}

// SetQbytes modify the qbytes of the message queue
func (mq *MessageQueue) SetQbytes(qbytes uint64) error {
	queueSet := &QueueSet{
		Uid:    mq.stats.Perm.Uid,
		Gid:    mq.stats.Perm.Gid,
		Mode:   mq.stats.Perm.Mode,
		Qbytes: qbytes,
	}
	return mq.Set(queueSet)
}

// SetUid modify the uid of the message queue
func (mq *MessageQueue) SetUid(uid uint32) error {
	queueSet := &QueueSet{
		Uid:    uid,
		Gid:    mq.stats.Perm.Gid,
		Mode:   mq.stats.Perm.Mode,
		Qbytes: mq.stats.Qbytes,
	}
	return mq.Set(queueSet)
}

// SetGid modify the gid of the message queue
func (mq *MessageQueue) SetGid(gid uint32) error {
	queueSet := &QueueSet{
		Uid:    mq.stats.Perm.Uid,
		Gid:    gid,
		Mode:   mq.stats.Perm.Mode,
		Qbytes: mq.stats.Qbytes,
	}
	return mq.Set(queueSet)
}

// SetMode modify the mode of the message queue
func (mq *MessageQueue) SetMode(mode uint16) error {
	queueSet := &QueueSet{
		Uid:    mq.stats.Perm.Uid,
		Gid:    mq.stats.Perm.Gid,
		Mode:   mode,
		Qbytes: mq.stats.Qbytes,
	}
	return mq.Set(queueSet)
}
