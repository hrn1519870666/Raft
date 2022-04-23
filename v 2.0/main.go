/**
  @author: 黄睿楠
  @since: 2022/4/21
  @desc:
**/

package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {

	//定义三个节点  节点编号 - 端口号
	nodeTable = map[string]string{
		"A": ":9000",
		"B": ":9001",
		"C": ":9002",
	}

	//运行程序时，指定节点编号:raft.exe A/B/C   os.Args[0]是raft.exe
	if len(os.Args) < 1 {
		log.Fatal("程序参数不正确")
	}

	// A B C
	id := os.Args[1]
	// 创建raft节点实例
	raft := NewRaft(id, nodeTable[id])

	// rpc服务注册
	go rpcRegister(raft)
	//发送心跳,只有当前节点为Leader节点时，才会开启心跳通道,向其他节点发送心跳
	go raft.heartbeat()
	//开启Http监听，这里设置A节点监听来自8080端口的请求
	//注意：测试时，如果关掉了A节点，则接收不到8080端口的请求
	if id == "A" {
		go raft.httpListen()
	}

	// 进行第一次选举
	go raft.startElection()

	//进行超时选举
	for {
		// 0.5秒检测一次
		time.Sleep(time.Millisecond * 5000)
		if raft.lastHeartBeatTime != 0 && (millisecond()-raft.lastHeartBeatTime) > int64(heartBeatTimeout*1000) {
			fmt.Printf("心跳检测超时")
			fmt.Println("即将重新开启选举")

			raft.reDefault()
			raft.setCurrentLeader("-1")
			raft.lastHeartBeatTime = 0

			go raft.startElection()
		}
	}

}