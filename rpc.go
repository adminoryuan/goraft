package main

import (
	"fmt"
	"net"
	"net/rpc"
	"strings"
	"time"
)

type Server struct {
	serverid int
	core     *raftCore
}
type Heartbeats struct {
	id   int
	term int
}

type RequestVoteArgs struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

func NewServer(serverid int) *Server {
	s := new(Server)
	s.serverid = serverid
	fmt.Println(s.serverid)
	return s
}
func (s *Server) StartServer() {

	rpcServer := rpc.NewServer()

	rpcServer.RegisterName("rpc", s)

	s.core = NewRaftCore(s.serverid, s)

	l, err := net.Listen("tcp", getCurrAddr(s.serverid))
	fmt.Println(getCurrAddr(s.serverid))

	if err != nil {
		fmt.Println("链接关闭")
	}

	go func() {
		for {
			listen, err := l.Accept()
			if err != nil {
				fmt.Println("error")
				continue
			}
			rpcServer.ServeConn(listen)
		}
	}()
}
func (s *Server) Heartbeats(h Heartbeats, r *RequestVoteReply) {
	s.core.lock.Lock()
	r.Term = s.core.currentTerm
	if s.core.currentTerm < h.term {
		r.VoteGranted = false
	} else if s.core.leaderid == -1 {
		r.VoteGranted = true
		s.core.leaderid = h.id
	} else {
		r.VoteGranted = true
		s.core.electionResetEvent = time.Now()
	}
	s.core.lock.Unlock()
}

// leader 已经选出
func (s *Server) RequestLeader(args RequestVoteArgs, reply *RequestVoteReply) error {
	s.core.lock.Lock()
	s.core.leaderid = args.CandidateId
	*&reply.VoteGranted = true
	s.core.setDefault()
	s.core.lock.Unlock()
	return nil
}
func (s *Server) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	s.core.lock.Lock()
	if s.core.voteFor == -1 {
		s.core.voteFor = args.CandidateId
		reply.VoteGranted = true
		s.core.electionResetEvent = time.Now()
	} else {
		reply.VoteGranted = false
	}

	reply.Term = args.Term
	s.core.lock.Unlock()
	return nil
}

// 转发请求到所有节点
func (s *Server) ForwardAllNode(method string, args interface{}, fun func(e error, reply interface{})) {
	fmt.Println(fmt.Sprintf("群发rpc请求%s", method))
	for k := range config {
		if k == s.core.id {
			continue
		}
		cli, err := rpc.DialHTTP("tcp", config[k])
		if err != nil {
			continue
		}
		var r *RequestVoteReply
		if strings.EqualFold(method, "rpc.RequestVote") {
			r = new(RequestVoteReply)
		}
		err = cli.Call(method, args, &r)
		cli.Close()
		fun(err, r)
	}
}
