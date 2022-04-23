## Raft分布式共识算法的Go语言实现

### 算法介绍

[论文（中文翻译版）](https://github.com/maemual/raft-zh_cn/blob/master/raft-zh_cn.md)

[动画演示](http://thesecretlivesofdata.com/raft/)



### V 1.0

#### 实现功能

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



#### 流程图

![在这里插入图片描述](https://img-blog.csdnimg.cn/4b0896ef8a3b43d8b819ce25f716a793.png)



#### 运行步骤

##### 1.编译
```go
 go build -o raft.exe
```

##### 2.开启三个端口，并分别执行raft.exe A 、raft.exe B 、 raft.exe C，代表开启三个节点（初始状态为追随者）
##### 3.三个节点会随机选举出领导者（A节点默认监听来自http的访问）,成功的节点会发送心跳检测到其他两个节点
##### 4.此时打开浏览器用http访问本地节点8080端口，带上节点需要同步打印的消息，比如：
`http://localhost:8080/req?message=我是hrn`
三个节点同时打印消息。



### V 2.0

#### V 1.0存在的问题

若某个节点宕机一段时间，之后再次加入网络，它仍然可以参与Leader选举，但这个节点的数据库中并没有宕机那段时间的日志记录，也就是说它不应该被允许成为Leader节点。

如果由于网络波动等原因，导致某个节点的数据库中的日志出现部分丢失、重复等情况，应该在日志复制时进行处理，使其与Leader节点的日志保持一致。当然，这时该节点也不应该被允许成为Leader节点。



#### 新增功能

安全性模块（选举限制和提交之前任期内的日志条目）：通过比较通信双方节点的term大小，以及日志的index大小，来确保状态机安全。

代码有待完善...
