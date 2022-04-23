/**
  @author: 黄睿楠
  @since: 2022/4/21
  @desc: RPC
**/

package main

import (
	"fmt"
	"log"
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

// TODO   第三个参数是一个匿名函数
func (rf *Raft) broadcast(method string, args interface{}, fun func(ok bool)) {

	// 遍历除自己以外的所有node
	for nodeID, port := range nodeTable {
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
		err = rp.Call(method, args, &bo)
		if err != nil {
			fun(false)
			continue
		}
		fun(bo)
	}
}

// 以下三个方法在raft.go中，通过broadcast方法调用
//请求投票RPC
func (rf *Raft) Vote(node NodeInfo, b *bool) error {

	// 加入了安全性验证，比较Leader和接收者的term
	// TODO   bug1:在下面的if判断中加入 || node.currentTerm < rf.node.currentTerm 之后，节点就不投票了
	if rf.votedFor != "-1" || rf.currentLeader != "-1" {
		*b = false
	} else {
		rf.setVoteFor(node.ID)
		fmt.Printf("投票成功，已投%s节点\n", node.ID)
		*b = true
	}
	return nil
}

// 第一次收到Leader的心跳：
func (rf *Raft) ReceiveFirstHeartbeat(node NodeInfo, b *bool) error {
	rf.setCurrentLeader(node.ID)
	rf.reDefault()

	fmt.Println("已发现网络中的领导节点，", node.ID, "成为了领导者！\n")
	*b = true
	return nil
}

// 之后收到Leader的心跳：
func (rf *Raft) ReceiveHeartbeat(node NodeInfo, b *bool) error {
	rf.setCurrentLeader(node.ID)
	rf.lastHeartBeatTime = millisecond()

	fmt.Printf("接收到来自领导节点%s的心跳检测\n", node.ID)
	fmt.Printf("当前时间为:%d\n\n", millisecond())
	*b = true
	return nil
}



// Leader节点的日志复制
// http.go调用，领导者收到了追随者节点转发过来的消息，然后广播到全网
func (rf *Raft) BroadcastMessage(message Message, b *bool) error {
	fmt.Printf("领导者节点接收到客户端的消息\n")

	// 设置日志任期为当前Leader节点的任期
	//TODO   bug2:rf.node.currentTerm不为0，但message.MsgTerm始终为0
	message.setMsgTerm(rf.node.currentTerm)
	fmt.Printf("消息索引为: %d , Leader收到该消息时的任期为: %d , 消息内容为 %s\n",message.MsgID,rf.node.currentTerm,message.Msg)
	rf.MessageStore[message.MsgID] = message
	*b = true
	fmt.Println("准备将消息进行广播...")
	num := 0
	go rf.broadcast("Raft.ReceiveMessage", message, func(ok bool) {
		if ok {
			num++
		}
	})

	for {
		//自己默认收到了消息，所以减1
		if num > raftCount/2-1 {
			fmt.Printf("全网已超过半数节点接收到消息\nraft验证通过，可以打印消息\n")
			fmt.Printf("消息索引为: %d , 消息内容为 %s\n",message.MsgID,message.Msg)
			rf.lastMessageTime = millisecond()
			fmt.Println("通知客户端：消息提交成功")
			go rf.broadcast("Raft.ConfirmationMessage", message, func(ok bool) {
			})
			break
		} else {
			//休息会儿
			time.Sleep(time.Millisecond * 100)
		}
	}
	return nil
}

//rpc.go BroadcastMessage方法调用，追随者接收消息，然后存储到数据库中，待领导者确认后打印
func (rf *Raft) ReceiveMessage(message Message, b *bool) error {
	fmt.Printf("接收到领导者节点发来的信息\n")
	fmt.Printf("消息索引为: %d , 消息内容为 %s\n",message.MsgID,message.Msg)

	/*
	// TODO   bug3:由于bug2无法setMsgTerm，导致下面的安全性验证也无法进行
	//进行安全性验证
	// 比较Leader和接收者的term
	if message.LeaderTerm >= rf.node.currentTerm {
		// 比较日志索引，如果接收者的最后一条日志的索引 = 新日志索引 - 1，则继续比较日志任期
		_,ok1 := rf.MessageStore[message.MsgID-1]
		// 比较日志任期
		ok2 := rf.MessageStore[message.MsgID-1].MsgTerm == message.MsgTerm
		if ok1 && ok2 {
			rf.MessageStore[message.MsgID] = message
			*b = true
			fmt.Println("已回复接收到消息，待领导者确认后打印")
		}
	} else {
		fmt.Println("安全性验证失败,拒绝接收消息")
		// TODO 安全性验证失败的操作:将接收者日志同步成Leader的日志
	}
	*/

	rf.MessageStore[message.MsgID] = message
	*b = true
	fmt.Println("已回复接收到消息，待领导者确认后打印")
	return nil

}

// rpc.go BroadcastMessage调用
//追随者节点的反馈得到领导者节点的确认，开始打印消息
func (rf *Raft) ConfirmationMessage(message Message, b *bool) error {
		for {
			if _, ok := rf.MessageStore[message.MsgID]; ok {
				fmt.Printf("raft验证通过，可以打印消息\n")
				fmt.Printf("消息索引为: %d , 消息内容为 %s\n\n",message.MsgID,message.Msg)
				rf.lastMessageTime = millisecond()
				break
			} else {
				//如果没有此消息，等一会再看看
				time.Sleep(time.Millisecond * 10)
			}

		}
	*b = true
	return nil
}