package client

import (
	"context"
	"micro/pb"
	"micro/pkg/discover"
	"micro/pkg/loadbalance"

	"github.com/opentracing/opentracing-go"
)

// 鉴权
type OAuthClient interface {
	CheckToken(ctx context.Context, tracer opentracing.Tracer, request *pb.CheckTokenRequest) (*pb.CheckTokenResponse, error)
}

// 定义客户端管理器、服务名称、负载均衡策略和链路追踪系统
type OAuthClientImpl struct {
	manager     ClientManager
	serviceName string
	loadbalance loadbalance.LoadBalance
	tracer      opentracing.Tracer
}

// OAuthClient工厂方法，初始化impl实例
func NewOAuthClient(serviceName string, lb loadbalance.LoadBalance, tracer opentracing.Tracer) (OAuthClient, error) {
	if serviceName == "" {
		serviceName = "oauth"
	}
	if lb == nil {
		lb = defaultLoadBalance
	}

	return &OAuthClientImpl{
		manager: &DefaultClientManager{
			serviceName:     serviceName,
			loadBalance:     lb,
			discoveryClient: discover.ConsulService,
			logger:          discover.Logger,
		},
		serviceName: serviceName,
		loadbalance: lb,
		tracer:      tracer,
	}, nil
}

// checkToken方法，使用改RPC客户端的业务服务可以直接初始化Impl实例，调用内部的CheckToken方法，就像在调用本地方法，但是方法内部却实现了RPC调用
func (impl *OAuthClientImpl) CheckToken(ctx context.Context, tracer opentracing.Tracer, request *pb.CheckTokenRequest) (*pb.CheckTokenResponse, error) {
	response := new(pb.CheckTokenResponse)
	if err := impl.manager.DecoratorInvoke("/pb.OAuthService/CheckToken", "token_check", tracer, ctx, request, response); err == nil {
		return response, nil
	} else {
		return nil, err
	}
}
