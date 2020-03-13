package sender

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/whoisnian/aMonitor-agent/internal/config"
	"github.com/whoisnian/aMonitor-agent/internal/util"
)

// 与storage建立的websocket连接
var conn *websocket.Conn

// 身份标识符
var token string

// Packet 数据包
type Packet struct {
	Category  string      // 数据类型
	MetaData  interface{} // 元数据
	Timestamp int64       // 时间戳
	Token     string      // 身份标识
}

// CreatePacket 将collector收集到的数据封装为数据包
func CreatePacket(msg interface{}, category string) Packet {
	return Packet{
		Category:  category,
		MetaData:  msg,
		Timestamp: time.Now().Unix(),
		Token:     token,
	}
}

// 发送消息
func send(msg interface{}) {
	var err error
	// []byte和string直接作为文本消息发送，否则创建数据包转换为json后再发送
	switch v := msg.(type) {
	case []byte:
		err = conn.WriteMessage(websocket.TextMessage, v)
	case string:
		err = conn.WriteMessage(websocket.TextMessage, []byte(v))
	case Packet:
		err = conn.WriteJSON(v)
	default:
		err = conn.WriteJSON(CreatePacket(v, util.TypeOf(v)))
	}
	if err != nil {
		log.Panicln(err)
	}
}

// Init 初始化websocket连接
func Init(CONFIG *config.Config) {
	token = CONFIG.Token

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
