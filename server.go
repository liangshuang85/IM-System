package main

import (
	"fmt"
	"net"
	"sync"
	"io"
	"time"
)

type Server struct {
	Ip   string
	Port int
	//在线用户列表
	OnlineMap map[string]*User
	//消息广播channel
	Message chan string

	mapLock sync.RWMutex
}

// server 接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
		OnlineMap:make(map[string]*User),
		Message: make(chan string),
	}
	return server
}

//监听Msg
func(this *Server)ListenMessage(){
	for{
		msg := <-this.Message

		//将msg发送给所有在线User
		this.mapLock.Lock()
		for _,cli := range this.OnlineMap{
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}


//广播消息
func(this *Server)BroadCast(user *User,msg string){
	sendMsg := "["+user.Addr+"]"+user.Name+":"+msg

	this.Message <-sendMsg
}


func (this *Server) Handler(conn net.Conn) {
	//当前链接服务
	//fmt.Println("链接建立成功")
	user := NewUser(conn,this)

	//用户上线
	user.Online()

	//监听用户是否活跃
	isLive := make(chan bool)

	//接受客户端发送的消息
	go func(){
		buf := make([]byte,4096)
		for{
			n,err := conn.Read(buf)
			
			if n == 0{
				//用户下线
				user.Offline()
				return
			}

			if err != nil && err != io.EOF{
				fmt.Println("Conn read err",err)
				return
			}
			
			//读取用户信息 并去掉\n
			msg := string(buf[:n-1])

			//将得到的消息广播
			user.DoMessage(msg)

			//有消息代表用户活跃
			isLive <- true
		}
	}()

		//当前handler阻塞
		for{
			select{
			case <- isLive:

			case <- time.After(time.Second * 120):
				user.SendMsg("被踢出会话")

				close(user.C)
				conn.Close()

				return
			}
		}
		
		

}

// 启动服务器接口
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.listen err:", err)
		return
	}
	//close linsten socket
	defer listener.Close()

	//启动监听goroutine
	go this.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listen accept err:", err)
			continue
		}
		//do handler
		go this.Handler(conn)
	}

}

