package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type LogEntry struct {
	Command interface{}
	Term    int
}

var timeOut time.Duration = time.Duration(150+rand.Intn(150)) * time.Millisecond

// 保存节点身份
type NodeState int

const (
	//跟随者
	Follower  NodeState = iota
	Candidate           //候选者
	Leader              //领导者
	Dead                //死亡了
)

type raftCore struct {
	lock sync.Mutex
	id   int

	currentTerm int

	leaderid int
	voteFor  int

	log []LogEntry

	state NodeState
	s     *Server

	electionResetEvent time.Time //选举时间
}

func NewRaftCore(id int, s *Server) *raftCore {
	core := new(raftCore)
	core.id = id
	core.s = s
	core.leaderid = -1
	core.state = Follower
	core.voteFor = -1
	go func() {
		core.lock.Lock()
		core.electionResetEvent = time.Now()
		core.lock.Unlock()
		core.CheckHeartbeats()
	}()
	return core
}

// 检测心跳
func (r *raftCore) CheckHeartbeats() {
	t := time.NewTicker(time.Microsecond * 3000)
	for {
		<-t.C

		if since := time.Since(r.electionResetEvent); since > timeOut {
			//选举超时，触发选举
			r.setDefault()

			fmt.Println("心跳超时")
			go r.startElection()
			break
		}
	}
}
func (r *raftCore) ProProCandidate() bool {

	SleepRandomTime()
	SleepRandomTime()
	if r.voteFor != -1 || r.leaderid != -1 {

		return false
	}
	r.lock.Lock()
	r.state = Candidate
	r.currentTerm += 1

	r.voteFor = r.id
	r.lock.Unlock()
	return true
}
func (r *raftCore) setDefault() {
	r.lock.Lock()

	r.voteFor = -1
	r.leaderid = -1
	r.state = -1

	r.lock.Unlock()
}

// 开始选举
func (r *raftCore) startElection() {
	for {
		if r.ProProCandidate() {
			if r.Election() {
				fmt.Println("当选了!")
				break
			} else {
				continue
			}
		} else {
			r.CheckHeartbeats()
			break
		}

	}

}

func (r *raftCore) Election() bool {
	r.lock.Lock()
	saveCurrTerm := r.currentTerm
	r.lock.Unlock()

	args := RequestVoteArgs{
		Term:        r.currentTerm,
		CandidateId: r.id,
	}
	currVote := 1
	reschan := make(chan *RequestVoteReply)
	go r.s.ForwardAllNode("rpc.RequestVote", args, func(e error, reply interface{}) {
		res := reply.(*RequestVoteReply)
		reschan <- res
	})
	for {
		select {
		case <-time.After(time.Second * time.Duration(10)):
			r.setDefault()
			fmt.Println("超时！")

			return false
		case res := <-reschan:

			if !res.VoteGranted || res.Term > r.currentTerm {
				//对方任期号大于自己
				r.setDefault()

				return false
			}
			r.lock.Lock()
			if res.Term == saveCurrTerm {
				currVote += 1
				if currVote*2 > getNodeLen() {
					r.lock.Unlock()
					return true
				}
			}
			r.lock.Unlock()
		}
	}
}
func (r *raftCore) becomeFollower(term int) {
	r.lock.Lock()
	r.state = Follower
	r.currentTerm = term
	r.voteFor = -1
	r.electionResetEvent = time.Now()
	r.lock.Unlock()
}

func (r *raftCore) StartLeader() {
	r.lock.Lock()
	r.state = Leader
	r.lock.Unlock()

	args := RequestVoteArgs{
		CandidateId: r.id,
	}
	//转发请求表示我已经成为leader了！
	r.s.ForwardAllNode("rpc.RequestLeader", args, func(e error, reply interface{}) {

	})
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			<-ticker.C
			fmt.Println("发送心跳！！")
			go r.SendHeartbeats()
			r.lock.Lock()
			if r.state != Leader {
				r.lock.Unlock()
				return
			}
			r.lock.Unlock()
		}
	}()
}

func (r *raftCore) SendHeartbeats() {
	h := Heartbeats{}
	go r.s.ForwardAllNode("rpc.Heartbeats", h, func(e error, reply interface{}) {

	})
}
