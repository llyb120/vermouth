package vermouth

import (
	"database/sql"
	"sync"
)

var once sync.Once

func initTransactionManager() {
	once.Do(func() {
		// 注册事务处理器
		RegisterAop("/**", 0, func(aopContext *Context) {
			var tx *sql.Tx
			defer func() {
				if err := recover(); err != nil {
					if tx != nil {
						// 回滚事务
						err = tx.Rollback()
						if err != nil {
							panic(err)
						}
					}

					// 判断是否是自定义错误
					//if myErr, ok := err.(*RuntimeError); ok {
					//	aopContext.GinContext.JSON(myErr.Code, myErr.Message)
					//	return
					//}
					// 不是我的异常，抛回给中间件处理
					panic(err)
				}
			}()
			// 判断是否需要事务
			if aopContext.ControllerInformation.Transaction {
				// 开启事务
				// fmt.Println("开启事务")
				var err error
				tx, err = GetDB().Begin()
				if err != nil {
					// todo 回滚事务
					panic(err)
				}
				aopContext.GinContext.Set("Vermouth:tx", tx)
			}
			// 执行方法
			aopContext.Call()
			// 提交事务
			if tx != nil {
				err := tx.Commit()
				if err != nil {
					panic(err)
				}
			}
		})
	})
}
