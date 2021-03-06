package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/junkeWu/filestore-server/common"
	cfg "github.com/junkeWu/filestore-server/config"
	mdb "github.com/junkeWu/filestore-server/db"
	"github.com/junkeWu/filestore-server/meta"
	"github.com/junkeWu/filestore-server/mq"
	"github.com/junkeWu/filestore-server/store/oss"
	util "github.com/junkeWu/filestore-server/utils"
)

// UploadHandler 文件上传接口
func UploadHandler(w http.ResponseWriter, r *http.Request) {
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
		// 实际写盘操作
		wt.Flush()
		// compute file sha1
		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		newFile.Seek(0, 0)
		// oss
		ossPath := "oss/" + fileMeta.Location
		if !cfg.AsyncTransferEnable {
			err = oss.Bucket().PutObject(ossPath, newFile)
			util.Must(err)
			// 已经在oss上
			fileMeta.Location = ossPath
		} else {
			data := mq.TransferData{
				FileHash:     fileMeta.FileSha1,
				CurLocation:  fileMeta.Location,
				DestLocation: ossPath,
				DesStoreType: common.StoreOSS,
			}
			fmt.Println("已经写入mq队列：", data)
			pubData, _ := json.Marshal(data)
			pubSuc := mq.Publish(
				cfg.TransExchangeName,
				cfg.TransOSSRoutingKey,
				pubData,
			)
			if !pubSuc {
				// TODO: 当前发送转移信息失败，稍后重试
				fmt.Println("发布失败")
			}
		}
		util.Must(err)

		// meta.UploadFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)
		// todo update into user_file table
		r.ParseForm()
		username := r.Form.Get("username")
		finished, err := mdb.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if finished {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed."))
		}
	}
}

// UploadSucHandler 文件上传成功handler
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload Finished")
}

// GetFileMeta get file meta
func GetFileMeta(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form["fileHash"][0]
	// fmeta := meta.GetFileMeta(fileHash)
	fmeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	data, err := json.Marshal(fmeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func GetFileMetaList(w http.ResponseWriter, r *http.Request) {
	// list := meta.GetFileMetaList()
	// list, err := meta.GetFileMetaListDB()
	r.ParseForm()
	limit, err := strconv.Atoi(r.Form.Get("limit"))
	if err != nil {
		log.Println(err)
		return
	}
	list, err := mdb.QueryUserFileMetas(r.Form.Get("username"), limit)
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
	fileHash := r.Form.Get("filehash")
	// fm := meta.GetFileMeta(fileHash)
	fm, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		log.Fatalln(err)
		return
	}
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
	fileHash := r.Form.Get("fileHash")

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
	fileHash := r.Form.Get("fileHash")
	fileMeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		log.Fatalln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var mutex sync.Mutex
	mutex.Lock()
	_, err = meta.RemoveFileMetaDB(fileHash)
	mutex.Unlock()
	if err != nil {
		log.Fatalln(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	os.Remove(fileMeta.Location)
	meta.RemoveFileMeta(fileHash)
	w.WriteHeader(http.StatusOK)
	http.Redirect(w, r, "/file/getList", http.StatusFound)
}

// TryFastUploadHandler fast upload
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 解析请求参数
	username := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 从文件表中查询相同hash的记录
	fileMeta, err := meta.GetFileMetaDB(fileHash)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
			Data: nil,
		}
		w.Write(resp.JSONBytes())
	}
	// 上传过则将文件信息写入用户文件表，返回成功
	finished, err := mdb.OnUserFileUploadFinished(username, fileHash, filename, int64(filesize))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if finished {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功。",
		}
		w.Write(resp.JSONBytes())
	} else {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请稍后重试。",
		}
		w.Write(resp.JSONBytes())
	}
}

func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileHash := r.Form.Get("filehash")

	// 从文件表查找记录
	row, err := mdb.GetFileMeta(fileHash)
	util.Must(err)
	url := oss.DownloadURL(row.FileAddr.String)
	w.Write([]byte(url))
}
