package db

import (
	"database/sql"
	"fmt"
	"log"

	mdb "github.com/junkeWu/filestore-server/db/mysql"
)

func OnFileUploadFinished(fileHash, filename, fileAddr string, filesize int64) bool {
	conn := mdb.DBConn()
	stmt, err := conn.Prepare(
		"insert ignore into tbl_file(`file_sha1`,`file_name`,`file_addr`,`file_size`,`status`) values (?,?,?,?,?)")
	if err != nil {
		fmt.Println("Failed to prepare statement, err: ", err.Error())
		return false
	}
	defer stmt.Close()
	exec, err := stmt.Exec(fileHash, filename, fileAddr, filesize, 1)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := exec.RowsAffected(); err == nil {
		if rf <= 0 {
			fmt.Printf("file with hash:%s has been uploaded before", fileHash)
		}
		return true
	}
	return false
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// GetFileMeta get file meta by mysql
func GetFileMeta(fileHash string) (*TableFile, error) {
	stmt, err := mdb.DBConn().Prepare(
		"select file_sha1, file_name, file_size, file_addr from tbl_file where file_sha1=? and status=1 limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	file := TableFile{}
	err = stmt.QueryRow(fileHash).Scan(&file.FileHash, &file.FileName, &file.FileSize, &file.FileAddr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &file, nil
}
func GetFileMetaListDB() ([]*TableFile, error) {
	stmt, err := mdb.DBConn().Prepare(
		"select file_sha1, file_name, file_size from tbl_file where status=?")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()
	var files []*TableFile

	rows, err := stmt.Query(1)
	for rows.Next() {
		var file TableFile
		err := rows.Scan(&file.FileHash, &file.FileName, &file.FileSize)
		if err != nil {
			log.Fatalln(err)
			return nil, err
		}
		files = append(files, &file)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return files, nil
}

func RemoveFileMeta(fileHash string) (bool, error) {
	row, err := mdb.DBConn().Exec("delete from tbl_file where file_sha1=?", fileHash)
	if err != nil {
		log.Fatalln(err)
		return false, err
	}
	_, err = row.RowsAffected()
	if err != nil {
		log.Fatalln(err)
		return false, err
	}
	return true, nil
}
