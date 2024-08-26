package vermouth

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

var urlCoverCache = sync.Map{}

func CoverUrlMiddleware(logPath string) gin.HandlerFunc {
	// 如果没有该目录则创建
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		os.MkdirAll(logPath, 0755)
	}
	return func(c *gin.Context) {
		url := c.Request.URL.String()
		// 去除query
		url = strings.Split(url, "?")[0]
		var coveredUrl string
		if value, ok := urlCoverCache.Load(url); ok {
			coveredUrl = value.(string) // 类型断言
		} else {
			c.Next()
			return
		}

		// 只有指定请求才处理
		// 复制请求的 body
		var bodyBytes []byte
		if c.Request.Method == http.MethodPost {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body.Close()                                    // 关闭原始 body
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 创建新的 body
		}

		// 复制请求的 header
		fullUrl := fmt.Sprintf("%s://%s%s?%s", func() string {
			if c.Request.TLS != nil {
				return "https"
			}
			return "http"
		}(), c.Request.Host, coveredUrl, c.Request.URL.RawQuery) // 添加查询参数
		newRequest, _ := http.NewRequest(c.Request.Method, fullUrl, bytes.NewBuffer(bodyBytes))
		for key, values := range c.Request.Header {
			for _, value := range values {
				newRequest.Header.Add(key, value)
			}
		}

		// 复制请求的 query 参数
		newRequest.URL.RawQuery = c.Request.URL.RawQuery

		// 继续处理原始请求
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 重新设置原始请求的 body

		// 使用自定义的 ResponseWriter 捕获响应体
		bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bw

		c.Next() // 调用后续处理

		go func() {
			// 发送复制的请求
			client := &http.Client{}
			resp, err := client.Do(newRequest)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError) // 处理错误
				return
			}
			defer resp.Body.Close()
			duplicateResponseBody, _ := io.ReadAll(resp.Body)
			originalResponseBody := bw.body.Bytes() // 获取响应体
			if !bytes.Equal(originalResponseBody, duplicateResponseBody) {
				logData := map[string]interface{}{
					"request": map[string]interface{}{
						"method":  c.Request.Method,
						"url":     c.Request.URL.String(),
						"headers": c.Request.Header,
						"body":    string(bodyBytes),
						"query":   c.Request.URL.RawQuery,
					},
					"original_response":  string(originalResponseBody),
					"duplicate_response": string(duplicateResponseBody),
				}

				// 记录日志
				logJSON, err := json.Marshal(logData)
				if err != nil {
					c.AbortWithStatus(http.StatusInternalServerError) // 处理日志记录错误
					return
				}
				// 在logfile目录按时间戳和uuid写入新的请求文件
				uuid := uuid()
				timestamp := time.Now().Format("2006-01-02_15-04-05")
				filename := fmt.Sprintf("%s/%s_%s.json", logPath, timestamp, uuid)
				err = os.WriteFile(filename, logJSON, 0644)
				if err != nil {
					c.AbortWithStatus(http.StatusInternalServerError) // 处理日志记录错误
					return
				}
			}
		}()

		// 输出原请求结果
		// c.Writer.Write(originalResponseBody)

		// // 比较请求结果
		// if !bytes.Equal(originalResponseBody, duplicateResponseBody) {
		// 	// 处理不匹配的情况
		// 	c.JSON(http.StatusConflict, gin.H{"error": "Responses do not match"})
		// 	return
		// }
	}
}

func uuid() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "" // 处理错误
	}
	// 设置版本和变体
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // 设置版本为0100
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // 设置变体为10

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16])
}
