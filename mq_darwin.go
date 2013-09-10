package sysv_mq

// Size of the queue in bytes.
func (mq *MessageQueue) Size() (int, error) {
	info, err := msgctl(mq.id, IPC_STAT)
	return int(info.msg_cbytes), err
}
