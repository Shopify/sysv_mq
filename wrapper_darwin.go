package sysv_mq

func cbytesFromStruct(info *_Ctype_struct___msqid_ds_new) uint64 {
	return uint64(info.msg_cbytes)
}
