package bilibili

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/cookies"
	"videoDynamicAcquisition/log"
	"videoDynamicAcquisition/models"
)

func TestMain(m *testing.M) {
	baseStruct.RootPath = "E:\\GoCode\\videoDynamicAcquisition"
	log.InitLog(baseStruct.RootPath)
	models.InitDB("spider:p0o9i8u7@tcp(database:3306)/video?charset=utf8mb4&parseTime=True&loc=Local", false, nil)
	//cookies.DataSource = models.WebSiteCookies{}
	cookies.FlushAllCookies()
	os.Exit(m.Run())
}

func TestHistory(t *testing.T) {
	vd := videoDetail{}
	vd.getResponse("BV117411r7R1")
}

func TestDynamic(t *testing.T) {
	dynamicVideoObject := dynamicVideo{}
	response := dynamicVideoObject.getResponse(0, 0, "", false)
	if response == nil {
		println("获取失败")
	} else {
		fmt.Printf("%v\n", response)
	}
}

/*
{"code":-101,"message":"账号未登录","ttl":1}
*/
func TestJSONDynamic(t *testing.T) {
	body := []byte(`{"code":-101,"message":"账号未登录","ttl":1}`)
	a := DynamicResponse{}
	err := json.Unmarshal(body, &a)
	if err != nil {
		print(err.Error())
		return
	}
	err = a.bindJSON(body)
	if err != nil {
		print(err.Error())
		return
	}
	fmt.Printf("%v\n", a)
}

func TestRelationAuthor(t *testing.T) {
	err := RelationAuthor(FollowAuthor, "3493118584293963", cookies.UserCookie{})
	if err != nil {
		println(err.Error())
		return
	}
}
func TestBVAV(t *testing.T) {
	av := Bv2Av("BV1bV411S7Le")
	println(av)
	bv := Av2Bv(411857180)
	println(bv)
}

func TestBiliSpider_GetUserDynamic(t *testing.T) {
	cookies.FlushAllCookies()
	dynamicBaseLineMap = map[string]int64{
		"卢生啊": 1711268019, "干煸花椰菜": 1711322236,
	}
	historyBaseLineMap = map[string]int64{
		"卢生啊": 1709186160, "干煸花椰菜": 1711322124,
	}
	resultChan := make(chan models.Video)
	closeChan := make(chan models.TaskClose)
	go Spider.GetVideoList(resultChan, closeChan)
	for {
		select {
		case v := <-resultChan:
			fmt.Printf("标题:%s\n", v.Title)
		case c := <-closeChan:
			fmt.Printf("关闭:%v\n", c)
			return
		}
	}
}

