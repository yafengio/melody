// Copyright 2015 Ola Holmström. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package melody implements a framework for dealing with WebSockets.
//
// Example
//
// A broadcasting echo server:
//
//
// func main() {
// 	m := melody.New()

// 	m.HandleMessage(func(s *melody.Session, msg []byte) {
// 		s.Write(msg)
// 	})

// 	handler := func(ctx *fasthttp.RequestCtx) {
// 		m.HandleRequest(ctx)
// 	}

// 	fasthttp.ListenAndServe(":5000", handler)
// }

package melody
