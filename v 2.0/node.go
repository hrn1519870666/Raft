/**
  @author: 黄睿楠
  @since: 2022/4/22
  @desc:
**/

package main

//所有服务器需要遵守的规则
type Node interface {

}

//如果commitIndex > lastApplied，则 lastApplied 递增，并将entries[lastApplied]应用到状态机中
func (rf *Raft) logApply(commitIndex int,entries []Log) {
	if commitIndex > rf.lastApplied {
		rf.setLastApplied(rf.lastApplied+1)
		rf.log=append(rf.log,entries[rf.lastApplied])
	}
}

//如果接收到来自新的领导人的附加日志（AppendEntries）RPC，且接收到的 RPC 请求或响应中，任期号T > currentTerm，则令 currentTerm = T，并切换为跟随者状态
func (rf *Raft) becomeFollower(term int,leaderId string) {
	if term > rf.currentTerm {
		rf.currentTerm =term
		rf.setStatus(0)
		rf.setCurrentLeader(leaderId)
	}
}

//恢复默认设置
func (rf *Raft) reDefault() {
	rf.setVote(0)
	rf.setVotedFor("-1")
	rf.setStatus(0)
}