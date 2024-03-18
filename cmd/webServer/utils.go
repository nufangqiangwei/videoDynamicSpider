package main

import (
	"archive/tar"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"videoDynamicAcquisition/baseStruct"
	"videoDynamicAcquisition/log"
)

// 用户通过post请求上传tar文件，程序解压到nginx的静态文件位置。
func deployWebSIteHtmlFile(gtx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.ErrorLog.Println(r)
			gtx.String(500, "Internal Server Error")
		}
	}()
	log.Info.Println("开始读取前端文件")

	file, err := gtx.FormFile("file")
	if err != nil {
		log.ErrorLog.Println(err)
		gtx.String(400, "Bad Request1")
		return
	}
	// 判断文件是否是tar.gz文件
	if !strings.HasSuffix(file.Filename, ".tar.gz") {
		log.ErrorLog.Println(err)
		gtx.String(400, "Bad Request2")
		return
	}

	uploadFile, err := file.Open()
	if err != nil {
		log.ErrorLog.Println(err)
		gtx.String(400, "Bad Request3")
		return
	}
	defer uploadFile.Close()
	gzRead, err := gzip.NewReader(uploadFile)
	if err != nil {
		log.ErrorLog.Println(err)
		return
	}
	defer gzRead.Close()
	tarRead := tar.NewReader(gzRead)
	readEOF := false
	destDir := path.Join(baseStruct.RootPath, "html")
	clearDirectory(destDir)
	for !readEOF {
		hdr, err := tarRead.Next()
		switch {
		case err == io.EOF:
			readEOF = true
			break
		case err != nil:
			readEOF = true
			break
		case hdr == nil:
			continue
		}
		if hdr == nil {
			continue
		}
		// 获取文件或目录的路径
		filePath := filepath.Join(destDir, hdr.Name)

		// 跳过顶级目录之前的文件
		if !strings.HasPrefix(hdr.Name, "dist/") {
			continue
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			// 找到dist目录，将这个目录下的文件夹在baseStruct.RootPath下的html文件夹下创建出来
			err := os.MkdirAll(filePath, 0755)
			if err != nil {
				log.ErrorLog.Println("上传前端文件。创建文件夹出错")
				log.ErrorLog.Println(err)
				continue
			}
		case tar.TypeReg:
			// 创建文件
			staticFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				log.ErrorLog.Println("上传前端文件。创建文件出错")
				log.ErrorLog.Println(err)
				continue
			}
			// 将文件内容拷贝到目标文件
			if _, err := io.Copy(staticFile, tarRead); err != nil {
				log.ErrorLog.Println("上传前端文件。写入文件出错")
				log.ErrorLog.Println(err)
			}
			staticFile.Close()
		}
	}
	clearDirectory(config.WebSiteStaticFolderPath)
	moveFolder(path.Join(destDir, "dist"), config.WebSiteStaticFolderPath)
	gtx.String(200, "deploy ok")
}

func moveFolder(sourceDir, targetDir string) error {
	// 遍历源目录下的所有文件和目录
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 构建目标路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, relPath)

		// 如果是目录，创建对应的目标目录
		if info.IsDir() {
			err := os.MkdirAll(targetPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			// 如果是文件，复制到目标路径并覆盖已存在的文件
			sourceFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer sourceFile.Close()

			targetFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer targetFile.Close()

			_, err = io.Copy(targetFile, sourceFile)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func clearDirectory(dirPath string) error {
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	for _, fileInfo := range fileInfos {
		filePath := filepath.Join(dirPath, fileInfo.Name())

		err = os.RemoveAll(filePath)
		if err != nil {
			return err
		}
	}

	return nil
}
