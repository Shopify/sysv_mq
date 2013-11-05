package sysv_mq

func cbytesFromStruct(info *_Ctype_struct___msqid_ds_new) int {
	return int(info.msg_cbytes)
}
