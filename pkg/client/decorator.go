package client

import (
	"context"
	"errors"
	"log"
	"micro/pkg/discover"
	"micro/pkg/loadbalance"
	"strconv"
	"time"

	"micro/pkg/bootstrap"
	conf "micro/pkg/config"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	zipkin "github.com/lksunshine/zipkin-go-opentracing"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrRPCService = errors.New("no rpc service")
)

// 定义默认负载均衡方法
var defaultLoadBalance loadbalance.LoadBalance = &loadbalance.WeightRoundRobinLoadBalance{}

type ClientManager interface {
	/*
		1. 调用ClientManager的before回调函数，进行发送RPC前的统一回调处理。
		2. 使用Hystrix的Do方法构造断路器进行保护。
		3. 调用DiscoveryClient的DiscoverServices方法获得服务提供方的服务实例列表。
		4. 调用负载均衡策略从服务实例列表中选取合适的服务实例
		5. 向选取的服务实例发送对应的RPC请求
		6. 调用ClientManager的after回调函数
		**/
	DecoratorInvoke(path string, hystrixName string,
		tracer opentracing.Tracer, ctx context.Context, inputVal interface{}, outVal interface{}) (err error)
}

type DefaultClientManager struct {
	serviceName     string
	logger          *log.Logger
	discoveryClient discover.DiscoveryClient
	loadBalance     loadbalance.LoadBalance
	after           []InvokerAfterFunc
	before          []InvokerBeforeFunc
}

type InvokerAfterFunc func() (err error)

type InvokerBeforeFunc func() (err error)

func (manager *DefaultClientManager) DecoratorInvoke(path string, hystrixName string,
	tracer opentracing.Tracer, ctx context.Context, inputVal interface{}, outVal interface{}) (err error) {
	// 遍历before切片中的所有函数进行执行
	for _, fn := range manager.before {
		if err = fn(); err != nil {
			return err
		}
	}

	if err = hystrix.Do(hystrixName, func() error {
		//熔断保护实现
		//调用服务发现客户端方法，获取服务实例列表
		instances := manager.discoveryClient.DiscoverServices(manager.serviceName, manager.logger)
		//调用负载均衡方法，选取服务实例
		if instance, err := manager.loadBalance.SelectService(instances); err == nil {
			//rpc端口大于0
			if instance.GrpcPort > 0 {
				//获取RPC端口并发送RPC请求
				if conn, err := grpc.NewClient(instance.Host+":"+strconv.Itoa(instance.GrpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()),
					grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(genTracer(tracer), otgrpc.LogPayloads())), grpc.WithTimeout(1*time.Second)); err == nil {
					// 发送RPC请求
					if err = conn.Invoke(ctx, path, inputVal, outVal); err != nil {
						return err
					}
				} else {
					return err
				}

			} else {
				return ErrRPCService
			}
		} else {
			return err
		}
		return nil
	}, func(e error) error {
		return e
	}); err != nil {
		return err
	} else {
		for _, fn := range manager.after {
			if err = fn(); err != nil {
				return err
			}
		}
		return nil
	}

}

func genTracer(tracer opentracing.Tracer) opentracing.Tracer {
	if tracer != nil {
		return tracer
	}
	zipkinUrl := "http://" + conf.TraceConfig.Host + ":" + conf.TraceConfig.Port + conf.TraceConfig.Url
	zipkinRecorder := bootstrap.HttpConfig.Host + ":" + bootstrap.HttpConfig.Port
	collector, err := zipkin.NewHTTPCollector(zipkinUrl)
	if err != nil {
		log.Fatalf("zipkin.NewHTTPCollector err: %v", err)
	}

	recorder := zipkin.NewRecorder(collector, false, zipkinRecorder, bootstrap.DiscoverConfig.ServiceName)

	res, err := zipkin.NewTracer(
		recorder, zipkin.ClientServerSameSpan(true),
	)
	if err != nil {
		log.Fatalf("zipkin.NewTracer err: %v", err)
	}
	return res

}
