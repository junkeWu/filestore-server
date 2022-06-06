package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	mdb "github.com/junkeWu/filestore-server/db"
	util "github.com/junkeWu/filestore-server/utils"
)

const PwdSalt = "$crtp1l$"

// SignupHandler  register by password and username
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	if len(username) < 3 || len(password) < 3 {
		w.Write([]byte("invalid parameters"))
		return
	}
	encodedPwd := util.Sha1([]byte(password + PwdSalt))
	up, err := mdb.UserSignUp(username, encodedPwd)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if up {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}
}

type respData struct {
	Data Data `json:"data,omitempty"`
}
type Data struct {
	Token    string `json:"Token,omitempty"`
	UserName string `json:"Username,omitempty"`
	Location string `json:"Location"`
}

// SignInHandler login check
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signin.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		encodedPwd := util.Sha1([]byte(password + PwdSalt))
		// 校验用户名密码
		isSign, err := mdb.UserSignIn(username, encodedPwd)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !isSign {
			w.Write([]byte("FAILED"))
			return
		}
		// 生成访问令牌
		token := GenToken(username)
		updateToken := mdb.UpdateToken(username, token)
		if !updateToken {
			w.Write([]byte("FAILED"))
			return
		}
		// 登录成功重定向到首页
		resp := util.RespMsg{
			Code: 0,
			Msg:  "OK",
			Data: Data{
				Token:    token,
				UserName: username,
				Location: "/static/view/home.html",
			},
		}
		w.Write(resp.JSONBytes())
	}
}

// GenToken 生成token
func GenToken(username string) string {
	// 40位字符 md5(username + timestamp + token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	prefixToken := util.MD5([]byte(username + ts + "_token_salt"))
	return prefixToken + ts[:8]
}

func GetHomeView(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/home.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}
}

func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	// token := r.Form.Get("token")

	// 验证token是否有效
	// isValid := IsTokenValid(token)
	// if !isValid {
	// 	w.WriteHeader(http.StatusForbidden)
	// 	return
	// }
	// 查询用户信息
	info, err := mdb.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: info,
	}
	w.Write(resp.JSONBytes())
}

// IsTokenValid : token是否有效
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token的时效性，是否过期
	// TODO: 从数据库表tbl_user_token查询username对应的token信息
	// TODO: 对比两个token是否一致
	return true
}
