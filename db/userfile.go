package db

import (
	"log"
	"time"

	mdb "github.com/junkeWu/filestore-server/db/mysql"
)

type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// OnUserFileUploadFinished insert data into tbl_file_user
func OnUserFileUploadFinished(username, fileHash, fileName string, fileSize int64) (bool, error) {
	stmt, err := mdb.DBConn().Prepare(
		"insert ignore into tbl_user_file(`user_name`,`file_sha1`,`file_name`,`file_size`, `upload_at`) values (?,?,?,?,?)",
	)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(username, fileHash, fileName, fileSize, time.Now())
	if err != nil {
		log.Println(err)
		return false, err
	}
	return true, nil
}

// QueryUserFileMetas get tbl_user_file data.
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	stmt, err := mdb.DBConn().Prepare(
		"select file_sha1, file_name, file_size, upload_at, last_update from tbl_user_file where user_name=? limit ?")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rows, err := stmt.Query(username, limit)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var userFiles []UserFile
	for rows.Next() {
		var userFile UserFile
		err := rows.Scan(&userFile.FileHash, &userFile.FileName, &userFile.FileSize, &userFile.UploadAt, &userFile.LastUpdated)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		userFiles = append(userFiles, userFile)
	}
	return userFiles, nil
}
