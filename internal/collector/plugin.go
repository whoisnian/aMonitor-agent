package collector

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/whoisnian/aMonitor-agent/internal/sender"
)

// StartPluginListen 开始监听插件信息上报接口
func StartPluginListen(ctx context.Context, wg *sync.WaitGroup, msgChan chan interface{}) {
	defer wg.Done()

	// 处理请求
	muxHandler := http.NewServeMux()
	muxHandler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			category := r.FormValue("category")
			if category != "" {
				var data interface{}
				dec := json.NewDecoder(r.Body)
				dec.Decode(&data)
				msgChan <- sender.CreatePacket(data, category)
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// 初始化http server
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           muxHandler,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 放入goroutine中进行监听
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panicln(err)
		}
	}()

	// 等待主程序通知goroutine结束
	<-ctx.Done()
	log.Println("Plugin collector receive ctx.Done.")

	// 通知http server结束
	ctxShutdown, canael := context.WithTimeout(context.Background(), 10*time.Second)
	defer canael()
	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Panicln(err)
	}

	log.Println("Close plugin collector.")
}
