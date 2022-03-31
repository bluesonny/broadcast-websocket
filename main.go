package main

import (
	. "broadcast-websocket/handler"
	"log"
	"net/http"
)

func main() {

	log.Println("推送服务开始...")
	manager := NewClientManager()
	go manager.Start()
	go manager.RedisSend()
	http.HandleFunc("/ws", manager.WsHandle)
	http.ListenAndServe(":9090", nil)
}
