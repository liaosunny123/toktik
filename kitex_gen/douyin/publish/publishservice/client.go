// Code generated by Kitex v0.4.4. DO NOT EDIT.

package publishservice

import (
	"context"
	client "github.com/cloudwego/kitex/client"
	callopt "github.com/cloudwego/kitex/client/callopt"
	publish "toktik/kitex_gen/douyin/publish"
)

// Client is designed to provide IDL-compatible methods with call-option parameter for kitex framework.
type Client interface {
	CreateVideo(ctx context.Context, Req *publish.CreateVideoRequest, callOptions ...callopt.Option) (r *publish.CreateVideoResponse, err error)
	ListVideo(ctx context.Context, Req *publish.ListVideoRequest, callOptions ...callopt.Option) (r *publish.ListVideoResponse, err error)
	CountVideo(ctx context.Context, Req *publish.CountVideoRequest, callOptions ...callopt.Option) (r *publish.CountVideoResponse, err error)
}

// NewClient creates a client for the service defined in IDL.
func NewClient(destService string, opts ...client.Option) (Client, error) {
	var options []client.Option
	options = append(options, client.WithDestService(destService))

	options = append(options, opts...)

	kc, err := client.NewClient(serviceInfo(), options...)
	if err != nil {
		return nil, err
	}
	return &kPublishServiceClient{
		kClient: newServiceClient(kc),
	}, nil
}

// MustNewClient creates a client for the service defined in IDL. It panics if any error occurs.
func MustNewClient(destService string, opts ...client.Option) Client {
	kc, err := NewClient(destService, opts...)
	if err != nil {
		panic(err)
	}
	return kc
}

type kPublishServiceClient struct {
	*kClient
}

func (p *kPublishServiceClient) CreateVideo(ctx context.Context, Req *publish.CreateVideoRequest, callOptions ...callopt.Option) (r *publish.CreateVideoResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.CreateVideo(ctx, Req)
}

func (p *kPublishServiceClient) ListVideo(ctx context.Context, Req *publish.ListVideoRequest, callOptions ...callopt.Option) (r *publish.ListVideoResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.ListVideo(ctx, Req)
}

func (p *kPublishServiceClient) CountVideo(ctx context.Context, Req *publish.CountVideoRequest, callOptions ...callopt.Option) (r *publish.CountVideoResponse, err error) {
	ctx = client.NewCtxWithCallOptions(ctx, callOptions)
	return p.kClient.CountVideo(ctx, Req)
}
