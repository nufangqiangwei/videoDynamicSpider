package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
	"videoDynamicAcquisition/log"
)

// Video 视频信息
type Video struct {
	Id                  int64 `json:"id" gorm:"primary_key"`
	WebSiteId           int64
	Authors             []VideoAuthor   `gorm:"foreignKey:VideoId;references:Id"`
	Tag                 []VideoTag      `gorm:"foreignKey:VideoId;references:Id"`
	CollectList         []CollectVideo  `gorm:"foreignKey:VideoId;references:Id"`
	ViewHistory         []VideoHistory  `gorm:"foreignKey:VideoId;references:Id"`
	VideoPlayData       []VideoPlayData `gorm:"foreignKey:VideoId;references:Id"`
	Title               string          `gorm:"size:255"`
	VideoDesc           string          `gorm:"size:2000"`
	Duration            int             `gorm:"default:0;index:duration"`
	Uuid                string          `gorm:"size:255;uniqueIndex:uuid"`
	Url                 string          `gorm:"size:255"`
	CoverUrl            string          `gorm:"size:255"`
	UploadTime          *time.Time      `gorm:"default:null;index:upload_time"`
	CreateTime          time.Time       `gorm:"default:CURRENT_TIMESTAMP(3)"`
	Baid                int64           `json:"aid"`
	StructAuthor        []Author        `gorm:"-"`
	StructTag           []Tag           `gorm:"-"`
	StructCollectList   []Collect       `gorm:"-"`
	StructViewHistory   *VideoHistory   `gorm:"-"`
	StructVideoPlayData *VideoPlayData  `gorm:"-"`
	Classify            *Classify       `gorm:"-"`
}

func (v *Video) GetByUid(websiteName, uid string) error {
	tx := GormDB.Joins(
		"inner join web_site on web_site.id = video.web_site_id",
	).Where("web_site.web_name=? and uuid = ?", websiteName, uid).First(v)
	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		log.ErrorLog.Println("获取视频错误: ")
		log.ErrorLog.Println(tx.Error.Error())
		return tx.Error
	}
	return nil
}

