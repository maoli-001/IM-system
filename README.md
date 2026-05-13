# IM-system

## 前置知识
### tcp编程
#### TCP 是什么？
面向连接（必须先建立连接才能通信）<br>
可靠传输（数据不丢、不乱序）<br>
全双工（两边可同时收发）<br>

**TCP 服务端流程：监听 → 接收连接 → 协程处理**
**TCP 客户端流程：连接服务端 → 发送数据 → 接收数据**

#### 核心API
|函数|作用|
|:--:|:--:|
|net.Listen("tcp", ":8080")	|创建 TCP 监听|
|listener.Accept()	|等待客户端连接|
|net.Dial("tcp", "ip:port")	|客户端连接服务端|
|conn.Write()	|发送数据|
|conn.Read()	|接收数据|
|conn.Close()	|关闭连接|
|conn.RemoteAddr()	|获取客户端地址|

## Start
### 构建基础server
创建`server.go`<br>
定义Server结构体<br>
```
type Server struct {
	Ip   string
	Port int
}
```
创建服务器实例<br>
启动服务器接口<br>
处理客户端请求<br>

创建`main.go`<br>

### 用户上线广播
新建`user.go`，定义User结构体
```
// 定义User结构体
type User struct {
	Name string
	Addr string
	C chan string
	conn net.Conn
}
```
创建User实例<br>
监听channel<br>

新增Server结构体属性
```
// 定义Server结构体
type Server struct {
	Ip   string
	Port int
	//在线客户列表
	OnlineMap map[string]*User
	maplock sync.RWMutex

	//消息广播channel
	Message chan string
}
```

处理用户上线<br>
```
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

	//当前Handler阻塞
	select {}
}
```

广播<br>
```
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

```

监听Message
```
func (this *Server) ListenMessage() {
	for {
	msg:=<-this.Message

		//将msg发给在线的user
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}
```
<img width="864" height="921" alt="image" src="https://github.com/user-attachments/assets/c41d41bf-48a7-4dd1-94a9-1e460d6de82f" />

<img width="789" height="1248" alt="image" src="https://github.com/user-attachments/assets/1f11cdcc-d61a-47e0-9701-b8390007c110" />

### 测试一下
在ubuntu终端打开<br>
```
go init
go mod tidy
go run .
```
再开两个端口，输入`nc 127.0.0.1 8888`模拟用户上线<br>
<img width="771" height="120" alt="image" src="https://github.com/user-attachments/assets/87c778ea-8764-4306-baa2-ea8e749f7e61" />
<img width="756" height="84" alt="image" src="https://github.com/user-attachments/assets/cedcfd5f-eaf3-4853-9cb3-6345a3127964" />
当有一个用户上线时，所有在线用户都将收到消息<br>

### 版本三：用户消息广播功能
`server.go`中`Handler`方法
```
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
```

### 版本四：用户业务逻辑封装
```
// 定义User结构体
type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// 创建User实例
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	//监听当前的channel
	go user.ListenMessage()

	return user
}
```
在`user,go`中user结构体添加一个`server *Server`<br>

把`server.go`中用户上线下线及用户处理消息的逻辑搬到`user.go`<br>
```
/ 用户的上线业务
func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.BroadCast(this, "已上线")
}

// 用户下线的业务
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播当前用户上线消息
	this.server.BroadCast(this, "下线")
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	this.server.BroadCast(this, msg)
}
```
同时修改`server.go`中的客户端处理逻辑
```
func (this *Server) Handler(conn net.Conn) {
	//创建User实例
	user := NewUser(conn, this)
	//将用户加入在线列表
	user.Online()

	//接收客户端消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("conn.Read error:", err)
				return
			}

			//提取用户消息
			msg := string(buf[:n-1])
			//用户针对msg进行处理
			user.DoMessage(msg)
		}
	}()
	//当前Handler阻塞
	select {}
}
```
