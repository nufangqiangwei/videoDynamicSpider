package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"time"
	"videoDynamicAcquisition/log"
)

type WriteFile struct {
	FolderPrefix     []string
	FileNamePrefix   string
	FileName         func(string) string
	file             *os.File
	writeNumber      int
	lastOpenFileName string
}

func (wf *WriteFile) getFileName(newFile bool) string {
	if wf.FileName != nil {
		return wf.FileName(wf.lastOpenFileName)
	}
	if wf.lastOpenFileName == "" || newFile {
		return fmt.Sprintf("%s-%s.json", wf.FileNamePrefix, time.Now().Format("2006-01-02-15-04-05"))
	}
	return wf.lastOpenFileName

}
func (wf *WriteFile) checkFileSize() {
	if wf.file == nil {
		filePath := append(wf.FolderPrefix, wf.getFileName(true))
		f, err := os.OpenFile(path.Join(filePath...), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.ErrorLog.Printf("打开新文件失败%s", err.Error())
			panic(err)
		}
		log.Info.Println(path.Join(filePath...))
		wf.file = f
	}
	for {
		fi, err := wf.file.Stat()
		if err != nil {
			log.ErrorLog.Printf("获取文件信息失败%s", err.Error())
			panic(err)
		}
		if fi.Size() >= maxFileSize {
			wf.file.Close()
			filePath := append(wf.FolderPrefix, wf.getFileName(true))
			f, err := os.OpenFile(path.Join(filePath...), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.ErrorLog.Printf("打开新文件失败%s", err.Error())
				panic(err)
			}
			wf.file = f
			wf.writeNumber = 0
		} else {
			break
		}
	}
}
func (wf *WriteFile) Write(data []byte) (int, error) {
	if wf.file == nil {
		wf.checkFileSize()
	}
	// 每写入两千行就检查下文件大小
	if wf.writeNumber%2000 == 0 {
		wf.checkFileSize()
	}
	wf.writeNumber++
	return wf.file.Write(data)
}
func (wf *WriteFile) WriteLine(data []byte) (int, error) {
	a, b := wf.Write(data)
	if b != nil {
		return a, b
	}
	return wf.Write([]byte{10})
}
func (wf *WriteFile) Close() {
	if wf.file == nil {
		return
	}
	wf.lastOpenFileName = wf.file.Name()
	wf.file.Close()
	wf.file = nil
}

func NewReaderJSONFile(rd io.Reader) ReadJsonFile {
	rf := ReadJsonFile{}
	rf.readObject = rd
	rf.cache = []byte{}
	return rf
}

type ReadJsonFile struct {
	readObject io.Reader
	cache      []byte
}

const (
	leftCurlyBrace       = '{'
	rightCurlyBrace      = '}'
	doubleQuotes    byte = 34 // "
	escapes         byte = 92 // \
)

// Line 读取一个完整的json对象，从123->{字符读到125->}字符。两边的字符必须是对称出现，返回这样的一个json字符串
func (ro *ReadJsonFile) Line() ([]byte, int, error) {
	var (
		buf      bytes.Buffer
		lastByte byte
	)
	started := false
	count := 0
	if len(ro.cache) > 0 {
		buf.Write(ro.cache)
		ro.cache = []byte{}
	}
	inStr := false
	lastByte = 0
	for {
		b := make([]byte, 1)
		_, err := ro.readObject.Read(b)
		if err != nil {
			if err == io.EOF {
				return buf.Bytes(), buf.Len(), io.EOF
			}
			return nil, 0, err
		}
		// 判断当前是否在一个字符串当中，如何当前在字符串当中，对括号的计算需要排除。 92->\ 转义符 34->"双引号
		if b[0] == doubleQuotes {
			// 查看前一个是否是转义符
			if lastByte != escapes {
				if inStr {
					inStr = false
				} else {
					inStr = true
				}
			}

		}

		if b[0] == leftCurlyBrace && !inStr {
			started = true
			count++
		}

		if started {
			buf.Write(b)
		}

		if b[0] == rightCurlyBrace && !inStr {
			count--
			if count == 0 {
				break
			}
		}

		lastByte = b[0]
	}

	return buf.Bytes(), buf.Len(), nil
}