// UpdateVideo 数据储存到数据库中，如果存在则更新，不存在则插入。many to many的表,如果中间表不存在就插入，存在就更新。这里只做增量更新，不做删除。删除操作，有别的同步方法自行执行。
func (v *Video) UpdateVideo() (bool, error) {
	if v.WebSiteId == 0 || v.Uuid == "" {
		return false, errors.New("缺少必要参数")
	}
	DBvideo := Video{}
	GormDB.Where("web_site_id=? and uuid=?", v.WebSiteId, v.Uuid).Find(&DBvideo)
	var isNew bool
	if DBvideo.Id == 0 {
		isNew = true
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

		GormDB.Create(&v)
		v.Authors = authorList
		v.Tag = tagList
		v.ViewHistory = historyList
		v.CollectList = collectList

		for _, author := range v.Authors {
			author.VideoId = v.Id
		}
		for _, tag := range v.Tag {
			tag.VideoId = v.Id
		}
		DBvideo.Id = v.Id
	} else {
		// 查出外键关联的数据
		GormDB.Preload("Authors").Preload("Tag").First(&DBvideo, DBvideo.Id)
	}
	var (
		have        bool
		saveAuthors []VideoAuthor
		lateAuthor  Author
		saveTags    []VideoTag
		lateTag     Tag
		saveHistory []VideoHistory
	)
	// 保存作者信息。video.Authors与v.Authors对比，如果v.Authors中有video.Authors中没有的，则插入。这里只做增量更新，不做删除。删除操作，有别的同步方法自行执行。
	// 排除已存在的作者信息
	for _, videoAuthor := range v.Authors {
		have = false
		for _, a := range DBvideo.Authors {
			if a.AuthorUUID == videoAuthor.AuthorUUID {
				have = true
				break
			}
		}
		if !have {
			// 没有存在视频作者信息,新插入的数据，先查询作者是否存在。
			lateAuthor = Author{}
			GormDB.Where("web_site_id=? and author_web_uid=?", v.WebSiteId, videoAuthor.AuthorUUID).Find(&lateAuthor)
			if lateAuthor.Id <= 0 {
				// 作者不存在，先插入作者信息。作者具体信息从v.StructAuthor中查找
				for _, a := range v.StructAuthor {
					if a.AuthorWebUid == videoAuthor.AuthorUUID {
						lateAuthor.WebSiteId = v.WebSiteId
						lateAuthor.AuthorName = a.AuthorName
						lateAuthor.AuthorWebUid = a.AuthorWebUid
						lateAuthor.Avatar = a.Avatar
						lateAuthor.AuthorDesc = a.AuthorDesc
						lateAuthor.FollowNumber = a.FollowNumber
						lateAuthor.CreateTime = time.Now()
						GormDB.Create(&lateAuthor)
						break
					}
				}
				if lateAuthor.Id == 0 {
					return isNew, errors.New("作者信息插入失败")
				}
			}
			videoAuthor.AuthorId = lateAuthor.Id
			videoAuthor.VideoId = DBvideo.Id
			saveAuthors = append(saveAuthors, videoAuthor)
		}

	}
	if len(saveAuthors) > 0 {
		GormDB.Save(&saveAuthors)
	}

	// 保存标签信息
	for _, videoTag := range v.Tag {
		have = false
		for _, t := range DBvideo.Tag {
			if t.TagId == videoTag.TagId {
				have = true
				break
			}
		}
		if !have {
			// 没有存在视频标签信息
			if videoTag.Id <= 0 {
				// 新插入的数据，先查询标签是否存在。
				lateTag = Tag{}
				if videoTag.TagId == 0 {
					// 该类型是bgm类型，查看name在tag表中是否存在，存在就取出这条数据。不存在就插入一条最新的tagId=max(id)+1的数据
					for _, t := range v.StructTag {
						if t.Id == 0 {
							GormDB.Where("name=?", t.Name).Find(&lateTag)
							if lateTag.Name == "" {
								GormDB.Exec("insert into tag (id,name) select max(id)+1,? from tag", t.Name)
							}
							GormDB.Where("name=?", t.Name).Find(&lateTag)
							break
						}
					}
				} else {
					GormDB.Where("id=?", videoTag.TagId).Find(&lateTag)
				}
				if lateTag.Name == "" {
					// 标签不存在，先插入标签信息。标签具体信息从v.StructTag中查找
					for _, t := range v.StructTag {
						if t.Id == videoTag.TagId {
							lateTag = t
							GormDB.Create(&lateTag)
							break
						}
					}
					if lateTag.Name == "" {
						return isNew, errors.New("标签信息插入失败")
					}
				}
			}
			videoTag.TagId = lateTag.Id
			videoTag.VideoId = DBvideo.Id
			saveTags = append(saveTags, videoTag)
		}
	}
	if len(saveTags) > 0 {
		GormDB.Save(&saveTags)
	}

	// 保存收藏信息,先查出StructCollectList所有的信息，然后在插入CollectVideo
	// 取出v.ViewHistory中的数据，添加video_id后保存
	if len(v.ViewHistory) > 0 {
		var databaseHistory VideoHistory
		GormDB.Model(&VideoHistory{}).Where("video_id=?", DBvideo.Id).Order("view_time desc").First(&databaseHistory)
		for _, history := range v.ViewHistory {
			if history.AuthorId == 0 {
				log.ErrorLog.Println("观看历史数据中没有用户信息")
				continue
			}
			history.WebSiteId = v.WebSiteId
			if history.ViewTime.After(databaseHistory.ViewTime) {
				history.VideoId = DBvideo.Id
				saveHistory = append(saveHistory, history)
			}
		}
	}
	if len(saveHistory) > 0 {
		GormDB.Save(&saveHistory)
	}

	// 视频分区数据
	if v.Classify != nil {
		var classify Classify
		GormDB.Model(&Classify{}).Where("id = ? and name = ?", v.Classify.Id, v.Classify.Name).Limit(1).Scan(&classify)
		if classify.Id == 0 {
			classify = Classify{
				Id:   v.Classify.Id,
				Name: v.Classify.Name,
			}
			GormDB.Create(&classify)
		}
		var videoClassify VideoClassify
		GormDB.Model(&VideoClassify{}).Where("video_id=?", DBvideo.Id).Limit(1).Scan(&videoClassify)
		if videoClassify.Id == 0 {
			videoClassify.VideoId = DBvideo.Id
			videoClassify.ClassifyId = classify.Id
			GormDB.Create(&videoClassify)
		}
	}

	// 播放数据
	if v.StructVideoPlayData != nil {
		v.StructVideoPlayData.VideoId = DBvideo.Id
		GormDB.Create(v.StructVideoPlayData)
	}

	if v.Id == 0 {
		*v = DBvideo
	}

	return isNew, nil
}

