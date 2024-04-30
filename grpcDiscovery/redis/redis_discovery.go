package redisDiscovery

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
)

var (
	redisDB           *redis.Client
	invalidGrpcClient map[string]int
)

func OpenRedis() {
	redisDB = redis.NewClient(&redis.Options{
		Addr:     "database:6379",
		Password: "", // no password set
		DB:       5,  // use default DB
	})
}

type WebSiteServer struct {
	WebSiteName   string
	ServerUrlList []string
}

const (
	grpcServerType = "grpcServerType"
)

func RegisterWebSite(webSiteServer WebSiteServer) error {
	if webSiteServer.WebSiteName == "" {
		return errors.New("webSiteName不能为空")
	}
	if webSiteServer.ServerUrlList == nil || len(webSiteServer.ServerUrlList) == 0 {
		return errors.New("serverUrlList不能为空")
	}
	if redisDB == nil {
		return errors.New("redis尚未初始化")
	}
	redisDB.SAdd(context.Background(), grpcServerType, webSiteServer.WebSiteName)
	redisDB.SAdd(context.Background(), webSiteServer.WebSiteName, webSiteServer.ServerUrlList[0])
	return nil
}

func GetWebSiteName() []string {
	s := redisDB.SScan(context.Background(), grpcServerType, 0, "*", 0)
	result, _, err := s.Result()
	if err != nil {
		return nil
	}

	return result
}
func GetSpecifyServer(webSiteName string) []string {
	s := redisDB.SScan(context.Background(), webSiteName, 0, "*", 0)
	result, _, err := s.Result()
	if err != nil {
		return nil
	}
	return result
}

func GetWebSiteServer() []WebSiteServer {
	webSiteList := GetWebSiteName()

	var webSiteServerList []WebSiteServer
	for _, webSiteName := range webSiteList {
		serverUrlList := redisDB.SMembers(context.Background(), webSiteName).Val()
		webSiteServerList = append(webSiteServerList, WebSiteServer{
			WebSiteName:   webSiteName,
			ServerUrlList: serverUrlList,
		})
	}
	return webSiteServerList
}
func InvalidGrpcClient(webSiteName string) {
	if invalidGrpcClient == nil {
		invalidGrpcClient = make(map[string]int)
	}
	_, ok := invalidGrpcClient[webSiteName]
	if ok {
		invalidGrpcClient[webSiteName] = 1
	}
	invalidGrpcClient[webSiteName]++
	if invalidGrpcClient[webSiteName] > 10 {
		delete(invalidGrpcClient, webSiteName)
	}
}
