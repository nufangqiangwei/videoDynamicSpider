package main

import (
	"time"
	"videoDynamicAcquisition/models"
	webSiteGRPC "videoDynamicAcquisition/proto"
)

func handleUserCollectList(vi *webSiteGRPC.CollectionInfo) {
	videoList := make([]*models.Video, 0)
	collectInfo := models.Collect{}
	models.GormDB.Model(&collectInfo).Where("bv_id = ?", vi.CollectionId).Scan(&collectInfo)
	models.GormDB.Model(&models.Video{}).Joins(
		"inner join collect_video on collect_video.video_id = video.id inner join collect on collect.id = collect_video.collect_id",
	).Where("collect.bv_id = ? and collect_video.is_del = false and collect_video.is_invalid = false", vi.CollectionId).Order("collect_video.mtime desc").Find(&videoList)
	if collectInfo.Id == 0 {
		collectInfo = models.Collect{
			BvId:     vi.CollectionId,
			Name:     vi.Name,
			AuthorId: vi.RequestUserId,
		}
		if vi.CollectionType == "folder" {
			// 个人创建的收藏夹
			collectInfo.Type = 1
		} else if vi.CollectionType == "subscription" {
			// 订阅的合集
			collectInfo.Type = 2
		}
		models.GormDB.Create(&collectInfo)
	}
	// 遍历两边的视频列表，找出差集
	for _, v := range vi.Video {
		have := false
		for _, aa := range videoList {
			if v.Uid == aa.Uuid {
				have = true
				break
			}
		}
		if !have {
			// 查找视频是否存在
			videoInfo := models.Video{}
			models.GormDB.Model(&models.Video{}).Where("uuid = ?", v.Uid).Scan(&videoInfo)
			if videoInfo.Id == 0 {
				updateTime := time.Unix(v.UpdateTime, 0)
				videoInfo = models.Video{
					WebSiteId:  vi.WebSiteId,
					Title:      v.Title,
					VideoDesc:  v.Desc,
					Duration:   int(v.Duration),
					Uuid:       v.Uid,
					CoverUrl:   v.Cover,
					UploadTime: &updateTime,
					StructAuthor: []models.Author{
						{
							AuthorWebUid: v.Authors[0].Uid,
							Avatar:       v.Authors[0].Avatar,
							AuthorName:   v.Authors[0].Name,
						},
					},
				}
				models.GormDB.Create(&videoInfo)
			}
			collectTime := time.Unix(v.CollectTime, 0)
			models.GormDB.Create(&models.CollectVideo{
				CollectId: collectInfo.Id,
				VideoId:   videoInfo.Id,
				Mtime:     &collectTime,
			})
		}
	}
	for _, v := range videoList {
		have := false
		var xx *webSiteGRPC.VideoInfoResponse
		for _, aa := range vi.Video {
			if v.Uuid == aa.Uid {
				have = true
				xx = aa
				break
			}
		}
		if !have {
			models.GormDB.Exec(`update collect_video set is_del = true where video_id = ?`, v.Id)
		} else if xx.IsInvalid {
			models.GormDB.Exec(`update collect_video set is_invalid = true where video_id = ?`, v.Id)
		}
	}
}