// GetVideoFullData 获取视频全部信息，包括作者，标签，收藏列表，观看历史列表。中间表，子表，和其他多对多的表数据
func GetVideoFullData(gromDb *gorm.DB, webSiteId int64, videoUuid string) *Video {
	video := Video{}
	var tx *gorm.DB
	tx = gromDb.Debug().Where("web_site_id=? and uuid=?", webSiteId, videoUuid).Preload("Authors").Preload("Tag").Preload("CollectList").Preload("ViewHistory").First(&video)
	if tx.Error != nil {
		println(tx.Error.Error())
		return nil
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
	return &video

}

func (v *Video) GetStructUniqueKey() string {
	return v.Uuid
}

type VideoAuthor struct {
	Id         int64  `json:"id" gorm:"primary_key"`
	Uuid       string `json:"uuid"` // 视频的uid
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

func (t *Tag) GetStructUniqueKey() string {
	return t.Name
}

func (t *Tag) GetOrCreate() {
	tx := GormDB.Where("name=?", t.Name).First(t)
	if tx.Error != nil {
		GormDB.Create(t)
	}
}

type VideoTag struct {
	Id      int64 `gorm:"primary_key"`
	VideoId int64 `gorm:"index:video_id"`
	TagId   int64 `gorm:"index:tag_id"`
}

type VideoPlayData struct {
	Id         int64     `json:"id" gorm:"primary_key"`
	VideoId    int64     `json:"video_id" gorm:"index:video_id"`
	View       int64     `gorm:"default:0"`          // 播放数
	Danmaku    int64     `gorm:"default:0"`          // 弹幕数
	Reply      int64     `gorm:"default:0"`          // 评论数
	Favorite   int64     `gorm:"default:0"`          // 收藏数
	Coin       int64     `gorm:"default:0"`          // 硬币数
	Share      int64     `gorm:"default:0"`          // 分享数
	NowRank    int64     `gorm:"default:0"`          // 当前排名
	HisRank    int64     `gorm:"default:0"`          // 历史最高排名
	Like       int64     `gorm:"default:0"`          // 点赞数
	Dislike    int64     `gorm:"default:0"`          // 点踩数
	Evaluation string    `gorm:"default:0;size:255"` // 综合评分
	CreateTime time.Time `gorm:"default:CURRENT_TIMESTAMP(3)"`
}

type Classify struct {
	Id    int64  `json:"id" gorm:"primary_key"`
	Name  string `json:"name" gorm:"size:255"`
	Level int    `json:"level"`
	Path  string `json:"path" gorm:"size:255"`
}
type VideoClassify struct {
	Id         int64 `gorm:"primary_key"`
	VideoId    int64 `gorm:"index:video_id"`
	ClassifyId int64 `gorm:"index:classify_id"`
}
