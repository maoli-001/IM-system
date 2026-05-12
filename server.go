package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

// 定义Server结构体
type Server struct {
	Ip   string
	Port int
	//在线客户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播channel
	Message chan string
}

// 创建服务器实例
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 给Server 结构体添加一个方法Start来启动服务器接口
func (this *Server) Start() {
	//监听端口
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen error:", err)
		return
	}

	//关闭监听
	defer listener.Close()
	//fmt.Println("服务器启动成功：127.0.0.1:8888")

	//用一个goroutine来监听Message
	go this.ListenMessage()

	//接收客户端连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept error:", err)
			continue
		}

		//处理客户端请求
		go this.Handler(conn)
	}
}

// Go中没有类，给Server结构体添加一个方法Handler来处理客户端上线
func (this *Server) Handler(conn net.Conn) {
	//创建User实例
	user := NewUser(conn)
	//将用户加入在线列表
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	//广播当前用户上线消息
	this.BroadCast(user, "已上线")

	//接收客户端消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				this.BroadCast(user, "下线")
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("conn.Read error:", err)
				return
			}

			//提取用户消息
			msg := string(buf[:n-1])
			//广播消息
			this.BroadCast(user, msg)
		}
	}()
	//当前Handler阻塞
	select {}
}

// 广播
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

// 监听Message
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		//将msg发给在线的user
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}
