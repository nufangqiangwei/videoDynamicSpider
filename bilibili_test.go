package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"
	"videoDynamicAcquisition/models"
)

func TestGetVideo(t *testing.T) {
	spider := makeBilibiliSpider()
	for _, info := range spider.getVideoList() {
		fmt.Printf("%+v\n", info)
	}
}
func TestUrl(t *testing.T) {
	data, err := os.ReadFile("data.json")
	if err != nil {
		println(err.Error())
		return
	}
	responseList := []bilbilVideoInfo{}
	err = json.Unmarshal(data, &responseList)
	if err != nil {
		println(err.Error())
		return
	}
	db, _ := sql.Open("sqlite3", path.Join("C:\\Code\\GO\\videoDynamicSpider", sqliteDaName))
	website := models.WebSite{
		WebName:          "bilibili",
		WebHost:          "https://www.bilibili.com/",
		WebAuthorBaseUrl: "https://space.bilibili.com/",
		WebVideoBaseUrl:  "https://www.bilibili.com/",
	}
	website.GetOrCreate(db)

	for _, info := range responseList {
		video := VideoInfo{
			WebSite:    "bilibili",
			Title:      info.Modules.ModuleDynamic.Major.Archive.Title,
			VideoUuid:  info.Modules.ModuleDynamic.Major.Archive.Bvid,
			Url:        info.Modules.ModuleDynamic.Major.Archive.JumpUrl,
			CoverUrl:   info.Modules.ModuleDynamic.Major.Archive.Cover,
			AuthorUuid: strconv.Itoa(info.Modules.ModuleAuthor.Mid),
			AuthorName: info.Modules.ModuleAuthor.Name,
			AuthorUrl:  info.Modules.ModuleAuthor.JumpUrl,
			PushTime:   info.Modules.ModuleAuthor.PubTs,
		}
		author := models.Author{AuthorName: video.AuthorName, WebSiteId: website.Id, AuthorWebUid: video.AuthorUuid}
		author.GetOrCreate(db)
		videoModel := models.Video{
			WebSiteId:  website.Id,
			AuthorId:   author.Id,
			Title:      video.Title,
			Url:        video.Url,
			Uuid:       video.VideoUuid,
			CoverUrl:   video.CoverUrl,
			UploadTime: video.PushTime,
		}
		videoModel.Save(db)
	}
}

func TestModifyVideoPushTime(t *testing.T) {
	data, err := os.ReadFile("data.json")
	if err != nil {
		println(err.Error())
		return
	}
	responseList := []bilbilVideoInfo{}
	err = json.Unmarshal(data, &responseList)
	if err != nil {
		println(err.Error())
		return
	}
	db, _ := sql.Open("sqlite3", path.Join("C:\\Code\\GO\\videoDynamicSpider", sqliteDaName))
	defer db.Close()
	rows, err := db.Query("select id,uuid from video")
	if err != nil {
		println(err.Error())
		return
	}
	var (
		videoId   int64
		videoUUId string
	)
	updateSql := make([]string, 210)
	index := 0
	for rows.Next() {
		err = rows.Scan(&videoId, &videoUUId)
		if err != nil {
			println(err.Error())
			break
		}
		print(videoUUId, "  ")
		for _, videoInfo := range responseList {
			if videoInfo.Modules.ModuleDynamic.Major.Archive.Bvid == videoUUId {
				println(videoInfo.Modules.ModuleDynamic.Major.Archive.Bvid)
				updateSql[index] = fmt.Sprintf("update video set upload_time=%d where id=%d", videoInfo.Modules.ModuleAuthor.PubTs, videoId)
				index++
				break
			}
		}

	}
	for _, update_sql := range updateSql {
		println(update_sql)
		if update_sql == "" {
			continue
		}
		r, err := db.Exec(update_sql)
		if err != nil {
			println(err.Error())
		}
		i, err := r.RowsAffected()
		println(i)
		if err != nil {
			println(err.Error())
		}
	}

}
