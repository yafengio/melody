//go:build ignore

package main

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/fasthttp/websocket"
)

func main() {
	conn, _, err := websocket.DefaultDialer.DialContext(context.Background(), "ws://127.0.0.1:15000", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	go func() {
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				log.Fatal(err)
			}

			slog.Info("read message", "mt", mt, "msg", string(msg))
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := 0; i < 3; i++ {
		select {
		case t := <-ticker.C:
			conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
		}
	}
	time.Sleep(time.Second)
}
