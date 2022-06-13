package handler

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	redisPool "github.com/junkeWu/filestore-server/cache/redis"
	dblayer "github.com/junkeWu/filestore-server/db"
	util "github.com/junkeWu/filestore-server/utils"
)

const (
	ChunkCount = "chunkcount"
)

// MultipartUploadInfo 分块的结构体
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int64
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// InitialMultipartUploadHandler 初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 获得redis的一个连接
	pool := redisPool.NewRedisPool().Get()
	defer pool.Close()
	// 生成分块上传的初始化信息
	info := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   int64(filesize),
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}
	// 将初始化信息写入redis缓存
	pool.Do("HSET", "MP_"+info.UploadID, ChunkCount, info.ChunkCount)
	pool.Do("HSET", "MP_"+info.UploadID, ChunkCount, info.FileHash)
	pool.Do("HSET", "MP_"+info.UploadID, ChunkCount, info.FileSize)
	// 将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "OK", info).JSONBytes())
}

// UploadPartHandler : 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	//	username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	// 2. 获得redis连接池中的一个连接
	rConn := redisPool.NewRedisPool().Get()
	defer rConn.Close()

	// 3. 获得文件句柄，用于存储分块内容
	fpath := "/data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 4. 更新redis缓存状态
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// CompleteUploadHandler 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 解析参数
	r.ParseForm()
	upId := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	fileHash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")
	// 获得连接池的一个连接
	pool := redisPool.NewRedisPool().Get()
	defer pool.Close()
	// 通过UploadID查询redis并判断所有分块都上传完成
	data, err := redis.Values(pool.Do("HGETALL", "MP_"+upId))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed.", nil).JSONBytes())
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}
	// todo 分块合并
	// 更新文件表和用户文件表
	fsize, _ := strconv.Atoi(filesize)
	dblayer.OnFileUploadFinished(fileHash, filename, "", int64(fsize))
	dblayer.OnUserFileUploadFinished(username, fileHash, filename, int64(fsize))

	// 6. 响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
	// 响应
}
