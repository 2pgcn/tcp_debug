package server

import (
	"bufio"
	"context"
	"github.com/2pgcn/tcp_debug/api"
	"go.uber.org/zap"
	"io"
	"net"
	"time"
)

type Uid int64
type Conn struct {
	ctx     context.Context
	Uid     Uid
	log     *zap.SugaredLogger
	conn    *net.TCPConn
	rr      *bufio.Reader
	rw      *bufio.Writer
	sendMsg chan *api.Msg
	rcvChan chan *api.Msg
}

func NewConn(uid Uid) *Conn {
	return &Conn{
		Uid:     uid,
		sendMsg: make(chan *api.Msg, 1024),
		rcvChan: make(chan *api.Msg, 1024),
	}
}

// 登陆成功后启动
func (c *Conn) Start() {
	go func() {
		for {
			select {
			case msg := <-c.sendMsg:
				if msg.Op == api.Op_CLOSEREPLY {
					c.Close()
					return
				}
				//todo error for 指数退让处理
				err := msg.WriteTcp(c.rw)
				if err != nil && err != io.EOF {
					c.log.Errorf("send msg error:%s", err.Error())
					time.Sleep(time.Second * 5)
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case msg := <-c.rcvChan:
				if msg.Op == api.Op_MSGREQ {
					err := msg.WriteTcp(c.rw)
					c.log.Errorf("send msg error:%s", err.Error())
				}
				msg.Op = api.Op_AUTHEEPLY
				c.SendMsg(msg)
			}
		}
	}()
}

func (c *Conn) SendMsg(msg *api.Msg) {
	c.sendMsg <- msg
}

func (c *Conn) Close() {
	msg := &api.Msg{
		Op: api.Op_CLOSEREPLY,
	}
	c.SendMsg(msg)
}
