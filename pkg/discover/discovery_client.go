package discover

import (
	"log"
	"micro/pkg/common"
	"sync"

	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
)

// 服务注册与发现客户端，封装注册、注销、发现方法
type DiscoveryClient interface {
	/**
	* 服务注册接口
	* @param serviceName 服务名
	* @param instanceId 服务实例Id
	* @param instancePort 服务实例端口
	* @param healthCheckUrl 健康检查地址
	* @param meta 服务实例元数据
	* @param weight 权重
	 */

	Register(instanceId, svcHost, healthCheckUrl, svcPort string, svcName string, weight int, meta map[string]string, tags []string, logger *log.Logger) bool

	/**
	 * 服务注销接口
	 * @param instanceId 服务实例Id
	 */
	DeRegister(instanceId string, logger *log.Logger) bool

	/**
	 * 发现服务实例接口
	 * @param serviceName 服务名
	 */
	DiscoverServices(serviceName string, logger *log.Logger) []*common.ServiceInstance
}

type DiscoveryClientInstance struct {
	Host string
	Port int
	// 连接 consul 的配置
	config *api.Config
	client consul.Client
	mutex  sync.Mutex
	// 服务实例缓存字段
	instancesMap sync.Map
}
