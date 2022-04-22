/**
  @author: 黄睿楠
  @since: 2022/4/21
  @desc: Raft节点类
**/

package main

import (
	"sync"
)

//所有属性，构造器，set方法
type Raft struct {

	//========== 所有服务器上的持久性状态 ==========

	//服务器已知最新的任期（在服务器首次启动时初始化为0，单调递增）
	currentTerm int

	//票投给了谁，也就是当前任期内收到选票的candidateId，如果没有投给任何候选人,则为空
	votedFor string

	//日志条目   index:领导人接收到该条目时的任期（初始索引为1）   message:每个条目包含了用于状态机的命令
	log []Log



	//========== 所有服务器上的易失性状态 ==========

	//已知已提交的最高的日志条目的索引（初始值为0，单调递增）
	commitIndex int

	//已经被应用到状态机的最高的日志条目的索引（初始值为0，单调递增）
	lastApplied int



	//========== 领导人（服务器）上的易失性状态 (选举后已经重新初始化) ==========

	//对于每一台服务器，发送到该服务器的下一个日志条目的索引（初始值为领导人最后的日志条目的索引+1）
	nextIndex []int

	//对于每一台服务器，已知的已经复制到该服务器的最高日志条目的索引（初始值为0，单调递增）
	matchIndex []int



	//========== 自定义的属性 ==========

	// 节点的id和端口号
	node *NodeInfo

	//本节点获得的投票数
	vote int

	//当前节点的领导
	currentLeader string

	//线程锁
	lock sync.Mutex

	//当前节点状态   0 follower  1 candidate  2 leader
	state int

	//发送最后一条消息的时间
	lastMessageTime int64

	//最后一次心跳检测时间
	lastHeartBeartTime int64

	//心跳超时时间(单位：秒)
	timeout int

	//接收投票成功通道
	voteCh chan bool

	//心跳信号
	heartBeat chan bool

}

type NodeInfo struct {
	ID   string
	Port string
}

type Log struct {
	index int
	term int
	message string
}



// 构造器
func NewRaft(id,port string) *Raft {
	node := new(NodeInfo)
	node.ID = id
	node.Port = port

	rf := new(Raft)
	//节点信息
	rf.node = node

	//设置任期
	rf.setCurrentTerm(0)
	//初始化日志
	rf.log=make([]Log,0,logCap)
	//初始化已知已提交的最高的日志条目的索引
	rf.setCommitIndex(0)
	//初始化已经被应用到状态机的最高的日志条目的索引
	rf.setLastApplied(0)
	//当前节点获得票数
	rf.setVote(0)
	//给A B C三个节点投票，给谁都不投
	rf.setVotedFor("-1")
	//0 follower
	rf.setStatus(0)
	//最后一次心跳检测时间
	rf.lastHeartBeartTime = 0
	rf.timeout = heartBeatTimeout
	//最初没有领导
	rf.setCurrentLeader("-1")
	//投票通道
	rf.voteCh = make(chan bool)
	//心跳通道
	rf.heartBeat = make(chan bool)
	return rf
}

//============= set方法，原子操作 =============

//设置任期
func (rf *Raft) setCurrentTerm(term int) {
	rf.lock.Lock()
	rf.currentTerm = term
	rf.lock.Unlock()
}

//设置为谁投票
func (rf *Raft) setVotedFor(id string) {
	rf.lock.Lock()
	rf.votedFor = id
	rf.lock.Unlock()
}

//设置已知已提交的最高的日志条目的索引
func (rf *Raft) setCommitIndex(commitIndex int) {
	rf.lock.Lock()
	rf.commitIndex = commitIndex
	rf.lock.Unlock()
}

//设置已经被应用到状态机的最高的日志条目的索引
func (rf *Raft) setLastApplied(lastApplied int) {
	rf.lock.Lock()
	rf.lastApplied = lastApplied
	rf.lock.Unlock()
}

//设置当前领导者
func (rf *Raft) setCurrentLeader(leader string) {
	rf.lock.Lock()
	rf.currentLeader = leader
	rf.lock.Unlock()
}

//设置当前状态
func (rf *Raft) setStatus(state int) {
	rf.lock.Lock()
	rf.state = state
	rf.lock.Unlock()
}

//投票累加
func (rf *Raft) voteAdd() {
	rf.lock.Lock()
	rf.vote++
	rf.lock.Unlock()
}

//设置投票数量
func (rf *Raft) setVote(num int) {
	rf.lock.Lock()
	rf.vote = num
	rf.lock.Unlock()
}