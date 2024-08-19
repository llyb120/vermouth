package test

import (
	"database/sql"
	"fmt"
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

	// 注册事务处理器
	vermouth.RegisterAop("*.*", 0, func(aopContext *vermouth.AopContext) {
		var tx *sql.Tx
		defer func() {
			if err := recover(); err != nil {
				// 判断是否是自定义错误
				if myErr, ok := err.(*MyError); ok {
					aopContext.GinContext.JSON(myErr.Code, myErr.Message)
					return
				}
				if tx != nil {
					// 回滚事务
					fmt.Println("回滚事务")
					err = tx.Rollback()
					if err != nil {
						fmt.Println("回滚失败")
					}
				}
				// 不是我的异常，抛回给中间件处理
				// panic(err)
			}
		}()
		// 判断是否需要事务
		if aopContext.MethodInformation.Transaction {
			// 开启事务
			fmt.Println("开启事务")
			var err error
			tx, err = vermouth.GetDB().Begin()
			if err != nil {
				panic(err)
			}
			aopContext.GinContext.Set("Vermouth:tx", tx)
		}
		// 执行方法
		defer func() {
			if tx != nil {
				// 提交事务
				fmt.Println("提交事务")
				tx.Commit()
			}
		}()
		aopContext.Call()
		// 提交事务
	})

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
