## Raft分布式共识算法的Go语言实现

### 实现功能：

##### Leader选举：

 - 节点间随机成为candidate状态并选举出Leader，且同时仅存在一个Leader
 - Leader节点定时发送心跳检测至其他Follower节点
 - Follower节点们超过一定时间未收到心跳检测，则重新开启选举

##### 日志复制：

 - 客户端通过http发送消息到节点A，如果A不是Leader则转发至Leader节点
 - Leader收到消息后向Follower节点进行广播
 - Follower节点收到消息，反馈给Leader，等待Leader确认
 - Leader收到全网超过二分之一的反馈后，本地进行打印，然后将确认收到反馈的信息提交到Follower节点
 - Follower节点收到确认提交信息后，打印消息



### 运行步骤：

##### 1.编译
```go
 go build -o raft.exe
```

##### 2.开启三个端口，并分别执行raft.exe A 、raft.exe B 、 raft.exe C，代表开启三个节点（初始状态为追随者）
##### 3.三个节点会随机选举出领导者（A节点默认监听来自http的访问）,成功的节点会发送心跳检测到其他两个节点
##### 4.此时打开浏览器用http访问本地节点8080端口，带上节点需要同步打印的消息，比如：
`http://localhost:8080/req?message=我是hrn`
三个节点同时打印消息。
