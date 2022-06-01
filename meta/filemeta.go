package meta

import "fmt"

// FileMeta file meta struct
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

func init() {
	if fileMetas != nil {
		return
	}
	fileMetas = make(map[string]FileMeta)
	fmt.Println("fileMetas init success")
}

// UploadFileMeta insert/upload file meta
func UploadFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// GetFileMeta get file meta object by sha1
func GetFileMeta(fsha1 string) FileMeta {
	return fileMetas[fsha1]
}
