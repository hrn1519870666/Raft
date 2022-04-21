/**
  @author: 黄睿楠
  @since: 2022/4/21
  @desc: 初始化配置信息
**/

package main

//节点数量
var raftCount = 3

//节点集   key:节点编号  value:端口号
var nodeTable map[string]string

//选举超时时间（单位：秒）
var timeout = 3

//心跳检测超时时间
var heartBeatTimeout = 7

//心跳检测频率（单位：秒）
var heartBeatTimes = 3

//用于存储消息
var MessageStore = make(map[int]string)