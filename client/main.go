package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const addr = ":8000"

var recentAction string
var mu sync.Mutex

type Message struct {
	Action *string `json:"action"`
	Type   string  `json:"type"`
	Data   Data    `json:"data"`
}

type Data struct {
	IP        *string `json:"ip,omitempty"`
	LastVisit *string `json:"last_visit,omitempty"`
	Name      *string `json:"name,omitempty"`
	ID        *int    `json:"id,omitempty"`
}

func main() {
	srv := http.Server{
		Addr:    addr,
		Handler: nil,
	}

	go func() {
		err := startFetchData()
		if err != nil {
			log.Println(err)
			srv.Shutdown(context.Background())
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		w.Write([]byte(recentAction))
		mu.Unlock()
	})
	log.Println("Server started")
	log.Println(srv.ListenAndServe())
}

const (
	pongWait   = 5 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

func startFetchData() error {
	conn, _, err := websocket.DefaultDialer.Dial("ws://agg-server:5000", nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	log.Println("connected to ws://agg-server:5000")
	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return err
			}
			log.Println("ping message send")

		default:
			var message Message
			err := conn.ReadJSON(&message)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				return err
			}

			if message.Action != nil {
				mu.Lock()
				recentAction = *message.Action
				mu.Unlock()
			}
		}
	}
}
