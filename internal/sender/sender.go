package sender

import (
	"context"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/whoisnian/aMonitor-agent/internal/config"
)

var conn *websocket.Conn

func send(msg interface{}) {
	var err error
	// []byte和string直接作为文本消息发送，否则转换为json后再发送
	switch v := msg.(type) {
	case []byte:
		err = conn.WriteMessage(websocket.TextMessage, v)
	case string:
		err = conn.WriteMessage(websocket.TextMessage, []byte(v))
	default:
		err = conn.WriteJSON(v)
	}
	if err != nil {
		log.Panicln(err)
	}
}

// Init 初始化websocket连接
func Init(CONFIG *config.Config) {
	var err error
	conn, _, err = websocket.DefaultDialer.Dial(CONFIG.StroageURL, nil)
	if err != nil {
		log.Panicln(err)
	}
}

// WaitAndSend 等待channel并发送数据
func WaitAndSend(ctx context.Context, wg *sync.WaitGroup, msgChan <-chan interface{}) {
	defer wg.Done()
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			// 接收到结束信号时发送完channel中的数据再退出
			log.Println("Sender receive ctx.Done.")
			for {
				select {
				case msg := <-msgChan:
					send(msg)
				default:
					log.Println("Close Sender.")
					return
				}
			}
		case msg := <-msgChan:
			send(msg)
		}
	}
}
