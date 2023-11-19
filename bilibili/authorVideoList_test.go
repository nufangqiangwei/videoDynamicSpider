package bilibili

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/models"
)

func TestGetOneMoreAuthor(t *testing.T) {
	var (
		authorList   []int64
		authorId     int64
		videoUuid    string
		pageIndex    int
		appendNumber int
		ok           bool
	)
	authorList = []int64{
		159,
	}

	db := baseStruct.CanUserDb()
	defer db.Close()
	for _, authorId = range authorList {
		pageIndex = 16
		appendNumber = 0
		author := models.Author{}
		author.Get(authorId, db)
		authorVideoNumber := models.BiliAuthorVideoNumber{
			AuthorId: authorId,
		}
		authorVideoNumber.GetAuthorVideoNumber(authorId, db)

		r, _ := db.Query("select uuid from video where author_id=?", authorId)

		saveUuidMap := map[string]int{}
		for r.Next() {
			err := r.Scan(&videoUuid)
			if err != nil {
				println("绑定数据错误")
				println(err.Error())
			}
			saveUuidMap[videoUuid] = 0
		}
		r.Close()

		videoPage := videoListPage{}

		for {
			res := videoPage.getResponse(author.AuthorWebUid, pageIndex)
			if res == nil {
				println("请求失败")
				break
			}
			for _, info := range res.Data.List.Vlist {
				_, ok = saveUuidMap[info.Bvid]
				if ok {
					continue
				}
				mv := models.Video{
					WebSiteId:  1,
					AuthorId:   author.Id,
					Title:      info.Title,
					Desc:       strings.Replace(info.Description, " ", "", -1),
					Duration:   HourAndMinutesAndSecondsToSeconds(info.Length),
					Uuid:       info.Bvid,
					Url:        "",
					CoverUrl:   info.Pic,
					UploadTime: time.Unix(int64(info.Created), 0),
				}

				mv.Save(db)
				appendNumber++

			}

			if authorVideoNumber.VideoNumber != res.Data.Page.Count && authorVideoNumber.VideoNumber == 0 {
				authorVideoNumber.VideoNumber = res.Data.Page.Count
				authorVideoNumber.UpdateNumber(db)
			}
			time.Sleep(time.Second)
			if (pageIndex * 50) >= authorVideoNumber.VideoNumber {
				break
			}
			if len(saveUuidMap)+appendNumber == authorVideoNumber.VideoNumber {
				break
			}
			pageIndex++
		}
		print(author.AuthorName)
		println("结束")
	}

}

func TestFind(t *testing.T) {
	fileName := fmt.Sprintf("%s\\video\\%s\\%d.json", baseStruct.RootPath, "2026561407", 1)
	data, _ := os.ReadFile(fileName)
	responseBody := new(VideoListPageResponse)
	json.Unmarshal(data, responseBody)
	db := baseStruct.CanUserDb()
	defer db.Close()
	for _, info := range responseBody.Data.List.Vlist {
		mv := models.Video{
			WebSiteId:  1,
			AuthorId:   240,
			Title:      info.Title,
			Desc:       strings.Replace(info.Description, " ", "", -1),
			Duration:   HourAndMinutesAndSecondsToSeconds(info.Length),
			Uuid:       info.Bvid,
			Url:        "",
			CoverUrl:   info.Pic,
			UploadTime: time.Unix(int64(info.Created), 0),
		}
		mv.Save(db)
	}

}
func TestDNS(t *testing.T) {
	userName := "root@mingzhe#YOKA:1"
	println(url.QueryEscape(userName))
}

/*
select follow, count(*)
from author
group by follow;

select a.id, a.author_name,a.follow,a.author_web_uid, vs.videos, bv.video_number
from author a
         left join (select video.author_id, count(*) as videos
                    from video
                    group by video.author_id) vs on vs.author_id = a.id
         left join bili_author_video_number bv on a.id = bv.author_id
where web_site_id = 1
  and bv.video_number is not null
  and vs.videos <> bv.video_number;



select a.author_name, video.author_id, count(video.id), bv.video_number
from video
         left join author a on a.id = video.author_id
         left join bili_author_video_number bv on video.author_id = bv.author_id
where bv.video_number is not null
group by video.author_id;
*/
