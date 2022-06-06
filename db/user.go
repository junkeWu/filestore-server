package db

import (
	"log"

	mdb "github.com/junkeWu/filestore-server/db/mysql"
)

// UserSignUp register by username and password
func UserSignUp(username, password string) (bool, error) {
	stmt, err := mdb.DBConn().Prepare(
		"insert ignore into tbl_user(`user_name`, `user_pwd`) values (?,?)",
	)
	if err != nil {
		log.Fatalln(err)
		return false, err
	}
	defer stmt.Close()
	row, err := stmt.Exec(username, password)
	if err != nil {
		log.Fatalln(err)
		return false, err
	}
	if af, err := row.RowsAffected(); err == nil && af > 0 {
		return true, err
	}
	return false, err
}

// UserSignIn login by username and password
func UserSignIn(username, password string) (bool, error) {
	stmt, err := mdb.DBConn().Prepare(
		"select user_name, user_pwd from tbl_user where user_name=? limit 1",
	)
	if err != nil {
		log.Fatalln(err)
		return false, err
	}
	defer stmt.Close()

	row, err := stmt.Query(username)
	if err != nil {
		log.Fatalln(err)
		return false, err
	} else if row == nil {
		log.Fatalln("username not found: " + username)
		return false, err
	}
	rows := mdb.ParseRows(row)
	if len(rows) > 0 && string(rows[0]["user_pwd"].([]byte)) == password {
		return true, nil
	}
	return false, err
}

// UpdateToken update user login's token
func UpdateToken(username, token string) bool {
	stmt, err := mdb.DBConn().Prepare(
		"replace into tbl_user_token(`user_name`,`user_token`) values (?,?)",
	)
	if err != nil {
		log.Fatalln(err)
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		log.Fatalln(err)
		return false
	}
	return true
}

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// GetUserInfo get user	by username
func GetUserInfo(username string) (User, error) {
	user := User{}
	stmt, err := mdb.DBConn().Prepare(
		"select user_name, signup_at from tbl_user where user_name=? limit 1",
	)
	if err != nil {
		log.Println(err)
		return user, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		log.Println(err)
		return user, err
	}
	return user, nil
}
