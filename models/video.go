package models

import (
	"gorm.io/gorm"
	"time"
	"videoDynamicAcquisition/utils"
)

// Video 视频信息
type Video struct {
	Id               int64 `json:"id" gorm:"primary_key"`
	WebSiteId        int64
	Authors          []VideoAuthor `gorm:"foreignKey:VideoId;references:Id"`
	Tag              []VideoTag    `gorm:"foreignKey:VideoId;references:Id"`
	Title            string
	VideoDesc        string
	Duration         int
	Uuid             string
	Url              string
	CoverUrl         string
	UploadTime       *time.Time
	CreateTime       time.Time `gorm:"default:CURRENT_TIMESTAMP(3)"`
	IsMultiplePeople bool      `gorm:"default:false"`
	View             int64     `gorm:"default:0"` // 播放数
	Danmaku          int64     `gorm:"default:0"` // 弹幕数
	Reply            int64     `gorm:"default:0"` // 评论数
	Favorite         int64     `gorm:"default:0"` // 收藏数
	Coin             int64     `gorm:"default:0"` // 硬币数
	Share            int64     `gorm:"default:0"` // 分享数
	NowRank          int64     `gorm:"default:0"` // 当前排名
	HisRank          int64     `gorm:"default:0"` // 历史最高排名
	Like             int64     `gorm:"default:0"` // 点赞数
	Dislike          int64     `gorm:"default:0"` // 点踩数
	Evaluation       string    `gorm:"default:0"` // 综合评分
}

func (v *Video) Save() bool {
	video := Video{}
	video.GetByUid(v.Uuid)
	var tx *gorm.DB
	if video.Id == 0 {
		tx = GormDB.Save(v)
	} else {
		for _, author := range v.Authors {
			author.VideoId = v.Id
		}
		tx = GormDB.Save(&v.Authors)
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
	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		utils.ErrorLog.Println("获取视频错误: ")
		utils.ErrorLog.Println(tx.Error.Error())
	}
}

type VideoAuthor struct {
	Id         int64  `json:"id" gorm:"primary_key"`
	Uuid       string `json:"uuid"`
	VideoId    int64  `json:"video_id" gorm:"index:video_id"`
	AuthorId   int64  `json:"author_id" gorm:"index:author_id"`
	Contribute string `json:"contribute"`
}
type Tag struct {
	Id   int64  `json:"id" gorm:"index:id"`
	Name string `json:"name"`
}
type VideoTag struct {
	Id      int64 `gorm:"primary_key"`
	VideoId int64 `gorm:"index:video_id"`
	TagId   int64 `gorm:"index:tag_id"`
}
