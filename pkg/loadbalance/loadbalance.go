package loadbalance

import (
	"errors"
	"micro/pkg/common"
)

// 负载均衡策略接口
type LoadBalance interface {
	SelectService(service []*common.ServiceInstance) (*common.ServiceInstance, error)
}

type WeightRoundRobinLoadBalance struct {
}

// 权重平滑负载均衡
func (loadBalance *WeightRoundRobinLoadBalance) SelectService(services []*common.ServiceInstance) (best *common.ServiceInstance, err error) {

	if len(services) == 0 {
		return nil, errors.New("service instances are not exist")
	}

	total := 0
	// 轮询每个服务实例
	for i := 0; i < len(services); i++ {
		w := services[i]
		if w == nil {
			continue
		}
		//将当前服务实例的权重加到它的当前权重上。这样做是为了实现权重轮询，即权重高的实例被选中的概率更大。
		w.CurrentWeight += w.Weight
		total += w.Weight
		if best == nil || w.CurrentWeight > best.CurrentWeight {
			best = w //将当前服务实例设置为最佳服务实例，一轮循环以后会选出初始权重最大的服务实例设置为best
		}
	}
	if best == nil {
		return nil, nil
	}
	//在下一次轮询时重置权重
	best.CurrentWeight -= total
	return best, nil
}
