package common

// 服务实例
type ServiceInstance struct {
	Host          string
	Port          int
	Weight        int //服务权重，固定不变
	CurrentWeight int //当前权重，动态调整（服务降级）

	GrpcPort int
}
