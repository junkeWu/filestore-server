package main

import (
	"fmt"
	"net/http"

	"github.com/junkeWu/filestore-server/handler"
)

func main() {
	fmt.Println("server start")
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMeta)
	http.HandleFunc("/file/getList", handler.GetFileMetaList)
	http.HandleFunc("/file/download", handler.DownloadFile)
	http.HandleFunc("/file/update", handler.UpdateFileMeta)
	http.HandleFunc("/file/delete", handler.DeleteFileMeta)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server, err: %s", err.Error())
	}
}
