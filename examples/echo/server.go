//go:build ignore

package main

import (
	"bytes"
	"log"
	"log/slog"

	"github.com/valyala/fasthttp"
	"github.com/yafengio/melody"
)

func main() {
	m := melody.New()

	m.HandleConnect(func(s *melody.Session) {
		slog.Info("contect", "session", s)
	})

	m.HandleDisconnect(func(s *melody.Session) {
		slog.Info("discontect", "session", s)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		s.Write(bytes.ToUpper(msg))
	})

	handler := func(ctx *fasthttp.RequestCtx) {
		m.HandleRequest(ctx)
	}

	if err := fasthttp.ListenAndServe(":15000", handler); err != nil {
		log.Fatal(err)
	}
}
