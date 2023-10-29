package client

import (
	"bufio"
	"context"
	"fmt"
	"github.com/2pgcn/tcp_debug/api"
	"github.com/2pgcn/tcp_debug/conf"
	"log"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var lastUid int64 = 0
var conns []net.Conn
var connsLock sync.Mutex
var (
	countDown  int64
	aliveCount int64
)

// 客户端简单代码,方便压测
func ClientTcp(ctx context.Context, c *conf.Client) {
	go result()
	for j := 0; j < int(c.StartNum); j++ {
		lastUid++
		conn, err := net.DialTimeout("tcp", c.DailUrl, time.Second*1)
		if err != nil {
			log.Println(err)
			continue
		}
		atomic.AddInt64(&aliveCount, 1)
		connsLock.Lock()
		conns = append(conns, conn)
		connsLock.Unlock()
		go clinetStart(ctx, conn, lastUid)
	}
	select {
	case <-ctx.Done():
		connsLock.Lock()
		for _, v := range conns {
			v.Close()
		}
		connsLock.Unlock()
	}
	afterExit := time.After(time.Second * 5)
	for {
		select {
		case <-afterExit:
			return
		default:
			if atomic.LoadInt64(&aliveCount) == int64(0) {
				return
			}
			time.Sleep(time.Second * 1)
		}
	}
}

func clinetStart(ctx context.Context, conn net.Conn, uid int64) {
	rr := bufio.NewReader(conn)
	rw := bufio.NewWriter(conn)
	//发送auth
	msg := &api.Msg{}
	msg.Op = api.Op_AUTHREQ
	msg.Body = strconv.FormatInt(uid, 10)
	err := msg.WriteTcp(rw)
	if err != nil {
		log.Println("auth error:%s", err)
	}
	msg.Reset()
	//for {
	//	err = msg.ReadTcp(rr)
	//	if err != nil {
	//		log.Println(err)
	//		break
	//	}
	//	if err == net.ErrClosed {
	//		atomic.AddInt64(&aliveCount, -1)
	//		return
	//	}
	//	if msg.Op == api.Op_AUTHEEPLY {
	//		//success
	//		break
	//	}
	//}
	////发送一个消息
	//msg.Op = api.Op_MSGREQ
	//msg.Body = "hello world"
	//err = msg.WriteTcp(rw)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	for {
		msg.Reset()
		atomic.AddInt64(&countDown, 1)
		err = msg.ReadTcp(rr)
		if err != nil {
			log.Println(err)
			break
		}
		if err == net.ErrClosed {
			atomic.AddInt64(&aliveCount, -1)
			return
		}
		msg.Reset()
		msg.Op = api.Op_MSGREQ
		err = msg.WriteTcp(rw)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(time.Second * 1)
	}
}

func result() {
	var (
		lastTimes int64
		interval  = int64(5)
	)
	for {
		nowCount := atomic.LoadInt64(&countDown)
		nowAlive := atomic.LoadInt64(&aliveCount)
		diff := nowCount - lastTimes
		lastTimes = nowCount
		fmt.Println(fmt.Sprintf("%s alive:%d down:%d down/s:%d", time.Now().Format("2006-01-02 15:04:05"), nowAlive, nowCount, diff/interval))
		time.Sleep(time.Second * time.Duration(interval))
	}
}
