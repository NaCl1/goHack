package webUtils // 修改包名

import (
	"github.com/imroc/req/v3"
	"log"
)

// ReqHttp 是导出的函数（大写）
func ReqHttp(url string) interface{} {
	client := req.C()
	res, err := client.R().Get(url)
	if err != nil {
		log.Println("请求失败:", err)
		return err
	}
	return res
}

// 测试使用可以，正式开发不建议使用，可控性差
// 快速使用，一般用在test的时候
func TestHttp(url string) interface{} {
	res := req.MustGet(url)
	return res
}
