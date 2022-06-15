package mq

import "github.com/junkeWu/filestore-server/common"

// TransferData 转移消息结构体
type TransferData struct {
	FileHash     string
	CurLocation  string
	DestLocation string
	DesStoreType common.StoreType
}
