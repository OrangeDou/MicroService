package setup

import (
	"log"
	"strconv"
	"time"

	conf "micro/pkg/config"

	"micro/sk-app/service/srv_redis"

	"github.com/go-redis/redis"
)

// 初始化Redis
func InitRedis() {
	log.Printf("init redis %s", conf.Redis.Password)
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Host,
		Password: conf.Redis.Password,
		DB:       conf.Redis.Db,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Printf("Connect redis failed. Error : %v", err)
	}
	log.Printf("init redis success")
	conf.Redis.RedisConn = client

	loadBlackList(client) //加载黑白名单列表
	initRedisProcess()
}

// 加载黑名单列表
func loadBlackList(conn *redis.Client) {
	conf.SecKill.IPBlackMap = make(map[string]bool, 10000)
	conf.SecKill.IDBlackMap = make(map[int]bool, 10000)

	//用户Id
	idList, err := conn.HGetAll(conf.Redis.IdBlackListHash).Result()

	if err != nil {
		log.Printf("hget all failed. Error : %v", err)
		return
	}

	for _, v := range idList {
		id, err := strconv.Atoi(v)
		if err != nil {
			log.Printf("invalid user id [%v]", id)
			continue
		}
		conf.SecKill.IDBlackMap[id] = true
	}

	//用户Ip
	ipList, err := conn.HGetAll(conf.Redis.IpBlackListHash).Result()
	if err != nil {
		log.Printf("hget all failed. Error : %v", err)
		return
	}

	for _, v := range ipList {
		conf.SecKill.IPBlackMap[v] = true
	}

	go syncIpBlackList(conn)
	go syncIdBlackList(conn)
	return
}

// 同步用户ID黑名单
func syncIdBlackList(conn *redis.Client) {
	for {
		idArr, err := conn.BRPop(time.Minute, conf.Redis.IdBlackListQueue).Result()
		if err != nil {
			log.Printf("brpop id failed, err : %v", err)
			continue
		}
		id, _ := strconv.Atoi(idArr[1])
		conf.SecKill.RWBlackLock.Lock()
		{
			conf.SecKill.IDBlackMap[id] = true
		}
		conf.SecKill.RWBlackLock.Unlock()
	}
}

// 同步用户IP黑名单
func syncIpBlackList(conn *redis.Client) {
	var ipList []string
	lastTime := time.Now().Unix()

	for {
		ipArr, err := conn.BRPop(time.Minute, conf.Redis.IpBlackListQueue).Result()
		if err != nil {
			log.Printf("brpop ip failed, err : %v", err)
			continue
		}

		ip := ipArr[1]
		curTime := time.Now().Unix()
		ipList = append(ipList, ip)

		if len(ipList) > 100 || curTime-lastTime > 5 {
			conf.SecKill.RWBlackLock.Lock()
			{
				for _, v := range ipList {
					conf.SecKill.IPBlackMap[v] = true
				}
			}
			conf.SecKill.RWBlackLock.Lock()

			lastTime = curTime
			log.Printf("sync ip list from redis success, ip[%v]", ipList)
		}
	}
}

// 初始化redis进程
func initRedisProcess() {
	log.Printf("initRedisProcess %d %d", conf.SecKill.AppWriteToHandleGoroutineNum, conf.SecKill.AppReadFromHandleGoroutineNum)
	for i := 0; i < conf.SecKill.AppWriteToHandleGoroutineNum; i++ {
		go srv_redis.WriteHandle()
	}

	for i := 0; i < conf.SecKill.AppReadFromHandleGoroutineNum; i++ {
		go srv_redis.ReadHandle()
	}
}
