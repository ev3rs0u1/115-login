package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ddliu/go-httpclient"
)

const (
	TokenURL  = "http://qrcodeapi.115.com/api/1.0/web/1.0/token"
	QrcodeURL = "http://qrcode.115.com/api/qrcode.php"
	StatusURL = "http://qrcodeapi.115.com/get/status/"
	LoginURL  = "http://passport.115.com/"
	SpaceURL  = "http://115.com/"
)

var (
	client *httpclient.HttpClient
	token  Token
	status Status
	space  Space
	info   Info
)

type Token struct {
	Code int64 `json:"code"`
	Data struct {
		Qrcode string `json:"qrcode"`
		Sign   string `json:"sign"`
		Time   int64  `json:"time"`
		UID    string `json:"uid"`
	} `json:"data"`
	Message string `json:"message"`
	State   int64  `json:"state"`
}

type Status struct {
	Code int64 `json:"code"`
	Data struct {
		Msg     string `json:"msg"`
		Status  int64  `json:"status"`
		Version string `json:"version"`
	} `json:"data"`
	Message string `json:"message"`
	State   int64  `json:"state"`
}

type Space struct {
	Sign string `json:"sign"`
	Size string `json:"size"`
	Time int64  `json:"time"`
}

type Info struct {
	Data struct {
		UserID   int64  `json:"USER_ID"`
		UserName string `json:"USER_NAME"`
	} `json:"data"`
	Msg   string `json:"msg"`
	State bool   `json:"state"`
}

func checkError(err error) {
	if err != nil {
		fmt.Println("[-]:", err)
		os.Exit(-1)
	}
}

func getTime() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
}

func getToken() {
	res, _ := client.Get(TokenURL, nil)
	bytes, _ := res.ReadAll()
	err := json.Unmarshal(bytes, &token)
	checkError(err)
}

func getQrcode() {
	params := map[string]string{
		"qrform": "1",
		"uid":    token.Data.UID,
		"_t":     getTime(),
	}
	res, _ := client.Get(QrcodeURL, params)
	defer res.Body.Close()
	file, _ := os.Create("qrcode.png")
	defer file.Close()
	_, err := io.Copy(file, res.Body)
	checkError(err)
	fmt.Println("请使用115手机客户端扫码登录")
}

func getCookie() string {
	ary := []string{}
	for k, v := range client.CookieValues(LoginURL) {
		ary = append(ary, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(ary, "; ")
}

func waitLogin() {
loop:
	for {
		params := map[string]string{
			"uid":  token.Data.UID,
			"sign": token.Data.Sign,
			"time": fmt.Sprintf("%d", token.Data.Time),
			"_":    getTime(),
		}
		res, _ := client.Get(StatusURL, params)
		bytes, _ := res.ReadAll()

		err := json.Unmarshal(bytes, &status)
		checkError(err)
		switch status.Data.Status {
		case 1:
			fmt.Println(status.Data.Msg)
		case 2:
			fmt.Println(status.Data.Msg)
			break loop
		}

	}
}

func startLogin() {
	params := map[string]string{
		"ct":  "login",
		"ac":  "qrcode",
		"v":   "android",
		"key": token.Data.UID,
	}
	res, _ := client.Get(LoginURL, params)
	if res.StatusCode == 200 {
		fmt.Println("登录成功!")
	}
}

func getSign() {
	params := map[string]string{
		"ct": "offline",
		"ac": "space",
		"_":  getTime(),
	}
	res, _ := client.Get(SpaceURL, params)
	bytes, _ := res.ReadAll()
	err := json.Unmarshal(bytes, &space)
	checkError(err)
}

func getInfo() {
	params := map[string]string{
		"ct":     "ajax",
		"ac":     "islogin",
		"is_ssl": "1",
	}
	res, _ := client.Get(LoginURL, params)
	bytes, _ := res.ReadAll()
	err := json.Unmarshal(bytes, &info)
	checkError(err)
	fmt.Println("UserID:", info.Data.UserID)
	fmt.Println("UserName:", info.Data.UserName)
}

func updateInfo() {
	params := map[string]string{
		"uid":    fmt.Sprintf("%d", info.Data.UserID),
		"sign":   space.Sign,
		"time":   fmt.Sprintf("%d", space.Time),
		"cookie": getCookie(),
	}
	res, _ := client.Get(UpdateURL, params)
	body, _ := res.ToString()
	fmt.Println(body)
}

func initClient() {
	client = httpclient.NewHttpClient().Defaults(httpclient.Map{
		httpclient.OPT_USERAGENT: "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.86 Safari/537.36",
		"Connection":             "keep-alive",
		"X-Requested-With":       "XMLHttpRequest",
	})
}

func main() {
	initClient()
	getToken()
	getQrcode()
	waitLogin()
	startLogin()
	getSign()
	getInfo()
}
