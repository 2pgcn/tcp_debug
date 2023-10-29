package server

import (
	"bufio"
	"context"
	"errors"
	"github.com/2pgcn/tcp_debug/api"
	"github.com/2pgcn/tcp_debug/conf"
	"go.uber.org/zap"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	ctx     context.Context
	mu      sync.Mutex
	wg      sync.WaitGroup
	c       *conf.Server
	log     *zap.SugaredLogger
	listens []*net.TCPListener
	isClose bool
	users   map[Uid]*Conn
}

func ServerTcp(ctx context.Context, c *conf.Server) *Server {
	s := &Server{
		ctx:     ctx,
		wg:      sync.WaitGroup{},
		c:       c,
		isClose: false,
		mu:      sync.Mutex{},
		users:   make(map[Uid]*Conn, 1024),
	}
	devLog, _ := zap.NewProduction()
	s.log = devLog.Sugar()
	for i := 0; i < len(c.Bind); i++ {
		addr, err := net.ResolveTCPAddr("tcp", c.Bind[i])
		if err != nil {
			panic(err)
		}
		listen, err := net.ListenTCP("tcp", addr)
		if err != nil {
			panic(err)
		}
		s.listens = append(s.listens, listen)
		go func() {
			s.wg.Add(1)
			s.listener(listen)
			s.wg.Done()
		}()
	}
	select {
	case <-ctx.Done():
		s.close()
	}
	s.wg.Wait()
	return s
}

func (s *Server) listener(listener *net.TCPListener) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			s.log.Errorf("conn error:%s", err.Error())
		}
		//判断是否close
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				s.log.Infof("net.ErrClosed listener is close")
				return
			}
			s.log.Errorf("AcceptTCP error:%s", err)
			continue
		}
		s.log.Debugf("start accept listener conn")
		go s.startConn(conn)
	}
}

func (s *Server) startConn(conn *net.TCPConn) {
	//登陆逻辑
	rr := bufio.NewReader(conn)
	rw := bufio.NewWriter(conn)
	msg := &api.Msg{}
	var uid Uid
	for {
		err := msg.ReadTcp(rr)
		if err != nil {
			s.log.Infof("read tcp error")
		}
		if msg.Op == api.Op_AUTHREQ {
			var uidInt int64
			uidInt, err = strconv.ParseInt(msg.GetBody(), 10, 64)
			if err != nil {
				s.log.Infof("auth error:body data err:%s", err)
				continue
			}
			uid = Uid(uidInt)
			break
		} else {
			msg.Reset()
			msg.Op = api.Op_AUTHEEPLY
			msg.Body = "need auth"
			err = msg.WriteTcp(rw)
			if err != nil {
				//todo 增加一个失败暂存区
				s.log.Errorf("send AUTHEEPLY error:%s", err)
			}
		}
	}
	c := NewConn(uid)
	c.rw = rw
	c.rr = rr
	c.log = s.log
	s.mu.Lock()
	s.users[uid] = c
	s.mu.Unlock()
	c.Start()
	msg.Reset()
	msg.Op = api.Op_MSGREQ
	msg.Body = "hello world"
	_ = msg.WriteTcp(c.rw)
	for {
		err := msg.ReadTcp(rr)
		if err != nil && err != io.EOF {
			s.log.Infof("read tcp error：%s", err)
			s.mu.Lock()
			delete(s.users, c.Uid)
			s.mu.Unlock()
			break
		}
		c.SendMsg(msg)
	}
}

func (s *Server) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isClose = true
	for i := 0; i < len(s.listens); i++ {
		s.listens[i].Close()
		s.log.Infof("cur listen:%d is closed", i)
	}
	for k, v := range s.users {
		v.Close()
		//todo 根据对端或者添加最后关闭时间定时器扫描
		delete(s.users, v.Uid)
		s.log.Debugf("cur users uid:%s is close", k)
	}
	for len(s.users) != 0 {
		s.log.Infof("cur users num is %d", len(s.users))
		time.Sleep(time.Second * 3)
	}
	s.log.Infof("cur user len is:%d", len(s.users))
}
