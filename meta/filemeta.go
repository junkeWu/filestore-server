package meta

import (
	"fmt"
	"log"

	mdb "github.com/junkeWu/filestore-server/db"
)

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

// UpdateFileMetaDB insert into data to mysql server
func UpdateFileMetaDB(fmeta FileMeta) bool {
	return mdb.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName, fmeta.Location, fmeta.FileSize)
}

// GetFileMeta get file meta object by sha1
func GetFileMeta(fsha1 string) FileMeta {
	return fileMetas[fsha1]
}

// GetFileMetaDB get file meta by db
func GetFileMetaDB(fileHash string) (*FileMeta, error) {
	meta, err := mdb.GetFileMeta(fileHash)
	if err != nil {
		return nil, err
	}
	return &FileMeta{
		FileSha1: meta.FileHash,
		FileName: meta.FileName.String,
		FileSize: meta.FileSize.Int64,
		Location: meta.FileAddr.String,
	}, nil
}

func GetFileMetaList() (fms []FileMeta) {
	for _, meta := range fileMetas {
		fms = append(fms, meta)
	}
	return fms
}

func GetFileMetaListDB() ([]*FileMeta, error) {
	files, err := mdb.GetFileMetaListDB()
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	var fileMetaList []*FileMeta
	for _, file := range files {
		var fileMeta FileMeta
		fileMeta.FileSha1 = file.FileHash
		fileMeta.FileName = file.FileName.String
		fileMeta.FileSize = file.FileSize.Int64
		fileMeta.Location = file.FileAddr.String
		fileMetaList = append(fileMetaList, &fileMeta)
	}
	return fileMetaList, nil
}

func RemoveFileMeta(fileHash string) {
	delete(fileMetas, fileHash)
}
func RemoveFileMetaDB(fileHash string) (bool, error) {
	meta, err := mdb.RemoveFileMeta(fileHash)
	if err != nil {
		log.Fatalln(err)
		return false, err
	}
	return meta, nil
}
