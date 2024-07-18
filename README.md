# Micro_service
## 秒杀系统的设计与实现
### 目录结构

### 实现步骤
#### 1. 服务注册与发现
使用Consul作为服务注册发现的组件，秒杀微服务的各个核心业务组件会注册到Consul并查询要通信的服务实例信息。
- pkg\common\service_instance.go：定义服务实例相关属性
- pkg\discover\discovery_client.go：定义服务注册与发现客户端，并声明注册、注销、发现方法
- pkg\discover\kit_consul_client.go: 定义注册、注销、发现方法
#### 2. 负载均衡策略
抵抗较大的并发流量，结合服务注册与发现，可以对服务实例的数量动态的增加和减少，来应对不同的并发流量。
使用service_instance来构建负载均衡体系。本项目使用带权重的平滑轮询策略实现负载均衡，每个实例具有一个默认的权重值以及当前动态权重，具体实现在：pkg\loadbalance\loadbalance.go