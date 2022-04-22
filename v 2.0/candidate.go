/**
  @author: 黄睿楠
  @since: 2022/4/22
  @desc:
**/

package main

import (
	"fmt"
	"time"
)

//候选人需要遵守的规则
type Candidate interface {

}



//在转变成候选人后就立即开始选举过程:重置选举超时计时器,发送请求投票的 RPC 给其他所有服务器
//如果接收到大多数服务器的选票，那么就变成领导人
//如果选举过程超时，则返回false
func (rf *Raft) election() bool {
	fmt.Println("开始进行领导者选举，向其他节点进行广播")   //
	go rf.broadcast("Raft.Vote", rf.node, func(ok bool) {
		rf.voteCh <- ok
	})
	for {
		// select中如果没有default语句，则会阻塞等待任一case
		// select语句中除default外，各case的执行顺序是完全随机的
		select {
		case ok := <-rf.voteCh:
			if ok {
				rf.voteAdd()
				fmt.Printf("获得来自其他节点的投票，当前得票数：%d\n", rf.vote)
			}
			if rf.vote > raftCount/2 && rf.currentLeader == "-1" {
				fmt.Println("获得超过网络节点二分之一的得票数，本节点被选举成为了leader")
				rf.becomeLeader()

				fmt.Println("向其他节点进行广播...")
				//使其他节点更新领导者信息，并切换为跟随者
				go rf.broadcast("Raft.ConfirmationLeader", rf.node, func(ok bool) {
					fmt.Println(ok)
				})
				//开启心跳检测通道
				rf.heartBeat <- true
				return true
			}
			// 若该节点没有被选举为Leader，比如票数不够，或其他节点抢先成为Leader，则阻塞

		case <-time.After(time.Second * time.Duration(timeout)):
			fmt.Println("领导者选举超时，节点变更为追随者状态\n")
			rf.reDefault()
			return false
		}
	}
}

func (rf *Raft) becomeLeader() {
	//节点状态变为2，代表leader
	rf.setStatus(2)
	//当前领导者为自己
	rf.setCurrentLeader(rf.node.ID)

	//TODO   初始化领导者的特有属性 nextIndex []int 和 matchIndex []int
}



