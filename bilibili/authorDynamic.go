package bilibili

import (
	"database/sql"
	"path"
	"strconv"
	"time"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/models"
)

func getNotFollowAuthorDynamic() {
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	authorList := models.GetAuthorList(db, 1)
rangeAuthor:
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
		r, err := db.Query("select biliOffset from video where web_site_id=1 and author_id=? order by upload_time desc", author.Id)

		if err != nil {
			offset = ""
		} else {
			err = r.Scan(&offset)
			if err != nil {
				offset = ""
			}
		}
		for {
			res := dynamicVideoObject.getResponse(0, mid, offset)
			for _, info := range res.Data.Items {
				mv := models.Video{
					WebSiteId:  1,
					AuthorId:   author.Id,
					Title:      info.Modules.ModuleDynamic.Major.Archive.Title,
					Desc:       info.Modules.ModuleDynamic.Major.Archive.Desc,
					Duration:   HourAndMinutesAndSecondsToSeconds(info.Modules.ModuleDynamic.Major.Archive.DurationText),
					Uuid:       info.Modules.ModuleDynamic.Major.Archive.Bvid,
					Url:        "",
					CoverUrl:   info.Modules.ModuleDynamic.Major.Archive.Cover,
					UploadTime: time.Unix(info.Modules.ModuleAuthor.PubTs, 0),
					BiliOffset: info.IdStr,
				}
				if !mv.Save(db) {
					goto rangeAuthor
				}
			}
			if !res.Data.HasMore {
				break
			}
		}

	}
}
