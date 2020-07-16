package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"regexp"
)

var (
	baseurl  = "http://test.hexiefamily.xin/static/static.tar.gz"
	basePath = "/Users/cj/Downloads/"
	logger   *log.Logger
)

func init(){
	fmt.Println("创建日记录日志文件")
	f,err:=os.OpenFile("Log.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 777)
	if err!=nil{
		log.Fatal("os.OpenFile err",err)
	}
	writers := []io.Writer{f,os.Stdout}
	//defer f.Close()  因为这里，就已经将文件关闭了，看来不能够随便使用defer
	fileAndStdoutWriter := io.MultiWriter(writers...)
	logger = log.New(fileAndStdoutWriter, "", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println("---> logger:check to make sure is works")
}
	
	
func main(){
	logger.Println("开始计算时间")
	t1 :=time.Now()
	HandleFile(baseurl)
	t2 :=time.Now()
	logger.Println("下载所有的文件总耗时：",t2.Sub(t1))
}
	

//walkFunc遍历指定路径下的所有文件，filepath.Walk("路径",walkFunc)
func walkFunc(path string,info os.FileInfo,err error)error{
	fmt.Println(path)
	return nil
}

//HandleFile is a function
//处理文件信息
func HandleFile(url string)error{
	if url==""{
		err:=fmt.Errorf("url是空的")
		logger.Println(err)
		return err
	}
	logger.Println("发起get请求")
	htmlData,err:= HttpGet(url)
	if htmlData==""{
		err:=fmt.Errorf("HttpGet中有错啦，请检查HttpGet数据")
		logger.Println(err)
		return err
	}
	if err!=nil{
		logger.Println(err)
		return err
	}
	cutUrl := strings.Split(url,baseurl)[1]
	logger.Println("正则匹配信息")
	re:=regexp.MustCompile(`<a href="(.*?)">(.*?)</a>`)
	result := re.FindAllStringSubmatch(htmlData,-1)
	if result==nil{
		err:=fmt.Errorf("正则匹配的数据为空")
		logger.Println(err)
		return err
	}
	logger.Println("正则匹配数据")
	for i:=0;i<len(result);i++{
		publicPath := cutUrl
		if len(result[i])!=3{
			err:=fmt.Errorf("正则出来的东西不是我想要的")
			logger.Println(err)
			return err
		}
		fmt.Println(result[i][2])
		if strings.Contains(result[i][2],"/"){
			logger.Println(result[i][2]+"是个文件夹")
			floderName:=strings.Replace(result[i][2],"/","",-1)
			if floderName!=result[i][2]{
				publicPath = publicPath+"/"+floderName
				fmt.Println("publicPath=============>",publicPath)
				logger.Println("创建文件夹")
				err=CreateFloder(publicPath)
				if err!=nil{
					logger.Println("CreateFloder error",err)
					return err
				}
				urlPath := baseurl+publicPath+"/"
				fmt.Println("urlPath============>",urlPath)
				logger.Println("递归")
				HandleFile(urlPath)
			}
		}else{
			publicPath = publicPath+result[i][2]
			logger.Println("遇到是文件，就将数据写入文件")
			err:=WriteFile(publicPath)
			if err!=nil{
				logger.Println(err)
				return err
			}
		}
	}
	return nil
}

//HttpGet 发起get请求
func HttpGet(url string) (string, error) {
	if url == "" {
		err := fmt.Errorf("传入的url为空")
		logger.Println(err)
		return "", err
	}
	if !strings.Contains(url, "http:") {
		err := fmt.Errorf("传入的url不正确")
		logger.Println(err)
		return "", err
	}
	resp, err := http.Get(url)
	if err != nil {
		err1 := fmt.Errorf("http.Get error===========>%v", err)
		logger.Println(err1)
		return "", err1
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err1 := fmt.Errorf("ioutil.ReadAll error===========>%v", err)
		logger.Println(err1)
		return "", err1
	}
	if strings.Contains(string(body), "404 page not found") {
		err := fmt.Errorf("找不到该网页")
		logger.Println(err)
		return "", err
	}
	return string(body), nil
}

//IsExist 判断文件是否存在
func IsExist(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		} else {
			return false
		}
	}
	return true
}


//CreateFloder 创建文件夹
func CreateFloder(publicPath string)(error){
	if publicPath==""{
		err:=fmt.Errorf("传入urlpath的地址空")
		logger.Println(err)
		return err
	}
	logger.Println(publicPath,"创建文件夹")
	err:=os.MkdirAll(basePath+"//"+publicPath,os.ModePerm)
	if err!=nil{
		err:=fmt.Errorf("创建文件错误啦:%v",err)
		logger.Println(err)
		return err
	}
	return nil
}

//WriteFile 写入文件
func WriteFile(publicPath string)error{
	if publicPath==""{
		err:=fmt.Errorf("传入的参数为空，请注意！！！")
		logger.Println(err)
		return err
	}
	htmlData,err:=HttpGet(baseurl+publicPath)
	if htmlData==""{
		err:=fmt.Errorf("HttpGet中有错啦，请检查HttpGet数据")
		logger.Println(err)
		return err
	}
	if err!=nil{
		logger.Println(err)
		return err
	}
	file,err:=os.Create(basePath+publicPath)
	defer file.Close()
	if err!=nil{
		err:=fmt.Errorf("os.Open失败,err=%v",err)
		logger.Println(err)
		return err
	}
	if file==nil{
		err:=fmt.Errorf("创建文件失败")
		logger.Println(err)
		return err
	}
	file.WriteString(htmlData)
	return nil
}