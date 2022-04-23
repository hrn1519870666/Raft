/**
  @author: 黄睿楠
  @since: 2022/4/21
  @desc: 配置信息
**/

package main

//节点数量
var raftCount = 3

//节点集   key:节点编号  value:端口号
var nodeTable map[string]string

//心跳检测超时时间
var heartBeatTimeout = 7

//心跳检测频率（单位：秒）
var heartBeatTimes = 3

//选举超时时间（单位：秒），从一个跟随者变成候选人开始计时，经过选举超时时间后，如果没有成为领导者，则结束选举过程，将自己变回跟随者
var timeout = 3
