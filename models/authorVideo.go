package models

// 当一个视频有多个作者的时候使用
type AuthorVideo struct {
	Id        int64  `gorm:"primary_key" json:"id"`
	AuthorId  int64  `gorm:"type:int;notnull" json:"author_id"`
	VideoId   int64  `gorm:"type:int;notnull" json:"video_id"`
	VideoUUID string `gorm:"type:varchar(255);notnull" json:"video_uuid"`
}
