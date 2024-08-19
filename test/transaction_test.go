package test

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/llyb120/vermouth"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTransaction(t *testing.T) {

	r := gin.Default()
	vermouth.RegisterControllers(r, NewTestController())

	var err error
	// 初始化数据库连接
	dsn := "root:root@tcp(127.0.0.1:3306)/test"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	vermouth.SetDB(db)

	req, _ := http.NewRequest("GET", "/api/test4", nil)

	// 创建一个响应记录器
	w := httptest.NewRecorder()

	// 处理请求
	r.ServeHTTP(w, req)

	// vermouth.GetDB().Exec("INSERT INTO user (name, age) VALUES (?, ?)", "John", 20)
	// c.JSON(200, gin.H{"message": "ok"})
}
