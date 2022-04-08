package handler

import (
	. "broadcast-websocket/config"
	. "broadcast-websocket/models"
	"context"
	uuid "github.com/satori/go.uuid"
	"io"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"strings"
	"time"
)

type clientManager struct {
	clients    map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
}
type Message struct {
	Status         string    `json:"status,omitempty"`
	Content        string    `json:"content,omitempty"`
	ClientSendTime time.Time `json:"client_send_time"`
}

func NewClientManager() clientManager {
	return clientManager{
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}
func (manager *clientManager) WsHandle(w http.ResponseWriter, req *http.Request) {
	//解析一个连接
	//conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	conn, err := websocket.Accept(w, req, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		//http.NotFound(res, req)
		//http 请求一个输出
		io.WriteString(w, "这是一个websocket,不是网站.")
		return
	}

	uid := uuid.NewV4()
	sha1 := uid.String()

	//初始化一个客户端对象
	client := &Client{id: sha1, socket: conn, send: make(chan *Message)}
	//把这个对象发送给 管道
	log.Printf("把这个对象发送给注册管道...%v\n", client)
	manager.register <- client

	log.Println("起一个写协程...")
	go client.Write(req.Context(), *manager)
	log.Println("监控连接是否断线...")
	client.Read(req.Context(), *manager)
}

func (manager *clientManager) Start() {
	for {
		select {
		case conn := <-manager.register: //新客户端加入
			log.Println("新客户端加入")
			manager.clients[conn] = true
			log.Println("总客户端数：", len(manager.clients))

			//jsonMessage, _ := json.Marshal(&Message{Status: "ok", ClientSendTime: time.Now()})
			msg := &Message{Status: "ok", ClientSendTime: time.Now()}
			conn.send <- msg
			//manager.send(jsonMessage, conn) //调用发送
		case conn := <-manager.unregister:
			log.Println("客户端离线")
			if _, ok := manager.clients[conn]; ok {
				log.Printf("关闭通道... %v\n", conn)
				close(conn.send)
				//fmt.Printf("关闭通道后，移除连接\n")
				delete(manager.clients, conn)
				log.Println("移除后总客户端数：", len(manager.clients))
				//jsonMessage, _ := json.Marshal(&Message{Content: "a socket has disconnected."})
				//manager.send(jsonMessage, conn)
			}
		case message := <-manager.broadcast: //读到广播管道数据后的处理
			//fmt.Println("消息内容：" + string(message))
			for conn := range manager.clients {
				log.Println("每个客户端发送数据：", conn.id)

				select {
				case conn.send <- message:
					//fmt.Println("发送到每个每个客户端 ")
				default:
					log.Println("要关闭连接啊")
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
			//fmt.Printf("共发送%d个客户端", len(manager.clients))
		}
	}
}
func (manager *clientManager) RedisSend() {
	var ctx = context.Background()
	redisSubscribe := RedisClient.Subscribe(ctx, ViperConfig.Redis.Key)
	_, err := redisSubscribe.Receive(ctx)
	if err != nil {
		log.Printf("redis 订阅出错 %v\n", err)
		log.Fatal(err)
	}
	for msg := range redisSubscribe.Channel() {
		msg.Payload = strings.Trim(msg.Payload, "\"")
		//jsonMessage, _ := json.Marshal(&Message{Status: "ok", Content: msg.Payload, ClientSendTime: time.Now()})
		sendMsg := &Message{Status: "ok", Content: msg.Payload, ClientSendTime: time.Now()}
		log.Printf("redis读取数据：channel=%s\n", msg.Channel)
		//log.Printf("redis读取数据：channel=%s message=%s\n", msg.Channel, msg.Payload)
		manager.broadcast <- sendMsg //激活start 程序 入广播管道
	}

}
