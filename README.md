# Melody

This project is a fork of the latest version of [olahol/melody](https://github.com/olahol/melody) that continues its development independently.

> :notes: Minimalist websocket framework for Fasthttp.

Melody is websocket framework based on [github.com/fasthttp/websocket](https://github.com/fasthttp/websocket)
that abstracts away the tedious parts of handling websockets. It gets out of
your way so you can write real-time apps. Features include:

* [x] Clear and easy interface similar to `net/http` or Gin.
* [x] A simple way to broadcast to all or selected connected sessions.
* [x] Message buffers making concurrent writing safe.
* [x] Automatic handling of sending ping/pong heartbeats that timeout broken sessions.
* [x] Store data on sessions.

## Install

```bash
go get github.com/maratrixx/melody
```


```go
package main

import (
	"net/http"

	"github.com/maratrixx/melody"
)

func main() {
	m := melody.New()

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		s.Write(msg)
	})

	handler := func(ctx *fasthttp.RequestCtx) {
		m.HandleRequest(ctx)
	}

	fasthttp.ListenAndServe(":5000", handler)
}
```

## Tests

Here are the test results:
```bash
$ go test -timeout 100s github.com/maratrixx/melody -failfast -v -count=1 -cover

output:

=== RUN   TestEcho
--- PASS: TestEcho (0.82s)
=== RUN   TestEchoBinary
--- PASS: TestEchoBinary (0.81s)
=== RUN   TestWriteClosedServer
--- PASS: TestWriteClosedServer (0.54s)
=== RUN   TestWriteClosedClient
--- PASS: TestWriteClosedClient (0.54s)
=== RUN   TestUpgrader
--- PASS: TestUpgrader (0.52s)
=== RUN   TestBroadcast
--- PASS: TestBroadcast (4.03s)
=== RUN   TestClose
--- PASS: TestClose (0.58s)
=== RUN   TestLen
--- PASS: TestLen (0.66s)
=== RUN   TestSessions
--- PASS: TestSessions (0.71s)
=== RUN   TestPingPong
--- PASS: TestPingPong (0.53s)
=== RUN   TestHandleClose
--- PASS: TestHandleClose (0.53s)
=== RUN   TestHandleError
--- PASS: TestHandleError (0.53s)
=== RUN   TestErrClosed
--- PASS: TestErrClosed (0.53s)
=== RUN   TestErrSessionClosed
--- PASS: TestErrSessionClosed (0.52s)
=== RUN   TestErrMessageBufferFull
--- PASS: TestErrMessageBufferFull (0.53s)
=== RUN   TestSessionKeys
--- PASS: TestSessionKeys (0.83s)
=== RUN   TestSessionKeysConcurrent
--- PASS: TestSessionKeysConcurrent (0.52s)
=== RUN   TestMisc
--- PASS: TestMisc (0.52s)
=== RUN   TestHandleSentMessage
--- PASS: TestHandleSentMessage (1.05s)
=== RUN   TestConcurrentMessageHandling
=== RUN   TestConcurrentMessageHandling/text_should_error
=== RUN   TestConcurrentMessageHandling/text_should_not_error
=== RUN   TestConcurrentMessageHandling/binary_should_error
=== RUN   TestConcurrentMessageHandling/binary_should_not_error
--- PASS: TestConcurrentMessageHandling (2.94s)
    --- PASS: TestConcurrentMessageHandling/text_should_error (0.92s)
    --- PASS: TestConcurrentMessageHandling/text_should_not_error (0.56s)
    --- PASS: TestConcurrentMessageHandling/binary_should_error (0.93s)
    --- PASS: TestConcurrentMessageHandling/binary_should_not_error (0.53s)
PASS
coverage: 98.1% of statements
ok  	github.com/maratrixx/melody	18.486s	coverage: 98.1% of statements
```