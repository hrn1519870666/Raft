/**
  @author: 黄睿楠
  @since: 2022/4/21
  @desc: 自定义工具
**/

package main

//Go自带源代码库有两个rand包，同时使用会造成冲突，导入时利用包的别名机制解决此问题
import (
	crypto_rand "crypto/rand"
	"log"
	"math/big"
	math_rand "math/rand"
	"time"
)

//返回一个十位数的随机数，作为消息id
func getRandom() int {
	x := big.NewInt(10000000000)
	for {
		result, err := crypto_rand.Int(crypto_rand.Reader, x)
		if err != nil {
			log.Panic(err)
		}
		if result.Int64() > 1000000000 {
			return int(result.Int64())
		}
	}
}

//产生随机值
func randRange(min, max int64) int64 {
	//用于心跳信号的时间
	math_rand.Seed(time.Now().UnixNano())
	return math_rand.Int63n(max-min) + min
}

//获取当前时间的毫秒数
func millisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
