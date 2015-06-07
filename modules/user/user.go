// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package user

import (
	"os"
	"os/user"
)

func CurrentUsername() string {
	curUser, err := user.Current()
	if err == nil {
		return curUser.Username
	}

	curUserName := os.Getenv("USER")
	if len(curUserName) > 0 {
		return curUserName
	}

	return os.Getenv("USERNAME")
}
