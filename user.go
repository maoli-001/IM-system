package main

import (
	"net"
)

// 定义User结构体
type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

// 创建User实例
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}
	//监听当前的channel
	go user.ListenMessage()

	return user
}

// 监听user对应的channel消息
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
