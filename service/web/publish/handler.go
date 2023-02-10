package publish

import (
	"context"
	"log"
	"toktik/constant/config"
	"toktik/kitex_gen/douyin/publish"
	publishService "toktik/kitex_gen/douyin/publish/publishservice"
	"toktik/service/web/mw"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
)

var publishClient publishService.Client

func init() {
	r, err := consul.NewConsulResolver(config.EnvConfig.CONSUL_ADDR)
	if err != nil {
		log.Fatal(err)
	}
	publishClient, err = publishService.NewClient(config.PublishServiceName, client.WithResolver(r))
	if err != nil {
		log.Fatal(err)
	}
}

func Action(ctx context.Context, c *app.RequestContext) {
	userId := c.GetUint32(mw.USER_ID_KEY)
	publishResp, err := publishClient.CreateVideo(ctx, &publish.CreateVideoRequest{
		UserId: uint32(userId),
	})
	if err != nil {
		c.JSON(
			consts.StatusOK,
			&publish.CreateVideoResponse{
				StatusCode: 1,
				StatusMsg:  err.Error(),
			},
		)
		return
	}
	c.JSON(
		consts.StatusOK,
		publishResp,
	)
}
