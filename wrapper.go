package sysv_mq

/*
#cgo CFLAGS: -O2
#include <stdlib.h>
#include <sys/types.h>
#include <sys/ipc.h>
#include <sys/msg.h>
#include <fcntl.h>
#include <errno.h>
#include <string.h>
#include <sys/msg.h>

typedef struct _sysv_msg {
  long mtype;
  char mtext[1];
} sysv_msg;
*/
import "C"
import "unsafe"
import "errors"

const (
	IPC_CREAT  = C.IPC_CREAT
	IPC_EXCL   = C.IPC_EXCL
	IPC_NOWAIT = C.IPC_NOWAIT

	IPC_STAT = C.IPC_STAT
	IPC_SET  = C.IPC_SET
	IPC_RMID = C.IPC_RMID

	MemoryAllocationError   = "malloc failed to allocate memory"
	MessageBiggerThanBuffer = "message length is longer than the size of the buffer"
)

// msgop(2)
// int msgsnd(int msqid, const void *msgp, size_t msgsz, int msgflg)
func msgsnd(key int, message []byte, buffer *C.sysv_msg, maxSize int, mtype int, flags int) error {
	if len(message) > maxSize {
		return errors.New(MessageBiggerThanBuffer)
	}

	msgSize := C.size_t(len(message))

	buffer.mtype = C.long(mtype)
	if msgSize > 0 {
		C.memcpy(unsafe.Pointer(&buffer.mtext), unsafe.Pointer(&message[0]), msgSize)
	}

	_, err := C.msgsnd(C.int(key), unsafe.Pointer(buffer), msgSize, C.int(flags))

	return err
}

// msgop(2)
// ssize_t msgrcv(int msqid, void *msgp, size_t msgsz, long msgtyp, int msgflg);
func msgrcv(key int, mtype int, buffer *C.sysv_msg, strSize int, flags int) ([]byte, int, error) {
	length, err := C.msgrcv(C.int(key), unsafe.Pointer(buffer), C.size_t(strSize), C.long(mtype), C.int(flags))

	if err != nil {
		return nil, 0, err
	}

	return C.GoBytes(unsafe.Pointer(&buffer.mtext), C.int(length)), int(buffer.mtype), nil
}

// msgget(2)
// int msgget(key_t key, int msgflg);
func msgget(key int, mode int) (int, error) {
	res, err := C.msgget(C.key_t(key), C.int(mode))

	if err != nil {
		return -1, err
	}

	return int(res), nil
}

// ftok(3):
// key_t ftok(const char *pathname, int proj_id);
func ftok(path string, projId int) (int, error) {
	cs := C.CString(path)

	if cs == nil {
		return 0, errors.New(MemoryAllocationError)
	}

	defer C.free(unsafe.Pointer(cs))

	res, err := C.ftok(cs, C.int(projId))

	if err != nil {
		return -1, err
	}

	return int(res), nil
}

// msgctl(2)
// int msgctl(int msqid, int cmd, struct msqid_ds *buf);
func msgctl(key int, cmd int) (*C.struct_msqid_ds, error) {
	info := new(C.struct_msqid_ds)
	_, err := C.msgctl(C.int(key), C.int(cmd), info)

	return info, err
}

// The buffer is malloced, and not handled by Go, because SysV MQs do
// runtime-length inline arrays that Go does not support without a bunch of reflection
func allocateBuffer(strSize int) (*C.sysv_msg, error) {
	// you can't reliably take the size of C structs from go (yay platform-dependent padding/alignment
	// differences) so we manually construct what should basically just be sizeof(C.sysv_msg) + strSize.
	// Fortunately there's only one other member besides the variable-length buffer, so it's not too bad.
	bufferSize := C.size_t(strSize) + C.size_t(unsafe.Sizeof(C.long(1)))
	buffer := (*C.sysv_msg)(C.malloc(bufferSize))

	if buffer == nil {
		return buffer, errors.New(MemoryAllocationError)
	}

	return buffer, nil
}

func freeBuffer(buffer *C.sysv_msg) {
	C.free(unsafe.Pointer(buffer))
}

// Wraps msgctl(key, IPC_RMID).
func ipcDestroy(key int) error {
	_, err := msgctl(key, IPC_RMID)
	return err
}

// Wraps msgctl(key, IPC_STAT).
func ipcStat(key int) (*QueueStats, error) {
	info, err := msgctl(key, IPC_STAT)
	if err != nil {
		return nil, err
	}

	perm := QueuePermissions{
		Uid:  uint32(info.msg_perm.uid),
		Gid:  uint32(info.msg_perm.gid),
		Cuid: uint32(info.msg_perm.cuid),
		Cgid: uint32(info.msg_perm.cgid),
		Mode: uint16(info.msg_perm.mode),
	}

	stat := &QueueStats{
		Perm:  perm,
		Stime: int64(info.msg_stime),
		// Rtime:  int64(info.msg_rtime), // https://github.com/Shopify/sysv_mq/issues/10
		Ctime:  int64(info.msg_ctime),
		Cbytes: cbytesFromStruct(info),
		Qnum:   uint64(info.msg_qnum),
		Qbytes: uint64(info.msg_qbytes),
		Lspid:  int32(info.msg_lspid),
		Lrpid:  int32(info.msg_lrpid),
	}

	return stat, nil
}
