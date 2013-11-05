package sysv_mq

func cbytesFromStruct(info *_Ctype_struct_msqid_ds) int {
	return int(info.__msg_cbytes)
}
