package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"lrp/internal/conn"
	"net"
	"unsafe"
)

//Str2Bytes string转byte 通过修改结构避免拷贝
func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

//Bytes2Str byte转string
func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// IntToBytes int类型转为byte类型
func IntToBytes(n int) []byte {
	data := int64(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

//BytesToInt 字节转int类型
func BytesToInt(bys []byte) int {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}

//AddrStringToByte 网络地址 “1.1.1.1:80” -> {1,1,1,1,11,11} 端口大端序
func AddrStringToByte(addr string, netType string) ([]byte, error) {
	if netType == "tcp" {
		address, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			fmt.Println("地址格式转换失败")
			return nil, err
		}
		port := make([]byte, 2)
		binary.BigEndian.PutUint16(port, uint16(address.Port))
		result := append(address.IP.To4(), port...)
		return result, nil
	} else {
		address, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			fmt.Println("地址格式转换失败")
			return nil, err
		}
		port := make([]byte, 2)
		binary.BigEndian.PutUint16(port, uint16(address.Port))
		result := append(address.IP.To4(), port...)
		return result, nil
	}
}

//AddrStringToByte 网络地址 {1,1,1,1,11,11} -> “1.1.1.1:80” 端口大端序
func AddrByteToString(addr []byte) string {
	ip := net.IP(addr[0:4])
	port := binary.BigEndian.Uint16(addr[4:6])
	destAddr := ip.String() + ":" + fmt.Sprint(port)
	return destAddr
}

//IsChanClosed 判断channel关闭
func IsChanClosed(ch <-chan int) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

func EncodeSend(conn conn.Conn, data []byte) error {
	if _, err := conn.Write(append(IntToBytes(len(data)), data...)); err != nil {
		return err
	}
	return nil
}

func DecodeReceive(reader io.Reader) ([]byte, error) {
	hb := make([]byte, 8)
	if _, err := io.ReadFull(reader, hb); err != nil {
		return nil, err
	}
	//todo header check
	payload := make([]byte, BytesToInt(hb))
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}
