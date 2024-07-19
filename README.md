# Micro_service
## 秒杀系统的设计与实现
### 目录结构
- pkg 基础组件
-    bootstrap: 初始化配置信息
### 实现步骤
#### 1. 服务注册与发现
  使用Consul作为服务注册发现的组件，秒杀微服务的各个核心业务组件会注册到Consul并查询要通信的服务实例信息。
- pkg\common\service_instance.go：定义服务实例相关属性
- pkg\discover\discovery_client.go：定义服务注册与发现客户端，并声明注册、注销、发现方法
- pkg\discover\kit_consul_client.go: 定义注册、注销、发现方法
#### 2. 负载均衡策略
  抵抗较大的并发流量，结合服务注册与发现，可以对服务实例的数量动态的增加和减少，来应对不同的并发流量。
使用service_instance来构建负载均衡体系。本项目使用带权重的平滑轮询策略实现负载均衡，每个实例具有一个默认的权重值以及当前动态权重，具体实现在：
- pkg\loadbalance\loadbalance.go
#### 3. RPC客户端装饰器
  微服务架构中，各个服务之间需要通过RPC相互交互，在之前的1和2小节已经实现了各个服务注册以及发现、负载均衡方法，在开始实现各个服务实例前，需要定义他们之间相互通信的RPC客户端，用于统一封装业务服务提供的RPC接口客户端。
- pkg\client\oauth.go：鉴权服务RPC接口（CheckToken）
- pkg\client\decorator.go：RPC接口的装饰器方法，作用是对RPC接口中传入的RPC路径进行RPC请求，并进行熔断保护、链路追踪、服务发现、负载均衡。真正实现RPC调用的是在这个方法中，而CheckToken效果就相当于本地方法。
#### 4. Zookeeper集成
ZooKeeper 是一个分布式协调服务，主要用于解决分布式系统中的数据管理问题，如统一配置管理、命名服务、分布式同步、组服务等。本项目的秒杀活动和秒杀商品信息会存储在ZooKeeper中，使用watcher机制实时更新信息。