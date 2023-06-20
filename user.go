package main

import (
	"net"
	"strings"
)


type User struct{
	Name string
	Addr string
	C chan string
	conn net.Conn	
	server *Server
}

//创建用户
func NewUser(conn net.Conn,server *Server)*User{
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C: 	  make(chan string),
		conn: conn,
		server: server,
	}
	//启动监听goroutine
	go user.ListenMessage()

	return user
}

//用户上线
func(this *User)Online(){
	//把用户加入OnlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	this.server.BroadCast(this,"已上线")
}

//用户下线
func(this *User)Offline(){
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap,this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this,"已下线")
}

//处理消息
func(this *User)DoMessage(msg string){
	//msg=who时查询在线用户
	if msg == "who"{
		this.server.mapLock.Lock()
		for _,user := range this.server.OnlineMap{
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
            this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	}else if len(msg) > 7 && msg[:7] == "rename|"{
		//renam|newname 
		newName := strings.Split(msg,"|")[1]
		//判断名字是否存在
		_,ok := this.server.OnlineMap[newName]
		if ok{
			this.SendMsg("该用户名已经被使用\n")
		}else{
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap,this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.SendMsg("您已经更新用户名：" + this.Name + "\n")
		}
	}else if len(msg) > 4 && msg[:3] == "to|"{
		//格式 to|name|msg

		//1. 获取用户名
		remoteName := strings.Split(msg,"|")[1]
		if remoteName == ""{
			this.SendMsg("消息格式不正确\n")
			return
		}

		//2.检查是否存在
		remoteUser,ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("用户不存在\n")
			return
		}
		//3.发送消息
		content := strings.Split(msg,"|")[2]
		if content == ""{
			this.SendMsg("不能发送空消息\n")
			return
		}

		remoteUser.SendMsg(this.Name+"对您说:"+content)
	}else{
		this.server.BroadCast(this,msg)
	}

}

//像客户端发送消息API
func(this *User)SendMsg(msg string){
	this.conn.Write([]byte(msg))
}


//监听当前User channel
func(this *User)ListenMessage(){
	for{
		msg := <-this.C

		this.conn.Write([]byte(msg+"\n"))
	}
}

