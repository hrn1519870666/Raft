/**
  @author: 黄睿楠
  @since: 2022/4/21
  @desc: RPC
**/

package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/rpc"
	"time"
)

//rpc服务注册
func rpcRegister(raft *Raft) {
	//注册一个raft节点
	err := rpc.Register(raft)
	if err != nil {
		log.Panic(err)
	}
	port := raft.node.Port
	//把服务绑定到http协议上
	rpc.HandleHTTP()
	//监听端口
	err = http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println("注册rpc服务失败", err)
	}
}

func (rf *Raft) broadcast(method string, args interface{}, fun func(ok bool)) {

	// 遍历所有node
	for nodeID, port := range nodeTable {
		//不用自己给自己广播
		if nodeID == rf.node.ID {
			continue
		}
		//连接远程rpc
		rp, err := rpc.DialHTTP("tcp", "127.0.0.1"+port)
		if err != nil {
			fun(false)
			continue
		}

		var bo = false
		// rpc调用以参数形式传入的method方法
		err = rp.Call(method, args, &bo)
		if err != nil {
			fun(false)
			continue
		}
		fun(bo)
	}
}

// 以下三个方法，通过rf.broadcast(method string, args interface{}, fun func(ok bool))形式进行调用
//投票
func (rf *Raft) Vote(leaderId string, b *bool) error {
	if rf.votedFor != "-1" || rf.currentLeader != "-1" {
		*b = false
	} else {
		rf.setVotedFor(leaderId)
		fmt.Printf("投票成功，已投%s节点\n", leaderId)
		*b = true
	}
	return nil
}

//选举成功后，首次确认领导者
func (rf *Raft) ConfirmationLeader(node NodeInfo, b *bool) error {
	rf.setCurrentLeader(node.ID)
	*b = true
	fmt.Println("已发现网络中的领导节点，", node.ID, "成为了领导者！\n")
	rf.reDefault()
	return nil
}

//http.go调用，追随者接收领导者传来的消息，然后存储到消息池中，待领导者确认后打印
func (rf *Raft) ReceiveMessage(message Message, b *bool) error {
	fmt.Printf("接收到领导者节点发来的信息，消息id为：%d\n", message.MsgID)
	MessageStore[message.MsgID] = message.Msg
	*b = true
	fmt.Println("已回复接收到消息，待领导者确认后打印")
	return nil
}

// 日志复制
// http.go调用，领导者收到了追随者节点转发过来的消息
func (rf *Raft) LeaderReceiveMessage(entries []Log, b *bool) error {
	fmt.Printf("领导者节点接收到转发过来的消息")
	rf.log=append(rf.log,entries...)
	*b = true
	fmt.Println("准备将消息进行广播...")
	num := 0
	go rf.broadcast("Raft.AppendEntries", entries, func(ok bool) {
		if ok {
			num++
		}
	})

	for {
		//自己默认收到了消息，所以减去一
		if num > raftCount/2-1 {
			fmt.Printf("全网已超过半数节点接收到消息，raft验证通过，可以打印消息")
			fmt.Println("消息为：")
			rf.lastMessageTime = millisecond()
			fmt.Println("准备将消息提交信息发送至客户端...")
			go rf.broadcast("Raft.ConfirmationMessage", entries, func(ok bool) {
			})
			break
		} else {
			//休息会儿
			time.Sleep(time.Millisecond * 100)
		}
	}
	return nil
}

// rpc.go LeaderReceiveMessage调用
//追随者节点的反馈得到领导者节点的确认，开始打印消息
func (rf *Raft) ConfirmationMessage(message Message, b *bool) error {
	go func() {
		for {
			if _, ok := MessageStore[message.MsgID]; ok {
				fmt.Printf("raft验证通过，可以打印消息，消息id为：%d\n", message.MsgID)
				fmt.Println("消息为：", MessageStore[message.MsgID], "\n")
				rf.lastMessageTime = millisecond()
				break
			} else {
				//如果没有此消息，等一会再看看
				time.Sleep(time.Millisecond * 10)
			}

		}
	}()
	*b = true
	return nil
}


// 日志复制
func (rf *Raft) AppendEntries (term,prevLogIndex,prevLogTerm,leaderCommit int,entries []Log,leaderId string) (int,bool) {
	fmt.Println("准备将消息进行广播...")

	if term < rf.currentTerm {
		return rf.currentTerm,false
	}

	for index, value := range rf.log {
		if prevLogIndex == value.index && prevLogTerm == value.term {
			for i:=0;i<=index;i++ {
				//如果一个已经存在的条目和新条目发生了冲突（因为索引相同，任期不同）
				if rf.log[i].term != term {
					//那么就删除这个已经存在的条目以及它之后的所有条目
					rf.log=append(rf.log[:i])
				}
			}

			//追加日志中尚未存在的任何新条目
			rf.log=append(rf.log,entries...)

			//如果领导人的已知已提交的最高日志条目的索引大于接收者的已知已提交最高日志条目的索引（leaderCommit > commitIndex）
			//则把接收者的已知已经提交的最高的日志条目的索引commitIndex 重置为
			//领导人的已知已经提交的最高的日志条目的索引 leaderCommit 或者是 上一个新条目的索引 取两者的最小值
			if leaderCommit > rf.commitIndex {
				rf.commitIndex = int(math.Min(float64(leaderCommit),float64(len(rf.log)-1)))
			}
			rf.setCurrentLeader(leaderId)
			rf.lastHeartBeartTime = millisecond()
			fmt.Printf("接收到来自领导节点%s的心跳检测\n", leaderId)
			fmt.Printf("当前时间为:%d\n\n", millisecond())
			return rf.currentTerm,true
		}
	}
	return rf.currentTerm,false
}

//请求投票
func (rf *Raft) RequestVote (term,lastLogIndex,lastLogTerm int,candidateId string) (int,bool) {

	//如果term < currentTerm返回 false
	if term < rf.currentTerm {
		return rf.currentTerm,false
	}

	//如果 votedFor 为空或者为 candidateId，并且候选人的日志至少和自己一样新，那么就投票给他
	if (rf.votedFor == "" || rf.votedFor == candidateId) && rf.log[len(rf.log)-1].index <= lastLogIndex && rf.log[len(rf.log)-1].term <= lastLogTerm{
		rf.votedFor=candidateId
		return rf.currentTerm,true
	}
	return rf.currentTerm,false
}

