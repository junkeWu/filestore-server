package main

import (
	"fmt"
	"net/http"

	"github.com/junkeWu/filestore-server/handler"
)

func main() {
	fmt.Println("server start")
	// http.HandleFunc("/static/view/home.html", handler.GetHomeView)

	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMeta)
	http.HandleFunc("/file/query", handler.GetFileMetaList)
	http.HandleFunc("/file/download", handler.DownloadFile)
	http.HandleFunc("/file/update", handler.UpdateFileMeta)
	http.HandleFunc("/file/delete", handler.DeleteFileMeta)

	// http.HandleFunc("/home", handler.GetHomeView)
	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server, err: %s", err.Error())
	}
}
