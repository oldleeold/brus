package gotools

import (
	"bytes"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Tang-RoseChild/mahonia"
)

func RandAuthToken() string {
	buf := make([]byte, 32)
	_, err := crand.Read(buf)
	if err != nil {
		return RandString(64)
	}

	return fmt.Sprintf("%x", buf)
}

// RandString 生成长度为length的随机字符串
func RandString(length int64) string {
	sources := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sourceLength := len(sources)
	var i int64 = 0
	for ; i < length; i++ {
		result = append(result, sources[r.Intn(sourceLength)])
	}

	return string(result)
}

// Md5 生成32位MD5摘要
func Md5(str string) string {
	m := md5.New()
	m.Write([]byte(str))

	return hex.EncodeToString(m.Sum(nil))
}

// RandNumber 生成0-max之间随机数
func RandNumber(max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return r.Intn(max)
}

// GBK2UTF8 GBK编码转换为UTF8
func GBK2UTF8(s string) (string, bool) {
	dec := mahonia.NewDecoder("gbk")

	return dec.ConvertStringOK(s)
}

// ReplaceStrings 批量替换字符串
func ReplaceStrings(s string, old []string, replace []string) string {
	if s == "" {
		return s
	}
	if len(old) != len(replace) {
		return s
	}

	for i, v := range old {
		s = strings.Replace(s, v, replace[i], 1000)
	}

	return s
}

func InStringSlice(slice []string, element string) bool {
	element = strings.TrimSpace(element)
	for _, v := range slice {
		if strings.TrimSpace(v) == element {
			return true
		}
	}

	return false
}

// EscapeJson 转义json特殊字符
func EscapeJson(s string) string {
	specialChars := []string{"\\", "\b", "\f", "\n", "\r", "\t", "\""}
	replaceChars := []string{"\\\\", "\\b", "\\f", "\\n", "\\r", "\\t", "\\\""}

	return ReplaceStrings(s, specialChars, replaceChars)
}

// FileExist 判断文件是否存在及是否有权限访问
func FileExist(file string) bool {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	if os.IsPermission(err) {
		return false
	}

	return true
}

// PanicToError Panic转换为error
func PanicToError(f func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf(PanicTrace(e))
		}
	}()
	f()
	return
}

// PanicTrace panic调用链跟踪
func PanicTrace(err interface{}) string {
	stackBuf := make([]byte, 4096)
	n := runtime.Stack(stackBuf, false)

	return fmt.Sprintf("panic: %v %s", err, stackBuf[:n])
}

// PrintAppVersion 打印应用版本
func PrintAppVersion(appVersion, GitCommit, BuildDate string) {
	versionInfo, err := FormatAppVersion(appVersion, GitCommit, BuildDate)
	if err != nil {
		panic(err)
	}
	fmt.Println(versionInfo)
}

// FormatAppVersion 格式化应用版本信息
func FormatAppVersion(appVersion, GitCommit, BuildDate string) (string, error) {
	content := `
   Version: {{.Version}}
Go Version: {{.GoVersion}}
Git Commit: {{.GitCommit}}
     Built: {{.BuildDate}}
   OS/ARCH: {{.GOOS}}/{{.GOARCH}}
`
	tpl, err := template.New("version").Parse(content)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, map[string]string{
		"Version":   appVersion,
		"GoVersion": runtime.Version(),
		"GitCommit": GitCommit,
		"BuildDate": BuildDate,
		"GOOS":      runtime.GOOS,
		"GOARCH":    runtime.GOARCH,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), err
}

// DownloadFile 文件下载
func DownloadFile(filePath string, rw http.ResponseWriter) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	filename := path.Base(filePath)
	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	_, err = io.Copy(rw, file)

	return err
}

// WorkDir 获取程序运行时根目录
func WorkDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	wd := filepath.Dir(execPath)
	if filepath.Base(wd) == "bin" {
		wd = filepath.Dir(wd)
	}

	return wd, nil
}

// WaitGroupWrapper waitGroup包装
type WaitGroupWrapper struct {
	sync.WaitGroup
}

// Wrap 包装Add, Done方法
func (w *WaitGroupWrapper) Wrap(f func()) {
	w.Add(1)
	go func() {
		defer w.Done()
		f()
	}()

}
