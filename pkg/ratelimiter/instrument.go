package main

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint" //go-kit中的一个RPC抽象，被服务器定义，被客户端调用
	"golang.org/x/time/rate"
)

var ErrLimitExceed = errors.New("Rate limit exceed!")

// 使用x/time/rate创建限流中间件,NewTokenBucketLimitterWithBuildIn方法传入limiter结构体，返回用于生成带限流功能的中间件，该中间件使用endpoint作为参数，为其添加限流功能
func NewTokenBucketLimitterWithBuildIn(bkt *rate.Limiter) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			if !bkt.Allow() {
				return nil, ErrLimitExceed
			}
			return next(ctx, request)
		}
	}
}

// 使用方法如下
// ratebucket := rate.NewLimiter(rate.Every(time.Second*1), 100)
// endpoint = NewTokenBucketLimitterWithBuildIn(ratebucket)(userPoint)
