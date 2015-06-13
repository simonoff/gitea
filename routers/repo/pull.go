// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"github.com/go-gitea/gitea/modules/base"
	"github.com/go-gitea/gitea/modules/middleware"
)

const (
	PULLS base.TplName = "repo/pull/list"
	PULL base.TplName = "repo/pull/pull"
)

func Pulls(ctx *middleware.Context) {
	issues(ctx, PULLS, true)
}

func Pull(ctx *middleware.Context) {
	ctx.Data["IsRepoToolbarPulls"] = true
	ctx.HTML(200, PULL)
}
