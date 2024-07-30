// Copyright (c) 2015 Ola Holmström. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package melody

type envelope struct {
	t      int
	msg    []byte
	filter filterFunc
}
