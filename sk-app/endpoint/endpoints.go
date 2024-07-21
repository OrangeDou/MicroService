package endpoint

import (
	"context"
	"micro/sk-app/model"

	service "micro/sk-app/service"

	"github.com/go-kit/kit/endpoint"
)

// 定义5个endpoint
type SkAppEndpoints struct {
	SecKillEndpoint        endpoint.Endpoint // 秒杀接口
	HeathCheckEndpoint     endpoint.Endpoint //健康检查接口
	GetSecInfoEndpoint     endpoint.Endpoint //获取秒杀信息接口
	GetSecInfoListEndpoint endpoint.Endpoint //获取秒杀信息列表接口
	TestEndpoint           endpoint.Endpoint //测试接口
}

func (ue SkAppEndpoints) HealthCheck() bool {
	return false
}

type SecInfoRequest struct {
	productId int `json:"id"`
}

type Response struct {
	Result map[string]interface{} `json:"result"`
	Error  error                  `json:"error"`
	Code   int                    `json:"code"`
}

type SecInfoListResponse struct {
	Result []map[string]interface{} `json:"result"`
	Num    int                      `json:"num"`
	Error  error                    `json:"error"`
}

// 处理秒杀信息相关请求，对应service层的SecInfoList方法
func MakeSecInfoEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		result, num, error := svc.SecInfoList() //调用service层核心逻辑
		return SecInfoListResponse{result, num, error}, nil
	}
}

// make endpoint
func MakeSecKillEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(model.SecRequest)
		ret, code, calError := svc.SecKill(&req)
		return Response{Result: ret, Code: code, Error: calError}, nil
	}
}

func MakeTestEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return Response{Result: nil, Code: 1, Error: nil}, nil
	}
}

// HealthRequest 健康检查请求结构
type HealthRequest struct{}

// HealthResponse 健康检查响应结构
type HealthResponse struct {
	Status bool `json:"status"`
}

// MakeHealthCheckEndpoint 创建健康检查Endpoint
func MakeHealthCheckEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		status := svc.HealthCheck()
		return HealthResponse{status}, nil
	}
}
func MakeSecInfoListEndpoint(svc service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		ret, num, error := svc.SecInfoList()
		return SecInfoListResponse{ret, num, error}, nil
	}
}
