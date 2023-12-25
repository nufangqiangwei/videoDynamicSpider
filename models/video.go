package models

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"strconv"
	"time"
	"videoDynamicAcquisition/utils"
)

// Video 视频信息
type Video struct {
	Id                int64 `json:"id" gorm:"primary_key"`
	WebSiteId         int64
	Authors           []VideoAuthor  `gorm:"foreignKey:VideoId;references:Id"`
	Tag               []VideoTag     `gorm:"foreignKey:VideoId;references:Id"`
	CollectList       []CollectVideo `gorm:"foreignKey:VideoId;references:Id"`
	ViewHistory       []VideoHistory `gorm:"foreignKey:VideoId;references:Id"`
	Title             string         `gorm:"size:255"`
	VideoDesc         string         `gorm:"size:2000"`
	Duration          int
	Uuid              string `gorm:"size:255"`
	Url               string `gorm:"size:255"`
	CoverUrl          string `gorm:"size:255"`
	UploadTime        *time.Time
	CreateTime        time.Time `gorm:"default:CURRENT_TIMESTAMP(3)"`
	View              int64     `gorm:"default:0"`          // 播放数
	Danmaku           int64     `gorm:"default:0"`          // 弹幕数
	Reply             int64     `gorm:"default:0"`          // 评论数
	Favorite          int64     `gorm:"default:0"`          // 收藏数
	Coin              int64     `gorm:"default:0"`          // 硬币数
	Share             int64     `gorm:"default:0"`          // 分享数
	NowRank           int64     `gorm:"default:0"`          // 当前排名
	HisRank           int64     `gorm:"default:0"`          // 历史最高排名
	Like              int64     `gorm:"default:0"`          // 点赞数
	Dislike           int64     `gorm:"default:0"`          // 点踩数
	Evaluation        string    `gorm:"default:0;size:255"` // 综合评分
	StructAuthor      []Author  `gorm:"-"`
	StructTag         []Tag     `gorm:"-"`
	StructCollectList []Collect `gorm:"-"`
}

func (v *Video) GetByUid(uid string) {
	tx := GormDB.Where("uuid = ?", uid).First(v)
	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		utils.ErrorLog.Println("获取视频错误: ")
		utils.ErrorLog.Println(tx.Error.Error())
	}
}

// UpdateVideo 数据储存到数据库中，如果存在则更新，不存在则插入。many to many的表,如果中间表不存在就插入，存在就更新。这里只做增量更新，不做删除。删除操作，有别的同步方法自行执行。
func (v *Video) UpdateVideo() error {
	if v.WebSiteId == 0 || v.Uuid == "" {
		return errors.New("缺少必要参数")
	}
	DBvideo := Video{}
	videoId := VideoRedis(v.Uuid)
	var err error
	if videoId == 0 {
		var (
			authorList  []VideoAuthor
			tagList     []VideoTag
			collectList []CollectVideo
			historyList []VideoHistory
		)
		authorList = v.Authors
		tagList = v.Tag
		historyList = v.ViewHistory
		collectList = v.CollectList

		v.Authors = nil
		v.Tag = nil
		v.ViewHistory = nil
		v.CollectList = nil

		err = GormDB.Create(&v).Error
		if err != nil {
			return err
		}
		redisDB.Set(context.Background(), v.Uuid, v.Id, 0)
		v.Authors = authorList
		v.Tag = tagList
		v.ViewHistory = historyList
		v.CollectList = collectList

		for _, author := range v.Authors {
			author.VideoId = v.Id
			author.AuthorId = AuthorRedis(author.AuthorUUID)
		}
		for _, tag := range v.Tag {
			tag.VideoId = v.Id

		}
		DBvideo.Id = v.Id
	} else {
		// 查出外键关联的数据
		GormDB.Preload("Authors").Preload("Tag").First(&DBvideo, videoId)
		err = GormDB.Model(&DBvideo).Updates(map[string]interface{}{
			"View":       v.View,
			"Danmaku":    v.Danmaku,
			"Reply":      v.Reply,
			"Favorite":   v.Favorite,
			"Coin":       v.Coin,
			"Share":      v.Share,
			"NowRank":    v.NowRank,
			"HisRank":    v.HisRank,
			"Like":       v.Like,
			"Dislike":    v.Dislike,
			"Evaluation": v.Evaluation,
		}).Error
	}
	var (
		saveAuthors []VideoAuthor
		saveTags    []VideoTag
		lateTag     Tag
		saveHistory []VideoHistory
	)
	// 保存作者信息。video.Authors与v.Authors对比，如果v.Authors中有video.Authors中没有的，则插入。这里只做增量更新，不做删除。删除操作，有别的同步方法自行执行。
	// 排除已存在的作者信息
	for _, a := range v.StructAuthor {
		if AuthorRedis(a.AuthorWebUid) == 0 {
			a.WebSiteId = v.WebSiteId
			a.CreateTime = time.Now()
			err = GormDB.Create(&a).Error
			if err != nil {
				return err
			}
			redisDB.Set(context.Background(), a.AuthorWebUid, a.Id, 0)
		}
	}
	for _, author := range v.Authors {
		if author.AuthorId == 0 {
			author.AuthorId = AuthorRedis(author.AuthorUUID)
		}
	}
	saveAuthors = utils.ArrayDifferenceByStruct(v.Authors, DBvideo.Authors, func(a VideoAuthor) string {
		return a.AuthorUUID
	})
	if len(saveAuthors) > 0 {
		GormDB.Save(&saveAuthors)
	}
	// 保存标签信息
	for _, videoTag := range v.StructTag {
		if TagRedis(videoTag.Name) == 0 {
			GormDB.Exec("insert into tag (id,name) select max(id)+1,? from tag", videoTag.Name)
			GormDB.Where("name=?", videoTag.Name).Find(&lateTag)
			redisDB.Set(context.Background(), videoTag.Name, videoTag.Id, 0)
		}
		v.Tag = append(v.Tag, VideoTag{
			TagId:   TagRedis(videoTag.Name),
			VideoId: v.Id,
		})
	}
	saveTags = utils.ArrayDifferenceByStruct(v.Tag, DBvideo.Tag, func(a VideoTag) string {
		return strconv.FormatInt(a.TagId, 10)
	})
	if len(saveTags) > 0 {
		GormDB.Save(&saveTags)
	}
	// 保存收藏信息,先查出StructCollectList所有的信息，然后在插入CollectVideo
	// 取出v.ViewHistory中的数据，添加video_id后保存
	for _, history := range v.ViewHistory {
		history.VideoId = DBvideo.Id
		saveHistory = append(saveHistory, history)
	}
	if len(saveHistory) > 0 {
		GormDB.Save(&saveHistory)
	}
	if v.Id == 0 {
		*v = DBvideo
	}
	return nil
}

