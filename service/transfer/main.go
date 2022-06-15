package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	"github.com/junkeWu/filestore-server/config"
	dblayer "github.com/junkeWu/filestore-server/db"
	"github.com/junkeWu/filestore-server/mq"
	"github.com/junkeWu/filestore-server/store/oss"
)

// ProcessTransfer : 处理文件转移
func ProcessTransfer(msg []byte) bool {
	log.Println(string(msg))
	// 解析msg
	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	// 根据临时存储文件路径，创建文件句柄
	fin, err := os.Open(pubData.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	// 通过文件句柄讲文件内容读出来并且上传到oss
	err = oss.Bucket().PutObject(
		pubData.DestLocation,
		bufio.NewReader(fin))
	if err != nil {
		log.Println(err.Error())
		return false
	}
	// 更新文件的存储路径到文件表
	_ = dblayer.UpdateFileLocation(
		pubData.FileHash,
		pubData.DestLocation,
	)
	return true
}

func main() {
	log.Println("文件转移服务启动中，开始监听转移任务队列...")
	mq.StartConsumer(
		config.TransOSSQueueName,
		"transfer_oss",
		ProcessTransfer)
}
