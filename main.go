package main

import (
	"fmt"
	"net/http"

	"github.com/junkeWu/filestore-server/handler"
)

func main() {
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMeta)
	http.HandleFunc("/file/query", handler.GetFileMeta)
	http.HandleFunc("/file/download", handler.DownloadFile)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server, err: %s", err.Error())
	}
}
