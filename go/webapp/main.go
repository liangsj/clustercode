package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

//c/s协议返回的json的结构体
// {
//    errno : -1/0, 错误码，0代表正常
//    errmsg: "" , 错误信息
//    data : {
//             resourceID : 111, 资源id，用于表示唯一的一条资源
//             count      : 1 ,  此资源被点赞的次数
//             Ctime      : 1231231, 此项被创建的时间
//  }
//

type Response struct {
	Errno  int    `json:"errno"`
	ErrMsg string `json:"errmsg"`
	Data   *Item  `json:"data"`
}
type Item struct {
	ResourceID int64 `json:"resourceID"`
	Count      int64 `json:"count"`
}

var dbAddr string
var port int

func init() {
	flag.StringVar(&dbAddr, "mysql", "mysql", "please input your mysql address,exp : 127.0.0.1")
	flag.IntVar(&port, "port", 3306, "please input your mysql port,exp : 3304")
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, world!\n")
}

func main() {
	flag.Parse()
	log.Println("server init")
	http.HandleFunc("/hello", HelloServer)
	http.HandleFunc("/praise/get", getPrasieCount)
	http.HandleFunc("/praise/set", setPrasieCount)
	log.Println("server start")
	log.Fatal(http.ListenAndServe(":80", nil))
}

func getPrasieCount(w http.ResponseWriter, req *http.Request) {
	resourceID, err := getResourceIDFromGet(req)
	if err != nil {
		returnErrMsg(w, -1, fmt.Sprintf("%v", err))
		return
	}

	dbOpenCommand := fmt.Sprintf("test:test@tcp(%s:%d)/praise?charset=utf8", dbAddr, port)
	db, err := sql.Open("mysql", dbOpenCommand)

	if err != nil {
		returnErrMsg(w, -1, fmt.Sprintf("%v", err))
		return
	}
	defer db.Close()

	sql := fmt.Sprintf("select count,ctime from praise_count where resource_id = %d limit 1", resourceID)
	rows, sqlerr := db.Query(sql)
	if sqlerr != nil {
		returnErrMsg(w, -1, fmt.Sprintf("%v", sqlerr))
		return
	}

	defer rows.Close()
	res := Response{Errno: 0}
	var count, ctime int64
	if rows.Next() {
		rows.Scan(&count, &ctime)
	}

	res.Data = &Item{ResourceID: resourceID, Count: count}

	retBytes, _ := json.Marshal(res)

	io.WriteString(w, string(retBytes))

}

func setPrasieCount(w http.ResponseWriter, req *http.Request) {
	resourceID, err := getResourceIDFromGet(req)
	if err != nil {
		returnErrMsg(w, -1, fmt.Sprintf("%v", err))
		return
	}

	dbOpenCommand := fmt.Sprintf("test:test@tcp(%s:%d)/praise?charset=utf8", dbAddr, port)
	db, err := sql.Open("mysql", dbOpenCommand)
	if err != nil {
		returnErrMsg(w, -1, fmt.Sprintf("%v", err))
		return
	}
	defer db.Close()
	sql := fmt.Sprintf("INSERT INTO praise_count  (resource_id,count ) VALUES (%d,1) ON DUPLICATE key UPDATE count=count+1", resourceID)
	_, err = db.Query(sql)
	if err != nil {
		returnErrMsg(w, -1, fmt.Sprintf("%v", err))
		return
	}
	retBytes, _ := json.Marshal(Response{Errno: 0, Data: &Item{}})
	io.WriteString(w, string(retBytes))

}

func returnErrMsg(w http.ResponseWriter, errno int, errmsg string) {

	retBytes, _ := json.Marshal(Response{Errno: errno, ErrMsg: errmsg})
	io.WriteString(w, string(retBytes))
}

func getResourceIDFromGet(req *http.Request) (int64, error) {
	vars := req.URL.Query()
	resourceIDStr := vars.Get("resource_id")
	resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	return resourceID, err
}
