package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/junkeWu/filestore-server/meta"
	util "github.com/junkeWu/filestore-server/utils"
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

		// store by fileMeta
		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "./temp/" + head.Filename,
			UploadAt: time.Now().Format("2006.01.02 15:04:15"),
		}

		// 在temp文件夹下新建文件
		newFile, err := os.Create("./temp/" + head.Filename)
		if err != nil {
			fmt.Printf("Failed to create file: %s\n", err.Error())
			return
		}
		wt := bufio.NewWriter(newFile)
		defer newFile.Close()
		// 调用io buffer写入磁盘
		fileMeta.FileSize, err = io.Copy(wt, file)
		if err != nil {
			fmt.Printf("Failer to save data into file, err: %s\n", err.Error())
			return
		}
		wt.Flush()
		// compute file sha1
		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		meta.UploadFileMeta(fileMeta)
		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

// UploadSucHandler 文件上传成功handler
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload Finished")
}

// GetFileMeta get file meta
func GetFileMeta(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form["filehash"][0]
	fmeta := meta.GetFileMeta(filehash)
	data, err := json.Marshal(fmeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func GetFileMetaList(w http.ResponseWriter, r *http.Request) {
	list := meta.GetFileMetaList()
	data, err := json.Marshal(&list)
	util.StatusInternalServer(w, err)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// find file meta
	filehash := r.Form.Get("filehash")
	fm := meta.GetFileMeta(filehash)
	// download file by path
	f, err := os.Open(fm.Location)
	util.StatusInternalServer(w, err)

	defer f.Close()
	// todo use buffer read file
	data, err := ioutil.ReadAll(f)
	// w.WriteHeader(http.StatusInternalServerError)
	// log.Println(err)
	util.StatusInternalServer(w, err)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\""+fm.FileName+"\"")
	w.Write(data)
}

// UpdateFileMeta update file meta
func UpdateFileMeta(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	operator := r.Form.Get("op")
	filename := r.Form.Get("filename")
	fileHash := r.Form.Get("filehash")

	if operator != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	fileMeta := meta.GetFileMeta(fileHash)
	fileMeta.FileName = filename
	meta.UploadFileMeta(fileMeta)

	data, err := json.Marshal(fileMeta)
	util.StatusInternalServer(w, err)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// DeleteFileMeta delete file meta
// todo 保持线程安全
func DeleteFileMeta(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form.Get("filehash")
	fileMeta := meta.GetFileMeta(fileHash)
	os.Remove(fileMeta.Location)
	meta.RemoveFileMeta(fileHash)
	w.WriteHeader(http.StatusOK)
	http.Redirect(w, r, "/file/getList", http.StatusFound)
}