// GetVideoFullData 获取视频全部信息，包括作者，标签，收藏列表，观看历史列表。中间表，子表，和其他多对多的表数据
func GetVideoFullData(gromDb *gorm.DB, webSiteId int64, videoUuid string) Video {
	video := Video{}
	var tx *gorm.DB
	tx = gromDb.Debug().Where("web_site_id=? and uuid=?", webSiteId, videoUuid).Preload("Authors").Preload("Tag").Preload("CollectList").Preload("ViewHistory").First(&video)
	if tx.Error != nil {
		println(tx.Error.Error())
	}
	authorIds := make([]int64, 0)
	tagIds := make([]int64, 0)
	collectIds := make([]int64, 0)
	for _, videoAthorInfo := range video.Authors {
		authorIds = append(authorIds, videoAthorInfo.AuthorId)
	}
	for _, videoTagInfo := range video.Tag {
		tagIds = append(tagIds, videoTagInfo.TagId)
	}
	for _, videoCollectInfo := range video.CollectList {
		collectIds = append(collectIds, videoCollectInfo.CollectId)
	}
	video.StructAuthor = make([]Author, 0)
	video.StructTag = make([]Tag, 0)
	video.StructCollectList = make([]Collect, 0)
	gromDb.Where("id in (?)", authorIds).Find(&video.StructAuthor)
	gromDb.Where("id in (?)", tagIds).Find(&video.StructTag)
	gromDb.Where("id in (?)", collectIds).Find(&video.StructCollectList)
	return video

}

type VideoAuthor struct {
	Id         int64  `json:"id" gorm:"primary_key"`
	Uuid       string `json:"uuid"`
	VideoId    int64  `json:"video_id" gorm:"index:video_id"`
	AuthorId   int64  `json:"author_id" gorm:"index:author_id"`
	Contribute string `json:"contribute" gorm:"size:255"`
	AuthorUUID string `json:"-" gorm:"size:255"`
}

// Tag b站的tagId=0指bgm需要重新指定个id
type Tag struct {
	Id   int64  `json:"id" gorm:"index:id"`
	Name string `json:"name" gorm:"size:255"`
}

type VideoTag struct {
	Id      int64 `gorm:"primary_key"`
	VideoId int64 `gorm:"index:video_id"`
	TagId   int64 `gorm:"index:tag_id"`
}
