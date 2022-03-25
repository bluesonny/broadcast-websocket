package main

import (
	. "broadcast-websocket/handler"
	"fmt"
	"net/http"
)

func main() {

	fmt.Println("推送服务开始...")
	manager := NewClientManager()
	go manager.Start()
	go manager.Send()
	http.HandleFunc("/ws", manager.WsHandle)
	http.ListenAndServe(":9090", nil)
}
