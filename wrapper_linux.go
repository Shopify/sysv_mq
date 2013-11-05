package sysv_mq

func cbytesFromStruct(info *_Ctype_struct_msqid_ds) uint64 {
	return uint64(info.__msg_cbytes)
}
