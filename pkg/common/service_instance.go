package common

// 服务实例
type ServiceInstance struct {
	Host          string
	Port          int
	Weight        int //服务权重
	CurrentWeight int

	GrpcPort int
}
