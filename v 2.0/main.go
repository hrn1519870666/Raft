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

	//运行程序时，指定节点编号:raft.exe B   os.Args[0]是raft.exe
	if len(os.Args) < 1 {
		log.Fatal("程序参数不正确")
	}

	// A B C
	id := os.Args[1]
	// 创建raft节点实例
	raft := NewRaft(id, nodeTable[id])

	// rpc服务注册
	go rpcRegister(raft)
	//发送心跳,只有当前节点为Leader节点时，才会收到通道开启的信息,向其他节点发送心跳
	go raft.heartbeat()
	//开启Http监听，这里设置A节点监听来自8080端口的请求
	if id == "A" {
		go raft.httpListen()
	}

	// 进行第一次选举
	go raft.startElection()

	//进行超时选举
	for {
		// 5000毫秒，即0.5秒检测一次
		time.Sleep(time.Millisecond * 5000)
		if raft.lastHeartBeartTime != 0 && (millisecond()-raft.lastHeartBeartTime) > int64(raft.timeout*1000) {
			fmt.Printf("心跳检测超时，已超过%d秒\n", raft.timeout)
			fmt.Println("即将重新开启选举")

			raft.reDefault()
			raft.setCurrentLeader("-1")
			raft.lastHeartBeartTime = 0

			go raft.startElection()
		}
	}

}

//完整的选举函数:分为两步，先切换为候选人，然后进行选举操作
func (rf *Raft) startElection() {
	for {
		// 将节点状态从 跟随者 切换为 候选人
		if rf.becomeCandidate() {
			//成为候选人节点后 向其他节点要票，如果raft节点顺利成为Leader，则退出for循环，结束选举
			if rf.election() {
				break
			} else {
				continue
			}
		} else {   // 如果raft节点不是候选人，则退出for循环
			break
		}
	}
}