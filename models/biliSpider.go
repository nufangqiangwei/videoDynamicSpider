package models

import (
	"gorm.io/gorm"
	"strconv"
	"time"
)

const (
	defaultUserId int64 = 764886
)

// UserSpiderParams b站抓取记录
type UserSpiderParams struct {
	Id             int64  `gorm:"primaryKey"`
	AuthorId       int64  `gorm:"index"`
	KeyName        string `gorm:"size:255;"`
	Values         string `gorm:"size:255"`
	LastUpdateTime time.Time
	WebSiteId      int64
	UserId         int64
}

func (m *UserSpiderParams) BeforeUpdate(tx *gorm.DB) (err error) {
	m.LastUpdateTime = time.Now()
	return nil
}

func GetUserSpiderParam(webSiteName, userName, keyName string) string {
	result := UserSpiderParams{}
	err := GormDB.Model(&UserSpiderParams{}).Joins("inner join author on author.id=user_spider_params.author_id").Joins("inner join web_site on web_site.id=user_spider_params.web_site_id").Where("user_spider_params.key_name = ? AND web_site.web_name = ? AND author.author_name = ?", keyName, webSiteName, userName).Find(&result).Error
	if err != nil {
		return ""
	}
	return result.Values
}
func SaveUserSpiderParam(webSiteName, userName, keyName, values string) error {
	result := UserSpiderParams{}
	err := GormDB.Model(&UserSpiderParams{}).Joins("inner join author on author.id=user_spider_params.author_id").Joins("inner join web_site on web_site.id=user_spider_params.web_site_id").Where("user_spider_params.key_name = ? AND web_site.web_name = ? AND author.author_name = ?", keyName, webSiteName, userName).Find(&result).Error
	if err != nil {
		return err
	}
	result.Values = values
	return GormDB.Save(&result).Error
}

/*
-- 按照author_id查询每个用户的关注作者的最后更新时间
select 'dynamic_baseline', f.user_id, TRUNCATE(UNIX_TIMESTAMP(max(v.upload_time)),0) as upload_time, max(v.upload_time)
from follow f
         inner join video_author va on va.author_id = f.author_id
         inner join video v on v.id = va.video_id
where f.user_id in (select author_id from user_spider_params where key_name = 'dynamic_baseline')
group by f.user_id
union
-- 按照author_id查询每个用户的最后观看的时间
select 'history_baseline', vh.author_id, TRUNCATE(UNIX_TIMESTAMP(max(vh.view_time)),0) as view_time, max(vh.view_time)
from video_history vh
where vh.author_id in (select author_id from user_spider_params where key_name = 'history_baseline')
group by vh.author_id;
*/

func GetDynamicBaseline(userId int64) string {
	sql := "select  TRUNCATE(UNIX_TIMESTAMP(max(v.upload_time)),0)  from follow f inner join video_author va on va.author_id = f.author_id inner join video v on v.id = va.video_id where f.user_id =? group by f.user_id"
	result := 0
	if GormDB.Raw(sql, userId).Scan(&result).Error != nil {
		return ""
	}
	return strconv.Itoa(result)
}
func GetHistoryBaseline(userId int64) string {
	sql := "select TRUNCATE(UNIX_TIMESTAMP(max(vh.view_time)),0) from video_history vh where vh.author_id=? group by vh.author_id"
	result := 0
	if GormDB.Raw(sql, userId).Scan(&result).Error != nil {
		return ""
	}
	return strconv.Itoa(result)
}
