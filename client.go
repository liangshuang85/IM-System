package main

import(
	"net"
	"fmt"
	"flag"
	"io"
	"os"
)

type Client struct{
	ServerIP string
	ServerPort int
	Name string
	conn net.Conn
	flag int 	//客户端的模式
}

func NewClient(serverIP string,serverPort int) *Client{
	//创建客户端对象
	client := &Client{
		ServerIP : serverIP,
		ServerPort : serverPort,
		flag : 999,
	}

	//连接server
	conn, err := net.Dial("tcp",fmt.Sprintf("%s:%d",serverIP,serverPort))

	if err != nil {
		fmt.Println("net.Dial error:",err)
		return nil
	}
	client.conn = conn

	//返回对象
	return client

}

// 处理server回应的消息，直接显示到标准输出
func (client *Client) DealResponse() {
	// 一旦client.conn有数据，直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

//查询当前哪些用户在线
func(client *Client)SelectUser(){
	sendMsg := "who\n"
	_,err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err: ", err)
		return
	}
}

//菜单
func(client *Client)menu()bool{
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}

}

//公聊模式
func (client *Client) PublicChat() {
	// 提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>请输入聊天内容，exit退出.")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发给服务器
		// 消息不为空立即发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err: ", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出.")
		fmt.Scanln(&chatMsg)
	}
}

//私聊模式
func(client *Client) PrivateChat(){
	var remoteName string
	var chatMsg string

	client.SelectUser()
	fmt.Println(">>>>请输入聊天对象的[用户名], exit退出: ")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>请输入消息内容，exit退出:")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			if len(chatMsg) != 0{
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err: ", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>请输入消息内容，exit退出:")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUser()
		fmt.Println(">>>>请输入聊天对象的[用户名], exit退出: ")
		fmt.Scanln(&remoteName)
	}

}

//更新用户名
func(client *Client)UpdateName()bool{
	fmt.Println(">>>>请输入用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n" // 封装协议
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err: ", err)
		return false
	}
	return true
}



func(client *Client)Run(){
	for client.flag != 0 {
		for !client.menu() {
		}

		// 根据不同的模式处理不同的业务
		switch client.flag {
		case 1:
			// 公聊模式
			client.PublicChat()
		case 2:
			// 私聊模式
			client.PrivateChat()
		case 3:
			// 更新用户名
			client.UpdateName()
		}
	}
	fmt.Println("退出！")
}

var serverIP string
var serverPort int

func init(){
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
    flag.IntVar(&serverPort, "port", 8888, "设置服务器端口(默认是8888)")

	//命令行解析
	flag.Parse()
}

func main(){

	client := NewClient("127.0.0.1",8888)
	if client == nil{
		fmt.Println(">>>>>> 连接服务器失败")
	} 
	fmt.Println(">>>>>> 连接服务器成功")
	
	// 单独开启一个goroutine去处理server的回执消息
	go client.DealResponse()

	client.Run()
}