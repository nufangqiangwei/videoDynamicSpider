package bilibili

import (
	"database/sql"
	"encoding/json"
	"os"
	"path"
	"strconv"
	"testing"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/models"
)

func TestFollowings(t *testing.T) {
	f := followings{}
	bilibiliCookies.readFile()
	allFollowing := f.getFollowings(1)
	data, err := json.Marshal(allFollowing)
	if err != nil {
		println(err.Error())
	}
	os.WriteFile("E:\\GoCode\\videoDynamicAcquisition\\bilibili-followings.json", data, 0666)
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	for _, author := range allFollowing {
		author.GetOrCreate(db)
	}
}

type JSONDATA struct {
	Page1 followingsResponse `json:"page1"`
	Page2 followingsResponse `json:"page2"`
	Page3 followingsResponse `json:"page3"`
	Page4 followingsResponse `json:"page4"`
	Page5 followingsResponse `json:"page5"`
	Page6 followingsResponse `json:"page6"`
	Page7 followingsResponse `json:"page7"`
	Page8 followingsResponse `json:"page8"`
}

func TestImportFollowingJSON(t *testing.T) {
	data, err := os.ReadFile("C:\\Code\\GO\\videoDynamicSpider\\bilibili\\bilibili-followings.json")
	if err != nil {
		println(err.Error())
		return
	}
	jsonData := JSONDATA{}
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		println(err.Error())
		return
	}
	db, _ := sql.Open("sqlite3", path.Join(baseStruct.RootPath, baseStruct.SqliteDaName))
	total := 0
	saveData(jsonData.Page1, db, 1)
	saveData(jsonData.Page2, db, 1)
	saveData(jsonData.Page3, db, 1)
	saveData(jsonData.Page4, db, 1)
	saveData(jsonData.Page5, db, 1)
	saveData(jsonData.Page6, db, 1)
	saveData(jsonData.Page7, db, 1)
	saveData(jsonData.Page8, db, 1)
	println(total)
}

func saveData(data followingsResponse, db *sql.DB, webSiteId int64) {
	for _, info := range data.Data.List {
		m := models.Author{
			WebSiteId:    webSiteId,
			AuthorWebUid: strconv.FormatInt(info.Mid, 10),
			AuthorName:   info.Uname,
			Avatar:       info.Face,
			Desc:         info.Sign,
		}
		m.GetOrCreate(db)
	}
}

func TestGOTO(t *testing.T) {
	getNotFollowAuthorDynamic()
}
