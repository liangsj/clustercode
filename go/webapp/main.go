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
	"github.com/gomodule/redigo/redis"
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

const (
	praise_count_cache_key_fmt string = "resource_%d_prasie_cache"
)

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

	praiseCount, err := getFromCache(resourceID)
	if err != nil && err != redis.ErrNil {
		returnErrMsg(w, -1, fmt.Sprintf("%v", err))
		return
	}
	if err != redis.ErrNil {

	}
	dbOpenCommand := fmt.Sprintf("test:test@tcp(%s:%d)/praise?charset=utf8", dbAddr, port)
	db, err := sql.Open("mysql", dbOpenCommand)

	if err != nil {
		returnErrMsg(w, -1, fmt.Sprintf("%v", err))
		return
	}
	defer db.Close()

	sql := fmt.Sprintf("select count from praise_count where resource_id = %d limit 0,1", resourceID)
	rows, sqlerr := db.Query(sql)
	if sqlerr != nil {
		returnErrMsg(w, -1, fmt.Sprintf("%v", sqlerr))
		return
	}

	defer rows.Close()
	res := Response{Errno: 0}
	var count int64
	if rows.Next() {
		rows.Scan(&count)
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
	retBytes, _ := json.Marshal(Response{Errno: 0})
	io.WriteString(w, string(retBytes))

}

func returnErrMsg(w http.ResponseWriter, errno int, errmsg string) {

	retBytes, _ := json.Marshal(Response{Errno: errno, ErrMsg: errmsg})
	io.WriteString(w, string(retBytes))
}

//将点赞数放入redis缓存中
func setToCache(resourceID int64, praiseCount int64) error {
	conn, err := redis.Dial("tcp", "redis:3304")
	if err != nil {
		return err
	}
	defer conn.Close()
	key := fmt.Sprintf(praise_count_cache_key_fmt, resourceID)
	_, err = conn.Do("SET", key, praiseCount)
	return err

}

//从redis中获取点赞数
func getFromCache(resourceID int64) (int64, error) {
	conn, err := redis.Dial("tcp", "redis:3304")
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	key := fmt.Sprintf(praise_count_cache_key_fmt, resourceID)
	praiseCount, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		return 0, err
	}
	return praiseCount, nil
}

func getResourceIDFromGet(req *http.Request) (int64, error) {
	vars := req.URL.Query()
	resourceIDStr := vars.Get("resource_id")
	resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	return resourceID, err
}
