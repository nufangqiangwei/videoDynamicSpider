package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
	"videoDynamicAcquisition/utils"
)

// Video 视频信息
type Video struct {
	Id                int64 `json:"id" gorm:"primary_key"`
	WebSiteId         int64
	Authors           []VideoAuthor  `gorm:"foreignKey:VideoId;references:Id"`
	Tag               []VideoTag     `gorm:"foreignKey:VideoId;references:Id"`
	ViewHistory       []VideoHistory `gorm:"foreignKey:VideoId;references:Id"`
	CollectList       []CollectVideo `gorm:"foreignKey:VideoId;references:Id"`
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

func (v *Video) sSave() bool {
	video := Video{}
	video.GetByUid(v.Uuid)
	var tx *gorm.DB
	if video.Id == 0 {
		tx = GormDB.Save(v)
	} else {
		var (
			authors     []VideoAuthor
			authorHave  bool
			saveAuthors []VideoAuthor
		)
		GormDB.Where("video_id=?", video.Id).Find(&authors)
		// 排除已存在的作者信息
		for _, author := range v.Authors {
			authorHave = false
			for _, a := range authors {
				if a.AuthorId == author.AuthorId {
					authorHave = true
					break
				}
			}
			if !authorHave {
				author.VideoId = video.Id
				saveAuthors = append(saveAuthors, author)
			}
		}
		tx = GormDB.Save(&saveAuthors)
		v.Id = video.Id
	}
	if tx.Error != nil {
		utils.ErrorLog.Println("保存视频错误: ")
		utils.ErrorLog.Println(tx.Error.Error())
		return false
	}
	return true
}

func (v *Video) GetByUid(uid string) {
	tx := GormDB.Where("uuid = ?", uid).First(v)
	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		utils.ErrorLog.Println("获取视频错误: ")
		utils.ErrorLog.Println(tx.Error.Error())
	}
}

// 数据储存到数据库中，如果存在则更新，不存在则插入。many to many的表,如果中间表不存在就插入，存在就更新。这里只做增量更新，不做删除。删除操作，有别的同步方法自行执行。
func (v *Video) UpdateVideo() error {
	if v.WebSiteId == 0 || v.Uuid == "" {
		return errors.New("缺少必要参数")
	}
	DBvideo := Video{}
	GormDB.Where("web_site_id=? and uuid=?", v.WebSiteId, v.Uuid).Find(&DBvideo)
	if DBvideo.Id == 0 {
		var (
			authorList  []VideoAuthor
			tagList     []VideoTag
			historyList []VideoHistory
			collectList []CollectVideo
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
					return errors.New("作者信息插入失败")
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
						return errors.New("标签信息插入失败")
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
	// 取出v.ViewHistory中的数据，添加video_id后保存
	for _, history := range v.ViewHistory {
		history.VideoId = DBvideo.Id
		saveHistory = append(saveHistory, history)
	}
	if len(saveHistory) > 0 {
		GormDB.Save(&saveHistory)
	}
	return nil
}

type VideoAuthor struct {
	Id         int64  `json:"id" gorm:"primary_key"`
	Uuid       string `json:"uuid"`
	VideoId    int64  `json:"video_id" gorm:"index:video_id"`
	AuthorId   int64  `json:"author_id" gorm:"index:author_id"`
	Contribute string `json:"contribute" gorm:"size:255"`
	AuthorUUID string `json:"-" gorm:"size:255"`
}

// b站的tagId=0指bgm需要重新指定个id
type Tag struct {
	Id   int64  `json:"id" gorm:"index:id"`
	Name string `json:"name" gorm:"size:255"`
}
type VideoTag struct {
	Id      int64 `gorm:"primary_key"`
	VideoId int64 `gorm:"index:video_id"`
	TagId   int64 `gorm:"index:tag_id"`
}
