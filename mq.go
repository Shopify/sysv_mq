package sysv_mq

/*
#include <stdlib.h>
typedef struct _sysv_msg {
  long mtype;
  char mtext[];
} sysv_msg;
*/
import "C"
import "errors"

// Represents the message queue
type MessageQueue struct {
	id     int
	config *QueueConfig
	buffer *C.sysv_msg
}

// Wraps the C structure "struct msgid_ds" (see msgctl(2))
type QueueStats struct {
	Perm   QueuePermissions
	Stime  int
	Rtime  int
	Ctime  int
	Cbytes int
	Qnum   int
	Qbytes int
	Lspid  int
	Lrpid  int
}

// Wraps the C structure "struct ipc_perm" (see msgctl(2))
type QueuePermissions struct {
	Uid  uint
	Gid  uint
	Cuid uint
	Cgid uint
	Mode uint
}

// QueueConfig is used to configure an instance of the message queue.
type QueueConfig struct {
	Mode     int // The mode of the message queue, e.g. 0600
	SendType int // Type to send messages as
	MaxSize  int // Size of the largest message to retrieve or send, allocates a buffer of this size

	Key int // SysV IPC key

	Path   string // The path to a file to obtain a SysV IPC key if Key is not set
	ProjId int    // ProjId for ftok to generate a SysV IPC key if Key is not set
}

// NewMessageQueue returns an instance of the message queue given a QueueConfig.
func NewMessageQueue(config *QueueConfig) (*MessageQueue, error) {
	mq := new(MessageQueue)
	mq.config = config
	err := mq.connect()

	if err != nil {
		return mq, errors.New("Error connecting to queue: " + err.Error())
	}

	mq.buffer, err = allocateBuffer(mq.config.MaxSize)

	if err != nil {
		return mq, errors.New("Error allocating buffer for queue: " + err.Error())
	}

	return mq, err
}

// Sends a string message to the queue of the type passed as the second argument.
func (mq *MessageQueue) Send(message string, msgType int) error {
	return msgsnd(mq.id, message, mq.buffer, mq.config.MaxSize, msgType, 0)
}

// Receive a string message with the type specified by the integer argument.
// Pass 0 to retrieve the message at the top of the queue, regardless of type.
func (mq *MessageQueue) Receive(msgType int) (string, error) {
	mq.buffer.mtype = C.long(msgType)
	return msgrcv(mq.id, msgType, mq.buffer, mq.config.MaxSize, 0)
}

func (mq *MessageQueue) Stat() (*QueueStats, error) {
	info, err := msgctl(mq.id, IPC_STAT)
	if err != nil {
		return nil, err
	}

	perm := QueuePermissions{
		Uid:  uint(info.msg_perm.uid),
		Gid:  uint(info.msg_perm.gid),
		Cuid: uint(info.msg_perm.cuid),
		Cgid: uint(info.msg_perm.cgid),
		Mode: uint(info.msg_perm.mode),
	}

	stat := &QueueStats{
		Perm:   perm,
		Stime:  int(info.msg_stime),
		Rtime:  int(info.msg_rtime),
		Ctime:  int(info.msg_ctime),
		Cbytes: cbytesFromStruct(info),
		Qnum:   int(info.msg_qnum),
		Qbytes: int(info.msg_qbytes),
		Lspid:  int(info.msg_lspid),
		Lrpid:  int(info.msg_lrpid),
	}

	return stat, nil
}

// Number of messages currently in the queue.
func (mq *MessageQueue) Count() (int, error) {
	info, err := mq.Stat()
	return info.Qnum, err
}

// Size of the queue in bytes.
func (mq *MessageQueue) Size() (int, error) {
	info, err := mq.Stat()
	return info.Cbytes, err
}

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
