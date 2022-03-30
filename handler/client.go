package handler

import (
	"context"
	"fmt"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
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
		c.socket.Close(websocket.StatusInternalError, "读关闭")
		fmt.Printf("defer 读关闭 %v\n", c)
	}()

	for {
		//_, _, err := c.socket.ReadMessage()
		err := wsjson.Read(ctx, c.socket, "")
		fmt.Printf("是在不停的读吗？ %v\n", c)
		if err != nil {
			fmt.Println("连接断开了...")
			break
		}

	}
}

//写入管道后激活这个进程
func (c *Client) Write(ctx context.Context, manager clientManager) {
	defer func() {
		manager.unregister <- c
		//c.socket.Close()
		c.socket.Close(websocket.StatusInternalError, "写关闭")
		fmt.Printf("defer 写关闭了 %v\n", c)
	}()

	for {
		select {
		case message, ok := <-c.send: //这个管道有了数据 写这个消息出去
			fmt.Printf("读通道状态 %v\n", ok)
			fmt.Printf("读通道状态连接 %v\n", c)
			if !ok {
				return
			}

			//err := c.socket.WriteMessage(websocket.TextMessage, message)
			//var msg Message
			//json.Unmarshal(message, &msg)
			err := wsjson.Write(ctx, c.socket, message)
			fmt.Printf("发送数据状态 %v\n", err)
			fmt.Printf("发送数据状态连接 %v\n", c)
			if err != nil {
				manager.unregister <- c
				//c.socket.Close()
				c.socket.Close(websocket.StatusInternalError, "写关闭")
				fmt.Println("写不成功数据就关闭了")
				break
			}
			fmt.Println("一条写数据结束")
		}
	}
}
