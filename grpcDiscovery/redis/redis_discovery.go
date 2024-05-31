package redisDiscovery

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
	"sync"
)

const (
	grpcServerType        = "grpcServerType"
	grpcServerTypeSubName = "grpcServerTypeSub"
)

var (
	redisDB           *redis.Client
	invalidGrpcClient map[string]int
	grpcServerMap     map[string][]string
	lock              sync.Mutex
)

func OpenRedis() {
	grpcServerMap = make(map[string][]string)
	redisDB = redis.NewClient(&redis.Options{
		Addr:     "database:6379",
		Password: "", // no password set
		DB:       5,  // use default DB
	})
	getWebSiteServer()
	go subChain()
}

type WebSiteServer struct {
	WebSiteName   string
	ServerUrlList []string
}

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

func getWebSiteName() []string {
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

func getWebSiteServer() {
	lock.Lock()
	defer lock.Unlock()
	webSiteList := getWebSiteName()

	for _, webSiteName := range webSiteList {
		serverUrlList := redisDB.SMembers(context.Background(), webSiteName).Val()
		fmt.Printf("%s\n", serverUrlList)
		for index, serverUrl := range serverUrlList {
			serverUrlList[index] = strings.Replace(serverUrl, "192.168.0.20", "database", 1)
		}
		fmt.Printf("%s\n", serverUrlList)
		grpcServerMap[webSiteName] = serverUrlList
	}
	return
}

func GetWebSiteServer() []WebSiteServer {
	lock.Lock()
	defer lock.Unlock()
	result := make([]WebSiteServer, 0)
	for webSiteName, serverUrlList := range grpcServerMap {
		result = append(result, WebSiteServer{
			WebSiteName:   webSiteName,
			ServerUrlList: serverUrlList,
		})
	}
	return result
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

func subChain() {
	grpcServerTypeSub := redisDB.Subscribe(context.Background(), grpcServerTypeSubName)
	defer grpcServerTypeSub.Close()
	ch := grpcServerTypeSub.Channel()
	for msg := range ch {
		println(msg.Payload)
		lock.Lock()
		infoList := strings.Split(msg.Payload, "-")
		serverType := infoList[0]
		serverName := infoList[1]
		serverUrl := infoList[2]
		if serverType == "register" {
			_, ok := grpcServerMap[serverName]
			if !ok {
				grpcServerMap[serverName] = make([]string, 0)
			}
			grpcServerMap[serverName] = append(grpcServerMap[serverName], serverUrl)
		} else if serverType == "unregister" {
			_, ok := grpcServerMap[serverName]
			if ok {
				for index, url := range grpcServerMap[serverName] {
					if url == serverUrl {
						grpcServerMap[serverName] = append(grpcServerMap[serverName][:index], grpcServerMap[serverName][index+1:]...)
						break
					}
				}
			}
		}
		lock.Unlock()
	}
}
