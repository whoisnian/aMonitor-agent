package sender

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/whoisnian/aMonitor-agent/internal/config"
	"github.com/whoisnian/aMonitor-agent/internal/util"
)

// 与storage建立的websocket连接
var conn *websocket.Conn

// websocket连接健康检测
var pongWait, lastPong int64 = 0, 0

// Packet 数据包
type Packet struct {
	Category  string      // 数据类型
	MetaData  interface{} // 元数据
	Timestamp int64       // 时间戳
}

// CreatePacket 将collector收集到的数据封装为数据包
func CreatePacket(msg interface{}, category string) Packet {
	return Packet{
		Category:  category,
		MetaData:  msg,
		Timestamp: time.Now().Unix(),
	}
}

// 健康检测
func healthCheck() {
	conn.SetPongHandler(func(data string) error {
		lastPong = time.Now().Unix()
		return nil
	})
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Panicln(err)
			break
		}
	}
}

// 发送消息
func send(msg interface{}) {
	if lastPong == 0 {
		lastPong = time.Now().Unix()
	} else if time.Now().Unix()-lastPong > pongWait {
		log.Panicln("Pong response timeout.")
	}

	var err error
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
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
	header := http.Header{"Authorization": {"Bearer " + CONFIG.Token}}

	var err error
	conn, _, err = websocket.DefaultDialer.Dial("ws://"+CONFIG.StroageAddr+"/ws", header)
	if err != nil {
		log.Panicln(err)
	}

	pongWait = CONFIG.Interval.CPU
	pongWait = util.Min(pongWait, CONFIG.Interval.MEM)
	pongWait = util.Min(pongWait, CONFIG.Interval.LOAD)
	pongWait = util.Min(pongWait, CONFIG.Interval.NET)
	pongWait = util.Min(pongWait, CONFIG.Interval.MOUNTS)
	pongWait = util.Min(pongWait, CONFIG.Interval.DISK)
	pongWait = pongWait * 2
	go healthCheck()
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
			if len(msgChan) > 0 {
				log.Println("Left", len(msgChan), "msg in mshChan.")
			}
		}
	}
}
