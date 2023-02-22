package main

import (
	"context"
	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/sirupsen/logrus"
	"log"
	"strconv"
	"time"
	"toktik/constant/biz"
	"toktik/constant/config"
	"toktik/kitex_gen/douyin/comment"
	"toktik/kitex_gen/douyin/comment/commentservice"
	"toktik/kitex_gen/douyin/favorite"
	favoriteService "toktik/kitex_gen/douyin/favorite/favoriteservice"
	"toktik/kitex_gen/douyin/feed"
	"toktik/kitex_gen/douyin/user"
	"toktik/kitex_gen/douyin/user/userservice"
	"toktik/logging"
	gen "toktik/repo"
	"toktik/repo/model"
	"toktik/storage"
)

var UserClient userservice.Client
var CommentClient commentservice.Client
var FavoriteClient favoriteService.Client

func init() {
	r, err := consul.NewConsulResolver(config.EnvConfig.CONSUL_ADDR)
	if err != nil {
		log.Fatal(err)
	}
	UserClient, err = userservice.NewClient(config.UserServiceName, client.WithResolver(r))
	if err != nil {
		log.Fatal(err)
	}
	CommentClient, err = commentservice.NewClient(config.CommentServiceName, client.WithResolver(r))
	if err != nil {
		log.Fatal(err)
	}
	FavoriteClient, err = favoriteService.NewClient(config.FavoriteServiceName, client.WithResolver(r))
	if err != nil {
		log.Fatal(err)
	}
}

// FeedServiceImpl implements the last service interface defined in the IDL.
type FeedServiceImpl struct{}

// ListVideos implements the FeedServiceImpl interface.
func (s *FeedServiceImpl) ListVideos(ctx context.Context, req *feed.ListFeedRequest) (resp *feed.ListFeedResponse, err error) {
	methodFields := logrus.Fields{
		"latest_time": req.LatestTime,
		"actor_id":    req.ActorId,
		"function":    "ListVideos",
	}
	logger := logging.Logger.WithFields(methodFields)
	logger.Debug("Process start")

	now := time.Now().UnixMilli()
	latestTime, err := strconv.ParseInt(*req.LatestTime, 10, 64)
	if err != nil {
		if _, ok := err.(*strconv.NumError); ok {
			latestTime = now
		} else {
			resp = &feed.ListFeedResponse{
				StatusCode: biz.Unable2ParseLatestTimeStatusCode,
				StatusMsg:  &biz.BadRequestStatusMsg,
				NextTime:   nil,
				VideoList:  nil,
			}
			return resp, nil
		}
	}

	find, err := findVideos(ctx, latestTime)
	if err != nil {
		resp = &feed.ListFeedResponse{
			StatusCode: biz.SQLQueryErrorStatusCode,
			StatusMsg:  &biz.InternalServerErrorStatusMsg,
			NextTime:   &now,
			VideoList:  nil,
		}
		return resp, nil
	}

	if len(find) == 0 {
		resp = &feed.ListFeedResponse{
			StatusCode: biz.OkStatusCode,
			StatusMsg:  &biz.OkStatusMsg,
			NextTime:   nil,
			VideoList:  nil,
		}
		return resp, nil
	}
	nextTime := find[len(find)-1].CreatedAt.Add(time.Duration(-1)).UnixMilli()

	var actorId uint32 = 0
	if req.ActorId != nil {
		actorId = *req.ActorId
	}
	videos, err := queryDetailed(ctx, logger, actorId, find)

	return &feed.ListFeedResponse{
		StatusCode: biz.OkStatusCode,
		StatusMsg:  &biz.OkStatusMsg,
		NextTime:   &nextTime,
		VideoList:  videos,
	}, nil
}

func findVideos(ctx context.Context, latestTime int64) ([]*model.Video, error) {
	video := gen.Q.Video
	return video.WithContext(ctx).
		Where(video.CreatedAt.Lte(time.UnixMilli(latestTime))).
		Order(video.CreatedAt.Desc()).
		Limit(biz.VideoCount).
		Offset(0).
		Find()
}

// QueryVideos implements the FeedServiceImpl interface.
func (s *FeedServiceImpl) QueryVideos(ctx context.Context, req *feed.QueryVideosRequest) (resp *feed.QueryVideosResponse, err error) {
	methodFields := logrus.Fields{
		"actor_id":  req.ActorId,
		"video_ids": req.VideoIds,
		"function":  "ListVideos",
	}
	logger := logging.Logger.WithFields(methodFields)
	logger.Debug("Process start")

	rst, err := query(ctx, logger, req.ActorId, req.VideoIds)
	if err != nil {
		resp = &feed.QueryVideosResponse{
			StatusCode: biz.UnableToQueryUser,
			StatusMsg:  &biz.InternalServerErrorStatusMsg,
			VideoList:  rst,
		}
		return resp, nil
	}

	return &feed.QueryVideosResponse{
		StatusCode: biz.OkStatusCode,
		StatusMsg:  &biz.OkStatusMsg,
		VideoList:  rst,
	}, nil
}

func queryDetailed(ctx context.Context, logger *logrus.Entry, actorId uint32, videos []*model.Video) (resp []*feed.Video, err error) {
	rVideos := make([]*feed.Video, 0, len(videos))

	for _, m := range videos {

		userResponse, err := UserClient.GetUser(ctx, &user.UserRequest{
			UserId:  m.UserId,
			ActorId: actorId,
		})
		if err != nil || userResponse.StatusCode != biz.OkStatusCode {
			logger.WithFields(logrus.Fields{
				"err": err,
			}).Debug("failed to get user info")
			continue
		}

		playUrl, err := storage.GetLink(m.FileName)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"err": err,
			}).Debug("failed to fetch play url")
			continue
		}

		coverUrl, err := storage.GetLink(m.CoverName)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"err": err,
			}).Debug("failed to fetch cover url")
			continue
		}

		favoriteCount, err := FavoriteClient.FavoriteCount(ctx, &favorite.FavoriteCountRequest{
			VideoId: m.ID,
		})
		if err != nil {
			logger.WithFields(logrus.Fields{
				"err": err,
			}).Debug("failed to fetch favorite count")
			continue
		}
		commentCount, err := CommentClient.CountComment(ctx, &comment.CountCommentRequest{
			ActorId: actorId,
			VideoId: m.ID,
		})
		if err != nil {
			logger.WithFields(logrus.Fields{
				"err": err,
			}).Debug("failed to fetch comment count")
			continue
		}
		isFavorite, err := FavoriteClient.IsFavorite(ctx, &favorite.IsFavoriteRequest{
			UserId:  actorId,
			VideoId: m.ID,
		})
		if err != nil {
			logger.WithFields(logrus.Fields{
				"err": err,
			}).Debug("unable to determine if the user liked the video")
			continue
		}

		rVideos = append(rVideos, &feed.Video{
			Id:            m.ID,
			Author:        userResponse.User,
			PlayUrl:       playUrl,
			CoverUrl:      coverUrl,
			FavoriteCount: favoriteCount.Count,
			CommentCount:  commentCount.CommentCount,
			IsFavorite:    isFavorite.Result,
			Title:         m.Title,
		})
	}

	return rVideos, nil
}

func query(ctx context.Context, logger *logrus.Entry, actorId uint32, videoIds []uint32) (resp []*feed.Video, err error) {
	find, err := gen.Q.Video.WithContext(ctx).Where(gen.Q.Video.ID.In(videoIds...)).Find()
	if err != nil {
		return nil, err
	}

	return queryDetailed(ctx, logger, actorId, find)
}
