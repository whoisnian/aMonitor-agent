package sender

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/whoisnian/aMonitor-agent/config"
)

var conn *websocket.Conn

// Init 初始化websocket连接
func Init(CONFIG *config.Config) {
	var err error
	conn, _, err = websocket.DefaultDialer.Dial(CONFIG.StroageURL, nil)
	if err != nil {
		log.Panicln(err)
	}
}

// WaitAndSend 等待channel并发送数据
func WaitAndSend(msgChan <-chan interface{}) {
	defer conn.Close()

	var err error
	for msg := range msgChan {
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
}
