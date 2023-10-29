package api

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestMsg_ReadWriteTcp(t *testing.T) {
	msgStr := "hello world hello world hello world hello world hello world" +
		"hello world hello world hello world hello world hello world" +
		"hello world hello world hello world hello world hello world" +
		"hello world hello world hello world hello world hello world" +
		""
	byteBuf := new(bytes.Buffer)
	msgStrLen := len([]byte(msgStr))
	msgProto := &Msg{
		Op:   Op_AUTHREPLY,
		Len:  int32(msgStrLen),
		Body: msgStr,
	}
	strings.NewReader(msgStr)
	wr := bufio.NewWriter(byteBuf)
	err := msgProto.WriteTcp(wr)
	if err != nil {
		t.Error(err)
	}
	rr := bufio.NewReader(byteBuf)
	err = msgProto.ReadTcp(rr)
	if err != nil {
		t.Error(err)
	}
	if string(msgProto.Body) != msgStr || int(msgProto.Len) != msgStrLen {
		t.Error("msgProto readTcp writeTcp error")
	}
	t.Log("success")
	return
}
