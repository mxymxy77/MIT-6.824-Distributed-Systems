package raft

var (
	electionTimeOutEvent = fsmEvent("election time out")
)

// 添加 FOLLOWER 状态下的处理函数
func (rf *Raft) addFollowerHandler() {
	rf.addHandler(FOLLOWER, electionTimeOutEvent, fsmHandler(startNewElection))
}

// election time out 意味着，
// 进入新的 term
// 并开始新一轮的选举
func startNewElection(rf *Raft, args interface{}) fsmState {
	// 先进入下一个 Term
	rf.currentTerm++
	// 先给自己投一票
	rf.votedFor = rf.me
	// 现在总的投票人数为 1，就是自己投给自己的那一票
	votesForMe := 1

	debugPrintf("[%s]成为 term(%d) 的 candidate ，开始竞选活动", rf, rf.state)

	// 根据自己的参数，生成新的 requestVoteArgs
	// 发给所有人的都是一样的，所以只用生成一份
	requestVoteArgs := rf.newRequestVoteArgs()

	// 通过 requestVoteReplyChan 获取 goroutine 获取的 reply
	requestVoteReplyChan := make(chan *RequestVoteReply, len(rf.peers))
	// 向每个 server 拉票

	for server := range rf.peers {
		// 跳过自己
		if server == rf.me {
			continue
		}
		go func(server int, args *RequestVoteArgs, replyChan chan *RequestVoteReply) {
			// 生成投票结果变量
			reply := new(RequestVoteReply)
			// 拉票
			ok := rf.sendRequestVote(server, args, reply)

			rf.rwmu.RLock()
			// 如果 rf 已经不是 CANDIDATE 了，不用反馈投票结果
			if ok &&
				rf.state == CANDIDATE {
				// 返回投票结果
				replyChan <- reply
			}
			rf.rwmu.RUnlock()
		}(server, requestVoteArgs, requestVoteReplyChan)
	}

	go func(replyChan chan *RequestVoteReply) {
		for {
			select {
			case <-rf.electionTimer.C: // 选举时间结束，需要开始新的选举
				rf.call(electionTimeOutEvent, nil)
				return
			case reply := <-requestVoteReplyChan: // 收到新的选票
if reply.Term > rf.currentTerm {
	rf.call(, args interface{})
}
				if reply.IsVoteGranted {
					// 投票给我的人数 +1
					votesForMe++
					// 如果投票任务过半，那我就是新的 LEADER 了
					if votesForMe > len(rf.peers)/2 {
						rf.call(winThisTermElectionEvent, nil)
						return
					}
				}
			}
		}
	}(requestVoteReplyChan)

	return CANDIDATE
}
