package transport

import (
	"context"
	"encoding/json"
	"errors"
	"micro/sk-app/model"
	"net/http"

	"github.com/go-kit/kit/log"

	endpts "micro/sk-app/endpoint"

	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	gozipkin "github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ErrorBadRequest = errors.New("invalid request parameter")
)

func MakeHttpHandler(ctx context.Context, endpoints endpts.SkAppEndpoints, zipkinTracer *gozipkin.Tracer, logger log.Logger) http.Handler {

	// mux构建高性能http路由器
	r := mux.NewRouter()
	// 链路追踪server
	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer, zipkin.Name("http-transport"))

	options := []kithttp.ServerOption{
		//kithttp.ServerErrorLogger(logger),
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		//kithttp.ServerErrorEncoder(kithttp.DefaultErrorEncoder),
		kithttp.ServerErrorEncoder(encodeError),
		zipkinServer,
	}

	// 定义接口
	r.Methods("GET").Path("/sec/info").Handler(kithttp.NewServer(
		// 将请求分配给endpoint处理，可以理解为handler
		endpoints.GetSecInfoEndpoint,
		// 对请求进行解码
		decodeSecInfoRequest,
		// 对响应进行编码
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/sec/list").Handler(kithttp.NewServer(
		endpoints.GetSecInfoListEndpoint,
		decodeSecInfoListRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Path("/sec/kill").Handler(kithttp.NewServer(
		endpoints.SecKillEndpoint,
		decodeSecKillRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/sec/test").Handler(kithttp.NewServer(
		endpoints.TestEndpoint,
		decodeSecInfoListRequest,
		encodeResponse,
		options...,
	))

	r.Path("/metrics").Handler(promhttp.Handler())

	// create health check handler
	r.Methods("GET").Path("/health").Handler(kithttp.NewServer(
		endpoints.HeathCheckEndpoint,
		decodeTestRequest,
		encodeResponse,
		options...,
	))

	return r
}

// 将秒杀信息请求解码成结构体
func decodeSecInfoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var secInfoRequest endpts.SecInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&secInfoRequest); err != nil {
		return nil, err
	}
	return secInfoRequest, nil

}

// 编码响应并返回
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func decodeHealthCheckRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return endpts.HealthRequest{}, nil
}

func decodeTestRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return endpts.HealthRequest{}, nil
}

func decodeSecInfoListRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeSecKillRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var secRequest model.SecRequest
	if err := json.NewDecoder(r.Body).Decode(&secRequest); err != nil {
		return nil, err
	}
	return secRequest, nil
}