func getRequest() *http.Request {
	url := "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/all"
	request, _ := http.NewRequest("GET", url, nil)
	q := request.URL.Query()
	q.Add("type", "video")

	request.URL.RawQuery = q.Encode()
	request.Header.Add("Cookie", "LIVE_BUVID=AUTO6015975742804240; i-wanna-go-back=-1; CURRENT_BLACKGAP=0; buvid_fp_plain=undefined; b_timer=%7B%22ffp%22%3A%7B%22333.1007.fp.risk_4609B7F9%22%3A%22182B928B2A2%22%2C%22333.788.fp.risk_4609B7F9%22%3A%22182B9289B66%22%2C%22333.880.fp.risk_4609B7F9%22%3A%221811447FB84%22%2C%22333.967.fp.risk_4609B7F9%22%3A%221816C2A2C0C%22%2C%22444.8.fp.risk_4609B7F9%22%3A%221825ED57D31%22%2C%22333.976.fp.risk_4609B7F9%22%3A%22182785823B2%22%2C%22888.2421.fp.risk_4609B7F9%22%3A%22180F4EBB0E9%22%2C%22444.7.fp.risk_4609B7F9%22%3A%221825F17A69A%22%2C%22444.13.fp.risk_4609B7F9%22%3A%22181B36A01FF%22%2C%22777.5.0.0.fp.risk_4609B7F9%22%3A%22180F4EBAACE%22%2C%22333.337.fp.risk_4609B7F9%22%3A%221829AE2AF70%22%2C%22333.999.fp.risk_4609B7F9%22%3A%221829AE2D3DC%22%2C%22444.55.fp.risk_4609B7F9%22%3A%221816C4F28E3%22%2C%22333.905.fp.risk_4609B7F9%22%3A%22182C61C93EE%22%2C%22333.1073.fp.risk_4609B7F9%22%3A%2218251E7956A%22%2C%22333.937.fp.risk_4609B7F9%22%3A%221825F14038A%22%2C%22888.70696.fp.risk_4609B7F9%22%3A%221825F16DC22%22%2C%22333.1193.fp.risk_4609B7F9%22%3A%22182B66389F4%22%7D%7D; header_theme_version=CLOSE; theme_style=light; FEED_LIVE_VERSION=V8; hit-new-style-dyn=1; hit-dyn-v2=1; DedeUserID=10932398; DedeUserID__ckMd5=d5bdc808e48784b2; b_ut=5; go-back-dyn=0; buvid3=FE124433-BDEA-4CC0-B3F1-E5A3DD37F23292667infoc; b_nut=1692186191; _uuid=76D551087-3E32-4E12-46FE-E6B410E1CC62733545infoc; buvid4=1EC43764-D911-026E-04BC-A666638B1AE237315-022012118-dJ30JF9%2BylKXU%2FQmcpKHPA%3D%3D; i-wanna-go-feeds=2; kfcSource=cps_comments_4624257; msource=cps_comments_4624257; deviceFingerprint=0981a6f5338d4165063a8c2ad4f98796; enable_web_push=DISABLE; CURRENT_QUALITY=80; rpdid=0zbfVGePcl|1a5eQ1fzL|15|3w1RgDx9; share_source_origin=COPY; fingerprint=bef0fc320e7e23f4933b01d05e9e99f9; buvid_fp=bef0fc320e7e23f4933b01d05e9e99f9; home_feed_column=5; SESSDATA=af9d2d49%2C1727695075%2Cf1fc5%2A42CjATOD0PoFtv52J1Fud_vvyaU4Axa-1ulpMcCyPqDqegf8vMbuFx9iN3GxSy6Eavy0ESVnZPM1JVSFByX1A1SWVUOU5qdnEtaXNjSzBVYjI5WmJqeDJkUnVDcEg5bnVObEY0UzNLM255MV9qZHFDcEV0Tk9OcXQ5WlAzRWlGNnpfcDQ1Y3pVZzZRIIEC; bili_jct=ed31da7513909c55624b4f3d67c3d2e2; bp_video_offset_10932398=916136530237456407; sid=6w9nn26c; bili_ticket=eyJhbGciOiJIUzI1NiIsImtpZCI6InMwMyIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTI0OTk4MzUsImlhdCI6MTcxMjI0MDU3NSwicGx0IjotMX0.EYTMIoECXBPwpgqW_BQR08TK6fNL44iw8-9Jp2dnOJI; bili_ticket_expires=1712499775; PVID=1; browser_resolution=1920-919; CURRENT_FNVAL=4048; bsource=search_google; b_lsid=E1EC106E4_18EB1E8C7CB")
	return request
}
func TestRequestUseProxy(t *testing.T) {
	r := getRequest()
	proxy, _ := url.Parse("http://127.0.0.1:1080")
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(r)
	if err != nil {
		fmt.Println("请求失败：", err)
		return
	}
	defer resp.Body.Close()
	println(resp.StatusCode)
	dynamicResponseBody := &DynamicResponse{}
	err = responseCodeCheck(resp, dynamicResponseBody, cookies.NewDefaultUserCookie(webSiteName))
	if err != nil {
		println(err.Error())
	}
	fmt.Printf("%v\n", *dynamicResponseBody)
}
