package api

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
)

var (
	_headDataLen = 2
	_opLen       = 2
	_headLen     = _headDataLen + _opLen

	_packOffSet    = 0
	_headLenOffset = _packOffSet + _headDataLen
	_opLenOffset   = _headLenOffset + _opLen
)

var MsgNullErr = errors.New("server msg is nil")

func (msg *Msg) ReadTcp(rr *bufio.Reader) (err error) {
	var (
		packLen [4]byte
		dataLen uint16
	)
	dataLenBuf := packLen[:_headLen]
	if _, err = io.ReadFull(rr, dataLenBuf); err != nil {
		return
	}
	dataLen = binary.BigEndian.Uint16(dataLenBuf[:_headLenOffset]) - uint16(_headLen)
	bodyData := make([]byte, dataLen)
	if _, err = io.ReadFull(rr, bodyData); err != nil {
		return
	}
	msg.Op = Op(binary.BigEndian.Uint16(dataLenBuf[_headLenOffset:_opLenOffset]))
	msg.Len = int32(dataLen)
	msg.Body = string(bodyData)
	return err
}

func (msg *Msg) WriteTcp(wr *bufio.Writer) (err error) {
	var (
		packLen uint16
	)
	packLen = uint16(_headLen + len(msg.Body))
	buf := make([]byte, packLen)
	binary.BigEndian.PutUint16(buf[_packOffSet:], packLen)
	binary.BigEndian.PutUint16(buf[_headLenOffset:], uint16(msg.Op))
	copy(buf[_opLenOffset:], msg.Body)
	if err = binary.Write(wr, binary.BigEndian, buf); err != nil {
		return
	}
	return wr.Flush()
}
