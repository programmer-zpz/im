// Copyright 2019 The OpenPitrix Authors. All rights reserved.
// Use of this source code is governed by a Apache license
// that can be found in the LICENSE file.

package im

import (
	"cloudbases.io/im/pkg/global"
	"cloudbases.io/im/pkg/manager"
	"cloudbases.io/im/pkg/pb"
)

type Client struct {
	pb.IdentityManagerClient
}

func NewClient() (*Client, error) {
	conn, err := manager.NewClient(global.Global().Config.Host, global.Global().Config.Port)
	if err != nil {
		return nil, err
	}

	return &Client{
		IdentityManagerClient: pb.NewIdentityManagerClient(conn),
	}, nil
}
