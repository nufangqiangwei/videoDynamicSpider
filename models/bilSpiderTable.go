package models

import "database/sql"

type BiliAuthorVideoNumber struct {
	AuthorId    int64
	VideoNumber int
}

func (b *BiliAuthorVideoNumber) CreateTale() string {
	return `create table bili_author_video_number
(
    author_id    integer not null
        constraint bili_author_video_number_pk
            unique,
    video_number integer not null
)`
}

func (b *BiliAuthorVideoNumber) GetAuthorVideoNumber(authorId int64, db *sql.DB) {
	r, err := db.Query("select video_number from bili_author_video_number where author_id=?", authorId)
	if err != nil {
		print("查询用户投稿数错误")
		println(err.Error())
		return
	}
	b.AuthorId = authorId
	if r.Next() {
		err = r.Scan(&b.VideoNumber)
		if err != nil {
			println(err.Error())
		}
	}
	r.Close()
}

func (b *BiliAuthorVideoNumber) UpdateNumber(db *sql.DB) {
	if b.AuthorId == 0 {
		println("AuthorId")
		return
	}
	if b.VideoNumber == 0 {
		println("VideoNumber")
		return
	}
	_, err := db.Exec("insert into bili_author_video_number values (?,?)", b.AuthorId, b.VideoNumber)
	if err != nil {
		println(err.Error())
		// 已存在这个up的数据，改为更新
		r, err := db.Exec("update bili_author_video_number set video_number=? where author_id=?", b.VideoNumber, b.AuthorId)
		println(r.LastInsertId())
		if err != nil {
			println(err.Error())
		}

	}
}
