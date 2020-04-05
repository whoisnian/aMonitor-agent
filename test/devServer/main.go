package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(messageType, string(p))
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct{ MachineID string }
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	res := struct{ Token string }{req.MachineID}

	content, _ := json.Marshal(res)
	w.Write(content)

	log.Println("SendToken: " + req.MachineID)
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/register", registerHandler)

	err := http.ListenAndServe("127.0.0.1:3000", nil)
	if err != nil {
		log.Panicln(err)
	}
}
