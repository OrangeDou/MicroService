package config

import (
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/go-redis/redis"

	"github.com/samuel/go-zookeeper/zk"
)

var (
	Redis       RedisConf
	Etcd        EtcdConf
	SecKill     SecKillConf
	MysqlConfig MysqlConf
	TraceConfig TraceConf
	Zk          ZookeeperConf
)

type ZookeeperConf struct {
	ZkConn        *zk.Conn
	SecProductKey string //商品键
}

type EtcdConf struct {
	EtcdConn          *clientv3.Client //链接
	EtcdSecProductKey string           //商品键
	Host              string
}

type TraceConf struct {
	Host string
	Port string
	Url  string
}

type MysqlConf struct {
	Host string
	Port string
	User string
	Pwd  string
	Db   string
}

// redis配置
type RedisConf struct {
	RedisConn            *redis.Client //链接
	Proxy2layerQueueName string        //队列名称
	Layer2proxyQueueName string        //队列名称
	IdBlackListHash      string        //用户黑名单hash表
	IpBlackListHash      string        //IP黑名单Hash表
	IdBlackListQueue     string        //用户黑名单队列
	IpBlackListQueue     string        //IP黑名单队列
	Host                 string
	Password             string
	Db                   int
}

type SecKillConf struct {
	RedisConf *RedisConf
	EtcdConf  *EtcdConf

	CookieSecretKey string

	ReferWhiteList []string //白名单

	AccessLimitConf AccessLimitConf

	RWBlackLock                  sync.RWMutex
	WriteProxy2LayerGoroutineNum int
	ReadProxy2LayerGoroutineNum  int

	IPBlackMap map[string]bool
	IDBlackMap map[int]bool

	//SecProductInfoMap map[int]*SecProductInfoConf

	AppWriteToHandleGoroutineNum  int
	AppReadFromHandleGoroutineNum int

	CoreReadRedisGoroutineNum  int
	CoreWriteRedisGoroutineNum int
	CoreHandleGoroutineNum     int

	AppWaitResultTimeout int

	CoreWaitResultTimeout int

	MaxRequestWaitTimeout int

	SendToWriteChanTimeout  int //
	SendToHandleChanTimeout int //
	TokenPassWd             string
}

// 访问限制
type AccessLimitConf struct {
	IPSecAccessLimit   int //IP每秒钟访问限制
	UserSecAccessLimit int //用户每秒钟访问限制
	IPMinAccessLimit   int //IP每分钟访问限制
	UserMinAccessLimit int //用户每分钟访问限制
}
