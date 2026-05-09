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
### 服务端
创建`server.go`<br>
定义Server结构体<br>
创建服务器实例<br>
启动服务器接口<br>
处理客户端请求<br>

创建`main.go`<br>
