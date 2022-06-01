package handler

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// UploadHandler 文件上传接口
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("come in")
	if r.Method == "GET" {
		// 返回上传html页面
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		// 接收客户端文件保存本地
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("upload fild failed: %s\n", err.Error())
			return
		}
		defer file.Close()
		// 在temp文件夹下新建文件
		newFile, err := os.Create("./temp/" + head.Filename)
		if err != nil {
			fmt.Printf("Failed to create file: %s\n", err.Error())
			return
		}
		wt := bufio.NewWriter(newFile)
		defer newFile.Close()
		// 调用io buffer写入磁盘
		_, err = io.Copy(wt, file)
		if err != nil {
			fmt.Printf("Failer to save data into file, err: %s\n", err.Error())
			return
		}
		wt.Flush()
		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

// UploadSucHandler 文件上传成功handler
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload Finished")
}
