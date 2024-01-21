package models

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
)

var (
	redisDB *redis.Client
)

func OpenRedis() {
	redisDB = redis.NewClient(&redis.Options{
		Addr:     "192.168.1.5:6379",
		Password: "", // no password set
		DB:       5,  // use default DB
	})
}

/*基础表缓存 uuid：tablePrimaryKeyId*/

func VideoRedis(videoUUID string) int64 {
	val, err := redisDB.Get(context.Background(), videoUUID).Result()
	if err != nil {
		return 0
	}
	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return num
}
func AuthorRedis(authorUUID string) int64 {
	val, err := redisDB.Get(context.Background(), authorUUID).Result()
	if err != nil {
		return 0
	}
	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return num
}
func TagRedis(tagName string) int64 {
	val, err := redisDB.Get(context.Background(), tagName).Result()
	if err != nil {
		return 0
	}
	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return num
}
func CollectRedis(collectBvid int64) int64 {
	val, err := redisDB.Get(context.Background(), strconv.FormatInt(collectBvid, 10)).Result()
	if err != nil {
		return 0
	}
	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

/*中间表缓存 */

func VideoAuthorRedis(videoPrimaryKey int64) int64 {
	key := fmt.Sprintf("videoAuthor-%d", videoPrimaryKey)
	val, err := redisDB.Get(context.Background(), key).Result()
	if err != nil {
		return -1
	}
	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return -1
	}
	return num
}
func AuthorVideoRedis(authorPrimaryKey int64) int64 {
	key := fmt.Sprintf("authorVideo-%d", authorPrimaryKey)
	val, err := redisDB.Get(context.Background(), key).Result()
	if err != nil {
		return -1
	}
	num, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return -1
	}
	return num
}

/*插入数据*/
func initBaseRedis() {
	var (
		tableCount  int64
		offsetIndex int
		videoList   []Video
		authorList  []Author
		tagList     []Tag
		collectList []Collect
	)
	cb := context.Background()
	GormDB.Table("video").Count(&tableCount)
	if tableCount == 0 {
		println("视频表总数查询失败")
		return
	}
	offsetIndex = 0
	for {
		videoList = make([]Video, 10000)
		GormDB.Table("video").Offset(offsetIndex * 10000).Limit(10000).Find(&videoList)
		for _, video := range videoList {
			redisDB.Set(cb, video.Uuid, video.Id, 0)
		}
		if int64(offsetIndex*10000) > tableCount {
			break
		}
		offsetIndex++
	}
	GormDB.Table("author").Count(&tableCount)
	if tableCount == 0 {
		println("作者表总数查询失败")
		return
	}
	offsetIndex = 0
	for {
		authorList = make([]Author, 10000)
		GormDB.Table("author").Offset(offsetIndex * 10000).Limit(10000).Find(&authorList)
		for _, author := range authorList {
			redisDB.Set(cb, author.AuthorWebUid, author.Id, 0)
		}
		if int64(offsetIndex*10000) > tableCount {
			break
		}
		offsetIndex++
	}
	GormDB.Table("tag").Count(&tableCount)
	if tableCount == 0 {
		println("标签表总数查询失败")
		return
	}
	offsetIndex = 0
	for {
		tagList = make([]Tag, 10000)
		GormDB.Table("tag").Offset(offsetIndex * 10000).Limit(10000).Find(&tagList)
		for _, tag := range tagList {
			redisDB.Set(cb, tag.Name, tag.Id, 0)
		}
		if int64(offsetIndex*10000) > tableCount {
			break
		}
		offsetIndex++
	}
	GormDB.Table("collect").Count(&tableCount)
	if tableCount == 0 {
		println("收藏表总数查询失败")
		return
	}
	offsetIndex = 0
	for {
		collectList = make([]Collect, 10000)
		GormDB.Table("collect").Offset(offsetIndex * 10000).Limit(10000).Find(&collectList)
		for _, collect := range collectList {
			redisDB.Set(cb, strconv.FormatInt(collect.BvId, 10), collect.Id, 0)
		}
		if int64(offsetIndex*10000) > tableCount {
			break
		}
		println("redis缓存完成")
		return
	}
}

func CreateVideoToRedis(videoUUID string, videoId int64) {
	if redisDB == nil {
		return
	}
	redisDB.Set(context.Background(), videoUUID, videoId, 0)
}

func CreateVideoAuthorToRedis(AuthorWebUid string, authorId int64) {
	if redisDB == nil {
		return
	}
	redisDB.Set(context.Background(), AuthorWebUid, authorId, 0)
}
