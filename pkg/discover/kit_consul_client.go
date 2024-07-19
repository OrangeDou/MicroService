package discover

import (
	"fmt"
	"log"
	"micro/pkg/common"
	"strconv"

	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

// 服务注册方法
func (consulClient *DiscoveryClientInstance) Register(instanceId, svcHost, healthCheckUrl, svcPort string, svcName string, weight int, meta map[string]string, tags []string, logger *log.Logger) bool {
	port, _ := strconv.Atoi(svcPort)

	//1. 构建服务实例元数据结构体
	fmt.Println(weight)
	serviceRegistration := &api.AgentServiceRegistration{
		ID:      instanceId,
		Name:    svcName,
		Address: svcHost,
		Port:    port,
		Meta:    meta,
		Tags:    tags,
		Weights: &api.AgentWeights{
			Passing: weight,
		},
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: "30s",
			HTTP:                           "http://" + svcHost + ":" + strconv.Itoa(port) + healthCheckUrl,
			Interval:                       "15s",
		},
	}

	//2. 发送服务注册到consul中
	err := consulClient.client.Register(serviceRegistration)

	if err != nil {
		if logger != nil {
			logger.Println("Register Service Error!")
		}
		return false
	}
	if logger != nil {
		logger.Println("Register Service Success!")
	}
	return true
}

// 服务注销方法
func (consulClient *DiscoveryClientInstance) DeRegister(instanceId string, logger *log.Logger) bool {

	// 构建包含服务实例 ID 的元数据结构体
	serviceRegistration := &api.AgentServiceRegistration{
		ID: instanceId,
	}
	// 发送服务注销请求
	err := consulClient.client.Deregister(serviceRegistration)

	if err != nil {
		if logger != nil {
			logger.Println("Deregister Service Error!")
		}
		return false
	}
	if logger != nil {
		logger.Println("Deregister Service Success!")
	}

	return true
}

// 服务查询方法：查询服务实例,根据服务名称获取服务中心的服务实例列表，并缓存在Map中，同时注册对该服务实例的监控，当有新服务上线或者旧服务下线时，对服务实例的监控就可发现，更更新本地缓存的服务实例列表
func (consulClient *DiscoveryClientInstance) DiscoverServices(serviceName string, logger *log.Logger) []*common.ServiceInstance {
	//  该服务已监控并缓存
	instanceList, ok := consulClient.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]*common.ServiceInstance)
	}
	//申请锁
	consulClient.mutex.Lock()
	// 再次检查是否监控
	instanceList, ok = consulClient.instancesMap.Load(serviceName)
	if ok {
		return instanceList.([]*common.ServiceInstance)
	} else {
		//并发注册对当前服务的监控,当serviceName对应的服务状态发生变化，watch的Handler就会执行
		go func() {
			// 监控map
			params := make(map[string]interface{})
			params["type"] = "service"
			params["service"] = serviceName
			// 进行监控，调用consul的watch方法
			plan, _ := watch.Parse(params)
			// 依赖于索引来决定何时触发 watch 事件的回调函数
			plan.Handler = func(u uint64, i interface{}) {
				if i == nil {
					return
				}
				v, ok := i.([]*api.ServiceEntry) //断言，服务实例在线列表
				if !ok {
					return //数据异常 忽略
				}
				// 没有服务实例在线，将当前的serviceName保存在缓存中
				if len(v) == 0 {
					consulClient.instancesMap.Store(serviceName, []*common.ServiceInstance{})
				}

				var healthServices []*common.ServiceInstance

				for _, service := range v {
					// AggregatedStatus返回服务实例的四个健康状态：maintenance > critical > warning > passing
					if service.Checks.AggregatedStatus() == api.HealthPassing {
						// 健康则加入健康列表
						healthServices = append(healthServices, newServiceInstance(service.Service))
					}
				} // 对当前服务实例健康状态检查完毕

				// 将当前服务加入健康列表
				consulClient.instancesMap.Store(serviceName, healthServices)
			}
			defer plan.Stop() // 停止watch
			plan.Run(consulClient.config.Address)
		}()
	}
	defer consulClient.mutex.Unlock() //解锁

	// 根据serviceName查询服务实例列表
	entries, _, err := consulClient.client.Service(serviceName, "", false, nil)
	// 如果出现错误，将 serviceName 对应的服务实例列表设置为空，存储在 consulClient.instancesMap中。
	if err != nil {
		consulClient.instancesMap.Store(serviceName, []*common.ServiceInstance{})
		if logger != nil {
			logger.Println("Discover Service Error!")
		}
		return nil //返回空列表
	}
	instances := make([]*common.ServiceInstance, len(entries))
	for i := 0; i < len(instances); i++ {
		instances[i] = newServiceInstance(entries[i].Service)
	}
	consulClient.instancesMap.Store(serviceName, instances)
	return instances
}

// 初始化服务实例
func newServiceInstance(service *api.AgentService) *common.ServiceInstance {
	rpcPort := service.Port - 1
	if service.Meta != nil {
		if rpcPortString, ok := service.Meta["rpcPort"]; ok {
			rpcPort, _ = strconv.Atoi(rpcPortString)
		}
	}
	return &common.ServiceInstance{
		Host:     service.Address,
		Port:     service.Port,
		GrpcPort: rpcPort,
		Weight:   service.Weights.Passing,
	}
}
func New(consulHost string, consulPort string) *DiscoveryClientInstance {
	port, _ := strconv.Atoi(consulPort)
	// 通过 Consul Host 和 Consul Port 创建一个 consul.Client
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulHost + ":" + strconv.Itoa(port)
	apiClient, err := api.NewClient(consulConfig)
	if err != nil {
		return nil
	}

	client := consul.NewClient(apiClient)

	return &DiscoveryClientInstance{
		Host:   consulHost,
		Port:   port,
		config: consulConfig,
		client: client,
	}
}
