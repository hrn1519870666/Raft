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

//跟随者需要遵守的规则
type Follower interface {

}

//响应来自候选人和领导人的请求




//如果在超过选举超时时间的情况之前没有收到当前领导人（即该领导人的任期需与这个跟随者的当前任期相同）的心跳/附加日志
//或者是给某个候选人投了票，就自己变成候选人
func (rf *Raft) becomeCandidate() bool {
	// 生成在[1500,5000)范围内的随机时间
	r := randRange(1500, 5000)
	//经过随机时间后，开始成为候选人
	time.Sleep(time.Duration(r) * time.Millisecond)
	//如果发现本节点已经投过票，或者已经存在领导者，则不用变身候选人状态
	if rf.state == 0 && rf.currentLeader == "-1" && rf.votedFor == "-1" {
		//将节点状态变为1
		rf.setStatus(1)
		//为自己投票
		rf.setVotedFor(rf.node.ID)
		//节点任期加1
		rf.setCurrentTerm(rf.currentTerm + 1)
		//当前没有领导
		rf.setCurrentLeader("-1")
		//为自己投票
		rf.voteAdd()
		fmt.Println("本节点已变更为候选人状态")
		fmt.Printf("当前得票数：%d\n", rf.vote)
		//开启选举通道
		return true
	} else {
		return false
	}
}