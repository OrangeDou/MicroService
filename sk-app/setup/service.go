package setup

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"micro/sk-app/service"
	"micro/sk-app/transport"

	"micro/sk-app/endpoint"
	"micro/sk-app/plugins"

	localconfig "micro/pkg/config"

	register "micro/pkg/discover"

	kitzipkin "github.com/go-kit/kit/tracing/zipkin"
	"golang.org/x/time/rate"
)

func InitServer(host string, servicePort string) {
	log.Printf("port is ", servicePort)

	flag.Parse()

	errChan := make(chan error)
	ratebucket := rate.NewLimiter(rate.Every(time.Second*1), 5000)

	var (
		skAppService service.Service
	)
	skAppService = service.SkAppService{}
	healthCheckEnd := endpoint.MakeHealthCheckEndpoint(skAppService)
	healthCheckEnd = plugins.NewTokenBucketLimitterWithBuildIn(ratebucket)(healthCheckEnd)
	healthCheckEnd = kitzipkin.TraceEndpoint(localconfig.ZipkinTracer, "heath-check")(healthCheckEnd)

	GetSecInfoEnd := endpoint.MakeSecInfoEndpoint(skAppService)
	GetSecInfoEnd = plugins.NewTokenBucketLimitterWithBuildIn(ratebucket)(GetSecInfoEnd)
	GetSecInfoEnd = kitzipkin.TraceEndpoint(localconfig.ZipkinTracer, "sec-info")(GetSecInfoEnd)

	GetSecInfoListEnd := endpoint.MakeSecInfoListEndpoint(skAppService)
	GetSecInfoListEnd = plugins.NewTokenBucketLimitterWithBuildIn(ratebucket)(GetSecInfoListEnd)
	GetSecInfoListEnd = kitzipkin.TraceEndpoint(localconfig.ZipkinTracer, "sec-info-list")(GetSecInfoListEnd)
	/**
	 * 秒杀接口单独限流
	 */
	secRatebucket := rate.NewLimiter(rate.Every(time.Microsecond*100), 1000)

	SecKillEnd := endpoint.MakeSecKillEndpoint(skAppService)
	SecKillEnd = plugins.NewTokenBucketLimitterWithBuildIn(secRatebucket)(SecKillEnd)
	//SecKillEnd = kitzipkin.TraceEndpoint(localconfig.ZipkinTracer, "sec-kill")(SecKillEnd)

	testEnd := endpoint.MakeTestEndpoint(skAppService)
	testEnd = kitzipkin.TraceEndpoint(localconfig.ZipkinTracer, "test")(testEnd)

	endpts := endpoint.SkAppEndpoints{
		SecKillEndpoint:        SecKillEnd,
		HeathCheckEndpoint:     healthCheckEnd,
		GetSecInfoEndpoint:     GetSecInfoEnd,
		GetSecInfoListEndpoint: GetSecInfoListEnd,
		TestEndpoint:           testEnd,
	}
	ctx := context.Background()
	//创建http.Handler
	r := transport.MakeHttpHandler(ctx, endpts, localconfig.ZipkinTracer, localconfig.Logger)

	//http server
	go func() {
		fmt.Println("Http Server start at port:" + servicePort)
		//启动前执行注册
		register.Register()
		handler := r
		errChan <- http.ListenAndServe(":"+servicePort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	error := <-errChan
	//服务退出取消注册
	register.Deregister()
	fmt.Println(error)
}
