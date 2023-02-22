package relation

import (
	"context"
	"strconv"
	bizConstant "toktik/constant/biz"
	bizConfig "toktik/constant/config"
	"toktik/kitex_gen/douyin/relation"
	relationService "toktik/kitex_gen/douyin/relation/relationservice"
	"toktik/logging"
	"toktik/service/web/mw"

	"github.com/cloudwego/hertz/pkg/app"
	httpStatus "github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/sirupsen/logrus"
)

var relationClient relationService.Client

func init() {
	r, err := consul.NewConsulResolver(bizConfig.EnvConfig.CONSUL_ADDR)
	if err != nil {
		panic(err)
	}
	relationClient, err = relationService.NewClient(bizConfig.RelationServiceName, client.WithResolver(r))
	if err != nil {
		panic(err)
	}
}

// RelationAction [POST] /relation/action/
func RelationAction(ctx context.Context, c *app.RequestContext) {
	var actionTypeInt int
	var relationResp *relation.RelationActionResponse

	methodFields := logrus.Fields{
		"method": "RelationAction",
	}
	logger := logging.Logger.WithFields(methodFields)
	logger.Debugf("Process start")
	userId := mw.GetAuthActorId(c)

	actionType, exist := c.GetQuery("action_type")
	if !exist {
		bizConstant.InvalidActionType.WithFields(&methodFields).LaunchError(c)
		return
	}
	actionTypeInt, err := strconv.Atoi(actionType)
	if err != nil || (actionTypeInt != 1 && actionTypeInt != 2) {
		err2Launch := bizConstant.InvalidActionType.WithFields(&methodFields)
		if err != nil {
			err2Launch = err2Launch.WithCause(err)
		}
		err2Launch.LaunchError(c)
		return
	}

	targetId, exist := c.GetQuery("to_user_id")
	if !exist {
		bizConstant.InvalidArguments.WithFields(&methodFields).LaunchError(c)
		return
	}
	targetIdInt, err := strconv.ParseInt(targetId, 10, 32)
	if err != nil {
		bizConstant.InvalidArguments.WithCause(err).WithFields(&methodFields).LaunchError(c)
		return
	}
	switch actionTypeInt {
	case 1:
		logger.WithFields(logrus.Fields{
			"userId":   userId,
			"toUserId": targetId,
		}).Debugf("Executing follow")
		relationResp, err = relationClient.Follow(ctx, &relation.RelationActionRequest{
			UserId:   userId,
			ToUserId: uint32(targetIdInt),
		})
	case 2:
		logger.WithFields(logrus.Fields{
			"userId":   userId,
			"toUserId": targetId,
		}).Debugf("Executing unfollow")
		relationResp, err = relationClient.Unfollow(ctx, &relation.RelationActionRequest{
			UserId:   userId,
			ToUserId: uint32(targetIdInt),
		})
	default:
		bizConstant.InvalidArguments.WithFields(&methodFields).LaunchError(c)
		return
	}

	if err != nil {
		bizConstant.RPCCallError.WithCause(err).WithFields(&methodFields).LaunchError(c)
		return
	}
	logger.WithFields(logrus.Fields{
		"response": relationResp,
	}).Debugf("Relation action success")
	c.JSON(
		httpStatus.StatusOK,
		relationResp,
	)
}

// FollowList [POST] /relation/follow/list/
func GetFollowList(ctx context.Context, c *app.RequestContext) {

	methodFields := logrus.Fields{
		"method": "GetFollowList",
	}
	logger := logging.Logger.WithFields(methodFields)
	logger.Debugf("Process start")
	userId := mw.GetAuthActorId(c)

	logger.WithFields(logrus.Fields{
		"userId": userId,
	}).Debugf("Executing get follow list")
	followListResp, err := relationClient.GetFollowList(ctx, &relation.FollowListRequest{
		UserId: userId,
	})

	if err != nil {
		bizConstant.RPCCallError.WithCause(err).WithFields(&methodFields).LaunchError(c)
		return
	}
	logger.WithFields(logrus.Fields{
		"response": followListResp,
	}).Debugf("get follow list success")
	c.JSON(
		httpStatus.StatusOK,
		followListResp,
	)
}

// FollowList [POST] /relation/follower/list/
func GetFollowerList(ctx context.Context, c *app.RequestContext) {

	methodFields := logrus.Fields{
		"method": "GetFollowerList",
	}
	logger := logging.Logger.WithFields(methodFields)
	logger.Debugf("Process start")
	userId := mw.GetAuthActorId(c)

	logger.WithFields(logrus.Fields{
		"userId": userId,
	}).Debugf("Executing get follower list")
	followerListResp, err := relationClient.GetFollowerList(ctx, &relation.FollowerListRequest{
		UserId: userId,
	})

	if err != nil {
		bizConstant.RPCCallError.WithCause(err).WithFields(&methodFields).LaunchError(c)
		return
	}
	logger.WithFields(logrus.Fields{
		"response": followerListResp,
	}).Debugf("get follower list success")
	c.JSON(
		httpStatus.StatusOK,
		followerListResp,
	)
}

// GetFriendList [POST] /relation/friends/list/
func GetFriendList(ctx context.Context, c *app.RequestContext) {

	methodFields := logrus.Fields{
		"method": "GetFriendList",
	}
	logger := logging.Logger.WithFields(methodFields)
	logger.Debugf("Process start")
	userId := mw.GetAuthActorId(c)

	logger.WithFields(logrus.Fields{
		"userId": userId,
	}).Debugf("Executing get friend list")
	friendListResp, err := relationClient.GetFriendList(ctx, &relation.FriendListRequest{
		UserId: userId,
	})

	if err != nil {
		bizConstant.RPCCallError.WithCause(err).WithFields(&methodFields).LaunchError(c)
		return
	}
	logger.WithFields(logrus.Fields{
		"response": friendListResp,
	}).Debugf("get friends list success")
	c.JSON(
		httpStatus.StatusOK,
		friendListResp,
	)
}
