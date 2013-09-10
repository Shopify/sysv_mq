package sysv_mq

// Size of the queue in bytes.
func (mq *MessageQueue) Size() (int, error) {
	info, err := msgctl(mq.id, IPC_STAT)
	return int(info.__msg_cbytes), err
}
