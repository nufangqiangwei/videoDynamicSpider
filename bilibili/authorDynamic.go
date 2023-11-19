package bilibili

import (
	"database/sql"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/models"
	"videoDynamicAcquisition/utils"
)

func getNotFollowAuthorDynamic() {
	db := baseStruct.CanUserDb()
	defer db.Close()
	authorList := models.GetAuthorList(db, 1)
	//rangeAuthor:
	rand.Seed(time.Now().Unix())
	for _, author := range authorList {
		if author.Follow {
			continue
		}
		var (
			mid    int
			err    error
			offset string
		)
		mid, err = strconv.Atoi(author.AuthorWebUid)
		if err != nil {
			continue
		}
		r, err := db.Query("select biliOffset from video where web_site_id=1 and author_id=? order by upload_time limit 1", author.Id)

		if err != nil {
			offset = ""
		} else if r.Next() {
			err = r.Scan(&offset)
			if err != nil {
				offset = ""
			}
		}

		saveError := false
		utils.Info.Printf("%s查询动态", author.AuthorName)
		for {
			utils.Info.Println(dynamicVideoObject.getRequest(mid, offset).URL.String())
			res := dynamicVideoObject.getResponse(0, mid, offset)
			if res == nil {
				utils.ErrorLog.Println("请求失败")
				break
			}
			for _, info := range res.Data.Items {
				if info.Type != "DYNAMIC_TYPE_AV" {
					continue
				}

				//baseline, ok = info.IdStr.(string)
				//if !ok {
				//	a, ok := info.IdStr.(int)
				//	if ok {
				//		baseline = strconv.Itoa(a)
				//	} else {
				//		saveError = true
				//		utils.ErrorLog.Println("未知的IDStr")
				//		utils.ErrorLog.Println(info.IdStr)
				//		break
				//	}
				//}
				mv := models.Video{
					WebSiteId:  1,
					AuthorId:   author.Id,
					Title:      info.Modules.ModuleDynamic.Major.Archive.Title,
					Desc:       strings.Replace(info.Modules.ModuleDynamic.Major.Archive.Desc, " ", "", -1),
					Duration:   HourAndMinutesAndSecondsToSeconds(info.Modules.ModuleDynamic.Major.Archive.DurationText),
					Uuid:       info.Modules.ModuleDynamic.Major.Archive.Bvid,
					Url:        "",
					CoverUrl:   info.Modules.ModuleDynamic.Major.Archive.Cover,
					UploadTime: time.Unix(info.Modules.ModuleAuthor.PubTs, 0),
				}
				if !mv.Save(db) {
					saveError = true
					break
					//goto rangeAuthor
				}
			}
			if saveError {
				break
			}
			if !res.Data.HasMore {
				break
			}
			offset = res.Data.Offset
			time.Sleep(time.Second * 10)
		}
		db.Exec("update author set follow=true where id=?", author.Id)
		utils.Info.Printf("%s查询动态完成\n", author.AuthorName)
		time.Sleep(time.Duration(rand.Intn(160)+60) * time.Second)
		return
	}
}

func getAuthorVideoList(authorId int64, db *sql.DB) {
	author := models.Author{}
	author.Get(authorId, db)
	if author.Follow {
		return
	}
	var pageIndex int
	videoPage := videoListPage{}
	utils.Info.Printf("%s查询视频列表\n", author.AuthorName)
	pageIndex = 1
	for {
		res := videoPage.getResponse(author.AuthorWebUid, pageIndex)
		if res == nil {
			utils.ErrorLog.Println("请求失败")
			break
		}
		for _, info := range res.Data.List.Vlist {
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
		}

		pageIndex++
	}
	db.Exec("update author set follow=true where id=?", author.Id)
	utils.Info.Printf("%s视频完成 ", author.AuthorName)
}
