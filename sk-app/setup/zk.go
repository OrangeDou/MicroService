package setup

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	conf "micro/pkg/config"

	"github.com/samuel/go-zookeeper/zk"
)

func InitZk() {
	var hosts = []string{"127.0.0.1:2181"}
	option := zk.WithEventCallback(waitSecProductEvent)
	conn, _, err := zk.Connect(hosts, time.Second*5, option)
	if err != nil {
		fmt.Println(err)
		return
	}

	conf.Zk.ZkConn = conn
	conf.Zk.SecProductKey = "/product"
	go loadSecConf(conn) // 加载秒杀
}

// 加载秒杀商品信息
func loadSecConf(conn *zk.Conn) {
	log.Printf("Connect zk sucess %s", conf.Etcd.EtcdSecProductKey)
	v, _, err := conn.Get(conf.Zk.SecProductKey)
	if err != nil {
		log.Printf("get product info failed, err : %v", err)
		return
	}
	log.Printf("get product info ")
	var secProductInfo []*conf.SecProductInfoConf
	err1 := json.Unmarshal(v, &secProductInfo)
	if err1 != nil {
		log.Printf("Unmsharl second product info failed, err : %v", err1)
	}
	// 更新商品信息
	updateSecProductInfo(secProductInfo)
}

func waitSecProductEvent(event zk.Event) {
	if event.Path == conf.Zk.SecProductKey {

	}
}

// 更新秒杀商品信息
func updateSecProductInfo(secProductInfo []*conf.SecProductInfoConf) {
	tmp := make(map[int]*conf.SecProductInfoConf, 1024)
	for _, v := range secProductInfo {
		tmp[v.ProductId] = v
	}
	conf.SecKill.RWBlackLock.Lock()
	conf.SecKill.SecProductInfoMap = tmp
	conf.SecKill.RWBlackLock.Unlock()

}
