package handler

import (
	. "broadcast-websocket/config"
	. "broadcast-websocket/models"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"io"
	"log"
	"net/http"
)

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}
type Message struct {
	Status  string `json:"recipient,omitempty"`
	Content string `json:"content,omitempty"`
}

func NewClientManager() ClientManager {
	return ClientManager{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}
func (manager *ClientManager) WsHandle(res http.ResponseWriter, req *http.Request) {
	//解析一个连接
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if error != nil {
		//http.NotFound(res, req)
		//http 请求一个输出
		io.WriteString(res, "这是一个websocket,不是网站.")
		return
	}

	uid := uuid.NewV4()
	sha1 := uid.String()

	//初始化一个客户端对象
	client := &Client{id: sha1, socket: conn, send: make(chan []byte)}
	//把这个对象发送给 管道
	fmt.Printf("把这个对象发送给注册管道...%v\n", client)
	manager.register <- client

	fmt.Println("起一个写协程...")
	go client.Write(*manager)
	fmt.Println("监控连接断线...")
	client.Read(*manager)
}

func (manager *ClientManager) Start() {
	for {
		select {
		case conn := <-manager.register: //新客户端加入
			fmt.Println("新客户端加入")
			manager.clients[conn] = true
			fmt.Println("总客户端数：", len(manager.clients))

			jsonMessage, _ := json.Marshal(&Message{Content: "ok"})
			conn.send <- jsonMessage
			//manager.send(jsonMessage, conn) //调用发送
		case conn := <-manager.unregister:
			fmt.Println("客户端离线")
			if _, ok := manager.clients[conn]; ok {
				fmt.Printf("关闭通道... %v\n", conn)
				close(conn.send)
				fmt.Printf("关闭通道后，移除连接\n")
				delete(manager.clients, conn)
				fmt.Println("移除后总客户端数：", len(manager.clients))
				//jsonMessage, _ := json.Marshal(&Message{Content: "a socket has disconnected."})
				//manager.send(jsonMessage, conn)
			}
		case message := <-manager.broadcast: //读到广播管道数据后的处理
			fmt.Println("消息内容：" + string(message))
			for conn := range manager.clients {
				fmt.Println("每个客户端", conn.id)

				select {
				case conn.send <- message:
					//fmt.Println("发送到每个每个客户端 ")
				default:
					fmt.Println("要关闭连接啊")
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		}
	}
}
func (manager *ClientManager) Send() {

	redisSubscribe := RedisClient.Subscribe(ViperConfig.Redis.Key)
	_, err := redisSubscribe.Receive()
	if err != nil {
		fmt.Printf("redis 订阅出错 %v\n", err)
		log.Fatal(err)
	}
	for msg := range redisSubscribe.Channel() {
		jsonMessage, _ := json.Marshal(&Message{Status: "ok", Content: msg.Payload})
		fmt.Printf("redis读取数据：channel=%s message=%s\n", msg.Channel, msg.Payload)
		manager.broadcast <- jsonMessage //激活start 程序 入广播管道
	}

}
