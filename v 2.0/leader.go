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

//领导者需要遵守的规则
type Leader interface {

}



//一旦成为领导人：发送空的附加日志（AppendEntries）RPC（心跳）给其他所有的服务器；在一定的空余时间之后不停的重复发送，以防止跟随者超时
func (rf *Raft) heartbeat() {
	//如果收到通道开启的信息,将会向其他节点进行固定频率的心跳发送
	if <-rf.heartBeat {
		for {
			fmt.Println("本节点开始发送心跳检测...")
			rf.broadcast("Raft.AppendEntries", rf.node, func(ok bool) {
				fmt.Println("收到回复:", ok)
			})
			rf.lastHeartBeartTime = millisecond()
			// 心跳检测频率
			time.Sleep(time.Second * time.Duration(heartBeatTimes))
		}
	}
}



//如果接收到来自客户端的请求：附加条目到本地日志中，在条目被应用到状态机后响应客户端
//如果对于一个跟随者，最后日志条目的索引值大于等于 nextIndex（lastLogIndex ≥ nextIndex），则发送从 nextIndex 开始的所有日志条目：
//如果成功：更新相应跟随者的 nextIndex 和 matchIndex
//如果因为日志不一致而失败，则 nextIndex 递减并重试
//假设存在 N 满足N > commitIndex，使得大多数的 matchIndex[i] ≥ N以及log[N].term == currentTerm 成立，则令 commitIndex = N