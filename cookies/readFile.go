package cookies

import (
	"os"
	"path"
	"strings"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/log"
)

// 实现 baseStruct.CookiesFlush 接口
type privateReadLocalFile struct {
}

func (wsc privateReadLocalFile) WebSiteList() []string {
	result := make([]string, 0)
	files, err := os.ReadDir(path.Join(baseStruct.RootPath, cookiesFileFolder))
	if err != nil {
		println("读取cookies文件夹失败")
		println(err.Error())
		if log.ErrorLog != nil {
			log.ErrorLog.Println("读取cookies文件夹失败")
		} else {
			println("读取cookies文件夹失败")
		}
		return result
	}
	for _, file := range files {
		result = append(result, file.Name())
	}
	return result
}
func (wsc privateReadLocalFile) UserList(webName string) []baseStruct.CacheUserCookies {
	result := make([]baseStruct.CacheUserCookies, 0)
	files, err := os.ReadDir(path.Join(baseStruct.RootPath, cookiesFileFolder, webName))
	if err != nil {
		if log.ErrorLog != nil {
			log.ErrorLog.Printf("读取%scookies文件夹失败", webName)
		} else {
			println("读取", webName, "网站cookies文件夹失败")
			println(err.Error())
		}
		return result
	}
	for _, file := range files {
		fileContent, err := os.ReadFile(path.Join(baseStruct.RootPath, cookiesFileFolder, webName, file.Name()))
		if err != nil {
			if log.ErrorLog != nil {
				log.ErrorLog.Printf("读取%scookies文件夹失败", webName)
			} else {
				println("读取", webName, "网站cookies文件夹失败")
				println(err.Error())
			}
			return result
		}
		result = append(result, baseStruct.CacheUserCookies{UserName: file.Name(), Content: strings.TrimSpace(string(fileContent))})
	}
	return result
}
func (wsc privateReadLocalFile) GetUserCookies(webSiteName, userName string) string {
	fileContent, err := os.ReadFile(path.Join(baseStruct.RootPath, cookiesFileFolder, webSiteName, userName))
	if err != nil {
		if log.ErrorLog != nil {
			log.ErrorLog.Printf("读取%scookies文件失败", userName)
		} else {
			println("读取", userName, "cookies文件失败")
			println(err.Error())
		}
		return ""
	}
	return string(fileContent)
}
func (wsc privateReadLocalFile) UpdateUserCookies(webSiteName, authorName, cookiesContent, userId string) error {
	// 将cookies保存到本地文件夹中
	if userId == blankUserName {
		return nil
	}
	webSitePath := path.Join(baseStruct.RootPath, cookiesFileFolder, webSiteName)
	err := os.MkdirAll(webSitePath, 0666)
	if err != nil {
		log.ErrorLog.Printf("创建文件夹出错,%s", err.Error())
		return nil
	}
	filePath := path.Join(webSitePath, authorName)
	log.ErrorLog.Println(webSiteName, "网站保存用户", authorName, "Cookies文件。文件地址是：", filePath)
	err = os.WriteFile(filePath, []byte(cookiesContent), 0666)
	if err != nil {
		log.ErrorLog.Printf(err.Error())
	}
	return err
}
func (wsc privateReadLocalFile) UserCookiesInvalid(webSiteName, authorName, cookiesContent, userId string) error {
	return nil
}
func (wsc privateReadLocalFile) GetTouristsCookies(webName string) []string {
	return []string{}
}
