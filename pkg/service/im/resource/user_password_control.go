/*
Copyright 2019 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resource

import (
	"context"
	"crypto/md5"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"openpitrix.io/logger"

	"cloudbases.io/im/pkg/constants"
	"cloudbases.io/im/pkg/global"
	"cloudbases.io/im/pkg/models"
	"cloudbases.io/im/pkg/pb"
)

func ComparePassword(ctx context.Context, req *pb.ComparePasswordRequest) (*pb.ComparePasswordResponse, error) {
	var user = &models.User{UserId: req.UserId}
	if err := global.Global().Database.Table(constants.TableUser).
		Take(user).Error; err != nil {
		logger.Errorf(ctx, "Get user [%s] failed: %+v", req.UserId, err)
		return nil, err
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password), []byte(req.GetPassword()),
	)
	if err != nil {
		logger.Errorf(ctx, "Compare password failed, md5(password): %x", md5.Sum([]byte(req.Password)))
		return &pb.ComparePasswordResponse{Ok: false}, nil
	}

	return &pb.ComparePasswordResponse{Ok: true}, nil
}

func ModifyPassword(ctx context.Context, req *pb.ModifyPasswordRequest) (*pb.ModifyPasswordResponse, error) {
	if req.Password == "" {
		err := status.Errorf(codes.InvalidArgument, "empty password")
		logger.Errorf(ctx, "%+v", err)
		return nil, err
	}

	attributes := map[string]interface{}{
		constants.ColumnPassword:   models.GetBcryptPassword(req.Password),
		constants.ColumnUpdateTime: time.Now(),
	}

	if err := global.Global().Database.Table(constants.TableUser).
		Where(constants.ColumnUserId+" = ?", req.UserId).
		Updates(attributes).Error; err != nil {
		logger.Errorf(ctx, "Modify user [%s] password failed: %+v", req.UserId, err)
		return nil, err
	}

	return &pb.ModifyPasswordResponse{UserId: req.UserId}, nil
}
