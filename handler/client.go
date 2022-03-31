package handler

import (
	"context"
	"errors"
	"log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"
)

type Client struct {
	id     string
	socket *websocket.Conn
	send   chan *Message
}

//监控连接状态
func (c *Client) Read(ctx context.Context, manager clientManager) {

	defer func() {
		manager.unregister <- c
		//c.socket.Close()
		c.socket.Close(websocket.StatusNormalClosure, "正常读关闭")
		log.Printf("defer 读关闭 %v\n", c)
	}()
	var (
		receiveMsg map[string]string
		err        error
	)

	for {
		err = wsjson.Read(ctx, c.socket, &receiveMsg)
		//log.Printf("是在不停的读吗，连接对象？ %v\n", c)
		log.Printf("是在不停的读吗，收到客户端消息？ %v\n", receiveMsg)
		if err != nil {
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				log.Printf("是在不停的读吗，正常关闭")
				break
			}

			log.Printf("是在不停的读吗，读取出错关闭？ %v\n", err)
			break
		}
		sendMsg := &Message{Status: "ok", Content: receiveMsg["content"], ClientSendTime: time.Now()}
		wsjson.Write(ctx, c.socket, sendMsg)

	}
}

//写入管道后激活这个进程
func (c *Client) Write(ctx context.Context, manager clientManager) {
	defer func() {
		manager.unregister <- c
		//c.socket.Close()
		c.socket.Close(websocket.StatusInternalError, "写关闭")
		log.Printf("defer 写关闭了 %v\n", c)
	}()

	for {
		select {
		case message, ok := <-c.send: //这个管道有了数据 写这个消息出去
			log.Printf("读通道状态 %v\n", ok)
			//fmt.Printf("读通道状态连接 %v\n", c)
			if !ok {
				return
			}

			//err := c.socket.WriteMessage(websocket.TextMessage, message)
			//var msg Message
			//json.Unmarshal(message, &msg)
			err := wsjson.Write(ctx, c.socket, message)
			log.Printf("发送数据状态 %v\n", err)
			//fmt.Printf("发送数据状态连接 %v\n", c)
			if err != nil {
				manager.unregister <- c
				//c.socket.Close()
				c.socket.Close(websocket.StatusNormalClosure, "写关闭")
				log.Println("写不成功数据就关闭了")
				break
			}
			log.Println("一条写数据结束")
		}
	}
}
