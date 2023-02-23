package feed

import (
	"context"
	"strconv"
	bizConstant "toktik/constant/biz"
	"toktik/constant/config"
	"toktik/kitex_gen/douyin/feed"
	feedService "toktik/kitex_gen/douyin/feed/feedservice"
	"toktik/logging"
	"toktik/service/web/mw"

	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"

	"github.com/cloudwego/hertz/pkg/app"
	httpStatus "github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/sirupsen/logrus"
)

var feedClient feedService.Client

func init() {
	r, err := consul.NewConsulResolver(config.EnvConfig.CONSUL_ADDR)
	if err != nil {
		logging.Logger.WithError(err).Fatal("init feed client failed")
		panic(err)
	}
	provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.FeedServiceName),
		provider.WithExportEndpoint(config.EnvConfig.EXPORT_ENDPOINT),
		provider.WithInsecure(),
	)
	feedClient, err = feedService.NewClient(
		config.FeedServiceName,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
	)
	if err != nil {
		logging.Logger.WithError(err).Fatal("init feed client failed")
		panic(err)
	}
}
func Action(ctx context.Context, c *app.RequestContext) {
	methodFields := logrus.Fields{
		"method": "FeedAction",
	}
	logger := logging.Logger.WithFields(methodFields)
	logger.Debugf("Process start")

	latestTime := c.Query("latest_time")
	if _, err := strconv.Atoi(latestTime); latestTime != "" && err != nil {
		bizConstant.InvalidLatestTime.
			WithCause(err).
			WithFields(&methodFields).
			LaunchError(c)
		return
	}

	var actorIdPtr *uint32
	switch c.GetString(mw.AuthResultKey) {
	case mw.AUTH_RESULT_SUCCESS, mw.AUTH_RESULT_NO_TOKEN:
		*actorIdPtr = c.GetUint32(mw.UserIdKey)
	default:
		bizConstant.UnAuthorized.WithFields(&methodFields).LaunchError(c)
		return
	}

	logger.WithFields(logrus.Fields{
		"latestTime": latestTime,
		"actorId":    actorIdPtr,
	}).Debugf("Executing get feed")
	response, err := feedClient.ListVideos(ctx, &feed.ListFeedRequest{
		LatestTime: &latestTime,
		ActorId:    actorIdPtr,
	})

	if err != nil {
		bizConstant.RPCCallError.
			WithCause(err).
			WithFields(&methodFields).
			LaunchError(c)
		return
	}

	logger.WithFields(logrus.Fields{
		"response": response,
	}).Debugf("Getting feed success")
	c.JSON(httpStatus.StatusOK, response)
}
