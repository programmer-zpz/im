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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"openpitrix.io/logger"

	"cloudbases.io/im/pkg/constants"
	"cloudbases.io/im/pkg/global"
	"cloudbases.io/im/pkg/models"
	"cloudbases.io/im/pkg/pb"
)

func GetUserGroupBindings(ctx context.Context, userIds, groupIds []string) ([]*models.UserGroupBinding, error) {
	var userGroupBindings []*models.UserGroupBinding
	if err := global.Global().Database.Table(constants.TableUserGroupBinding).
		Where(constants.ColumnGroupId+" in (?)", groupIds).
		Where(constants.ColumnUserId+" in (?)", userIds).
		Find(&userGroupBindings).
		Error; err != nil {
		logger.Errorf(ctx, "Get user group binding failed: %+v", err)
		return nil, err
	}

	return userGroupBindings, nil
}

func JoinGroup(ctx context.Context, req *pb.JoinGroupRequest) (*pb.JoinGroupResponse, error) {
	if len(req.UserId) == 0 || len(req.GroupId) == 0 {
		err := status.Errorf(codes.InvalidArgument, "empty user id or group id")
		logger.Errorf(ctx, "%+v", err)
		return nil, err
	}

	// check user in group
	userGroupBindings, err := GetUserGroupBindings(ctx, req.UserId, req.GroupId)
	if err != nil {
		return nil, err
	}
	if len(userGroupBindings) != 0 {
		err := status.Errorf(codes.PermissionDenied, "user already in group")
		logger.Errorf(ctx, "%+v", err)
		return nil, err
	}

	tx := global.Global().Database.Begin()
	{
		for _, groupId := range req.GroupId {
			for _, userId := range req.UserId {
				if err := tx.Create(models.NewUserGroupBinding(userId, groupId)).Error; err != nil {
					tx.Rollback()
					logger.Errorf(ctx, "Insert user group binding failed: %+v", err)
					return nil, err
				}
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		logger.Errorf(ctx, "Batch insert user group binding failed: %+v", err)
		return nil, err
	}

	return &pb.JoinGroupResponse{
		GroupId: req.GroupId,
		UserId:  req.UserId,
	}, nil
}

func LeaveGroup(ctx context.Context, req *pb.LeaveGroupRequest) (*pb.LeaveGroupResponse, error) {
	if len(req.UserId) == 0 || len(req.GroupId) == 0 {
		err := status.Errorf(codes.InvalidArgument, "empty user id or group id")
		logger.Errorf(ctx, "%+v", err)
		return nil, err
	}

	// check user in group
	userGroupBindings, err := GetUserGroupBindings(ctx, req.UserId, req.GroupId)
	if err != nil {
		return nil, err
	}
	if len(userGroupBindings) != len(req.UserId)*len(req.GroupId) {
		err := status.Errorf(codes.PermissionDenied, "user not in group")
		logger.Errorf(ctx, "%+v", err)
		return nil, err
	}

	if err := global.Global().Database.
		Where(constants.ColumnGroupId+" in (?)", req.GroupId).
		Where(constants.ColumnUserId+" in (?)", req.UserId).
		Delete(models.UserGroupBinding{}).Error; err != nil {
		logger.Errorf(ctx, "Delete user group binding failed: %+v", err)
		return nil, err
	}

	return &pb.LeaveGroupResponse{
		GroupId: req.GroupId,
		UserId:  req.UserId,
	}, nil
}

func GetGroupsByUserIds(ctx context.Context, userIds []string) ([]*models.Group, error) {
	var groups []*models.Group
	if err := global.Global().Database.
		Table(constants.TableGroup).
		Select("`group`.*").
		Joins("JOIN `user_group_binding` on `user_group_binding`.user_id in (?) AND `user_group_binding`.group_id=`group`.group_id", userIds).
		Scan(&groups).Error; err != nil {
		logger.Errorf(ctx, "Get groups by user id failed: %+v", err)
		return nil, err
	}

	return groups, nil
}

func GetUsersByGroupIds(ctx context.Context, groupIds []string) ([]*models.User, error) {
	var users []*models.User
	if err := global.Global().Database.
		Table(constants.TableUser).
		Select("`user`.*").
		Joins("JOIN `user_group_binding` on `user_group_binding`.group_id in (?) AND `user_group_binding`.user_id=`user`.user_id", groupIds).
		Scan(&users).Error; err != nil {
		logger.Errorf(ctx, "Get users by group id failed: %+v", err)
		return nil, err
	}

	return users, nil
}

func GetUserIdsByGroupIds(ctx context.Context, groupIds []string) ([]string, error) {
	rows, err := global.Global().Database.Table(constants.TableUserGroupBinding).
		Select(constants.ColumnUserId).
		Where(constants.ColumnGroupId+" in (?)", groupIds).
		Rows()
	if err != nil {
		logger.Errorf(ctx, "Get user ids by group id failed: %+v", err)
		return nil, err
	}
	var userIds []string
	for rows.Next() {
		var userId string
		rows.Scan(&userId)
		userIds = append(userIds, userId)
	}
	return userIds, nil
}
