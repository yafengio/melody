// Copyright (c) 2015 Ola Holmström. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package melody

import "errors"

var (
	ErrClosed            = errors.New("melody instance is closed")
	ErrSessionClosed     = errors.New("session is closed")
	ErrWriteClosed       = errors.New("tried to write to closed a session")
	ErrMessageBufferFull = errors.New("session message buffer is full")
)
