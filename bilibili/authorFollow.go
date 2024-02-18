package bilibili

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// api文档 https://socialsisteryi.github.io/bilibili-API-collect/docs/user/relation.html#操作关系
// api接口 https://api.bilibili.com/x/relation/modify
/*
### 操作用户关系

> https://api.bilibili.com/x/relation/modify

*请求方式：POST*

认证方式：Cookie（SESSDATA）或 APP

**正文参数（application/x-www-form-urlencoded）：**

| 参数名     | 类型 | 内容                      | 必要性          | 备注                                                     |
| ---------- | ---- | ------------------------- | --------------- | -------------------------------------------------------- |
| access_key | str  | APP 登录 Token            | APP 方式必要    |                                                          |
| fid        | num  | 目标用户mid               | 必要            |                                                          |
| act        | num  | 操作代码                  | 必要            | **操作代码见下表**                                       |
| re_src     | num  | 关注来源代码              | 必要            | 空间：11<br />视频：14<br />文章：115<br />活动页面：222 |
| csrf       | str  | CSRF Token（位于 Cookie） | Cookie 方式必要 | cookies中bili_jct的值                                       |

操作代码`act`：

| 代码 | 含义         |
| ---- | ------------ |
| 1    | 关注         |
| 2    | 取关         |
| 3    | 悄悄关注     |
| 4    | 取消悄悄关注 |
| 5    | 拉黑         |
| 6    | 取消拉黑     |
| 7    | 踢出粉丝     |

**json回复：**

根对象：

| 字段    | 类型 | 内容     | 备注                                                         |
| ------- | ---- | -------- | ------------------------------------------------------------ |
| code    | num  | 返回值   | 0：成功<br />-101：账号未登录<br />-102：账号被封停<br />-111：csrf校验失败<br />-400：请求错误<br />22001：不能对自己进行此操作<br />22003：用户位于黑名单 |
| message | str  | 错误信息 | 默认为0                                                      |
| ttl     | num  | 1        |                                                              |

**示例：**

关注`mid=14082`的用户

```shell
curl 'https://api.bilibili.com/x/relation/modify' \
    --data-urlencode 'fid=14082' \
    --data-urlencode 'act=1' \
    --data-urlencode 're_src=11' \
    --data-urlencode 'csrf=xxx' \
    -b 'SESSDATA=xxx'
```

<details>
<summary>查看响应示例：</summary>

```json
{
	"code": 0,
	"message": "0",
	"ttl": 1
}
```

</details>
*/
const (
	relationAuthorUrl  = "https://api.bilibili.com/x/relation/modify"
	FollowAuthor       = 1
	UnFollowAuthor     = 2
	HideFollowAuthor   = 3
	UnHideFollowAuthor = 4
	InToBlackList      = 5
	RemoveBlackList    = 6
	RemoveFollowers    = 7
)

type RelationAuthorRequestBody struct {
	Fid           int    `json:"fid"`
	Act           int    `json:"act"`
	ReSrc         int    `json:"re_src"`
	Csrf          string `json:"csrf"`
	GaiaSource    string `json:"gaia_source"`
	ExtendContent string `json:"extend_content"`
}
type RelationAuthorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
}

func RelationAuthor(action int, authorMid, user string) error {
	switch action {
	case FollowAuthor:
	case UnFollowAuthor:
	case HideFollowAuthor:
	case UnHideFollowAuthor:
	case InToBlackList:
	case RemoveBlackList:
	case RemoveFollowers:
		break
	default:
		return errors.New("未知的操作代码")
	}
	formData := url.Values{}
	formData.Set("fid", authorMid)
	formData.Set("act", strconv.Itoa(action))
	formData.Set("re_src", "11")
	formData.Set("csrf", biliCookiesManager.getUser(user).getCookiesKeyValue("bili_jct"))
	formEncoded := formData.Encode()

	request, err := http.NewRequest("POST", relationAuthorUrl, strings.NewReader(formEncoded))
	if err != nil {
		return err
	}
	request.Header.Add("Cookie", biliCookiesManager.getUser(user).cookies)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Referer", "https://space.bilibili.com/"+authorMid)
	request.Header.Add("Origin", "https://space.bilibili.com")
	request.Header.Add("Host", "api.bilibili.com")
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	result := new(RelationAuthorResponse)
	err = json.NewDecoder(response.Body).Decode(result)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", result)
	if result.Code != 0 {
		switch result.Code {
		case -101:
			return errors.New("账号未登录")
		case -102:
			return errors.New("账号被封停")
		case -111:
			return errors.New("csrf校验失败")
		case -400:
			return errors.New("请求错误")
		case 22001:
			return errors.New("不能对自己进行此操作")
		case 22003:
			return errors.New("用户位于黑名单")
		case 22014:
			return errors.New("已经关注用户，无法重复关注")
		default:
			return errors.New("未知错误: " + result.Message + " code: " + strconv.Itoa(result.Code))
		}
	}
	return nil
}
