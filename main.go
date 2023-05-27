package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"qrcp_pass/payload"
	"qrcp_pass/server"
	"qrcp_pass/util"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/c4milo/unpackit"
	"github.com/zserge/lorca"
)

// SIGHUP	1	Term	终端控制进程结束(终端连接断开)
// SIGINT	2	Term	用户发送INTR字符(Ctrl+C)触发
// SIGQUIT	3	Core	用户发送QUIT字符(Ctrl+/)触发
// SIGILL	4	Core	非法指令(程序错误、试图执行数据段、栈溢出等)
// SIGABRT	6	Core	调用abort函数触发
// SIGFPE	8	Core	算术运行错误(浮点运算错误、除数为零等)
// SIGKILL	9	Term	无条件结束程序(不能被捕获、阻塞或忽略)
// SIGSEGV	11	Core	无效内存引用(试图访问不属于自己的内存空间、对只读内存空间进行写操作)
// SIGPIPE	13	Term	消息管道损坏(FIFO/Socket通信时，管道未打开而进行写操作)
// SIGALRM	14	Term	时钟定时信号
// SIGTERM	15	Term	结束程序(可以被捕获、阻塞或忽略)
// SIGUSR1	30,10,16	Term	用户保留
// SIGUSR2	31,12,17	Term	用户保留
// SIGCHLD	20,17,18	Ign	子进程结束(由父进程接收)
// SIGCONT	19,18,25	Cont	继续执行已经停止的进程(不能被阻塞)
// SIGSTOP	17,19,23	Stop	停止进程(不能被捕获、阻塞或忽略)
// SIGTSTP	18,20,24	Stop	停止进程(可以被捕获、阻塞或忽略)
// SIGTTIN	21,21,26	Stop	后台程序从终端中读取数据时触发
// SIGTTOU	22,22,27	Stop	后台程序向终端中写数据时触发
// var xport = "21346"
// later change it to config.toml
var (
	downloadDir = "./download/"
	uploadDir   = "./upload/"
	gzurl       = "https://gitee.com/cjyzwg/qrcp_pass/raw/qrcp_static/static.tar.gz"
	unpackDir   = "./"
	defaultFile = "README.md"
)

func main() {
	//first unpack file and check defaultfile can not be deleted
	existed, _ := util.PathExists(downloadDir + defaultFile)
	if !existed {
		url := gzurl
		tempDir := unpackDir
		res, err := http.Get(url)
		if err != nil {
			fmt.Println("this url is not existed", err)
			panic(err)
		}
		xerr := unpackit.Unpack(res.Body, tempDir)
		if xerr != nil {
			fmt.Println("this decompress got wrong", xerr)
			panic(xerr)
		}
		fmt.Println("unpack is ok now")
	}

	//get data from standard input stream
	input := bufio.NewScanner(os.Stdin)
	var lastline string

	fmt.Println(`************************************************************
------------------------------------------------------------`)
	fmt.Printf(` 
	 $$$$$$\     $$$$$\                                         
	$$  __$$\    \__$$ |                                        
	$$ /  \__|      $$ | $$$$$$\   $$$$$$\   $$$$$$$\  $$$$$$$\ 
	$$ |            $$ |$$  __$$\  \____$$\ $$  _____|$$  _____|
	$$ |      $$\   $$ |$$ /  $$ | $$$$$$$ |\$$$$$$\  \$$$$$$\  
	$$ |  $$\ $$ |  $$ |$$ |  $$ |$$  __$$ | \____$$\  \____$$\ 
	\$$$$$$  |\$$$$$$  |$$$$$$$  |\$$$$$$$ |$$$$$$$  |$$$$$$$  |
	 \______/  \______/ $$  ____/  \_______|\_______/ \_______/ 
			    $$ |                                    
			    $$ |                                    
			    \__|                                    `)
	fmt.Println("")
	fmt.Println(`------------------------------------------------------------`)
	fmt.Println(" 传输(CJPass) v0.0.1  手机电脑文件传输 made by cj")
	fmt.Println(`------------------------------------------------------------
************************************************************`)
	fmt.Printf("请选择以下哪种方式（输入1或2）:\n")
	fmt.Printf("扫码传文件【1】:\n")
	fmt.Printf("扫码收文件【2】:\n")
	//only get one file
	fileExt := defaultFile
	opendownloaddir := downloadDir
	if runtime.GOOS == "windows" {
		opendownloaddir = strings.Replace(downloadDir, "/", "\\", -1)
	}
	// 逐行扫描
	for input.Scan() {
		line := input.Text()

		//upload file check file is not needed
		if line == "2" {
			lastline = line
			break
		} else {

			//download file need to check
			downloadfiles, err := ioutil.ReadDir(downloadDir)
			if err != nil {
				panic(err)
			}

			for _, downloadfile := range downloadfiles {
				if downloadfile.Name() != defaultFile {
					fileExt = downloadfile.Name()
				}
			}
			if strings.Index(strings.Replace(fileExt, " ", "", -1), "README.md") <= -1 {
				//have another file then can break
				lastline = line
				fmt.Printf("文件后缀名为:" + fileExt + ",开始打开网页:\n")
				break
			} else {
				//open download folder
				//tell user you need to add file to the download folder
				server.Open(opendownloaddir)
				fmt.Printf("已经打开" + downloadDir + "目录下:\n")
				fmt.Printf("请先放要传输的文件放到" + downloadDir + "目录下:\n")
				fmt.Printf("请选择以下哪种方式（输入1或2）:\n")
				fmt.Printf("传文件【1】:\n")
				fmt.Printf("收文件【2】:\n")
			}

		}

	}
	downloadfileext := downloadDir + fileExt
	var ars []string
	ars = append(ars, downloadfileext)
	// log.Println("ars is :", ars)

	//before from args we use this
	// ars := os.Args[1:]

	// var send bool
	// if len(ars) == 0 {
	// 	ars = append(ars, "xxx")
	// 	send = true
	// } else {
	// 	send = false
	// }
	payload, _ := payload.FromArgs(ars)

	// log.Println(payload)
	// return
	randomport := RangeRand(0, 40000)
	xport := strconv.FormatInt(randomport, 10)
	log.Println("Port is:", xport)
	app := &server.Server{}
	if lastline == "2" {
		app.IsUpload = true
		app.Uploaddir = uploadDir
		if runtime.GOOS == "windows" {
			app.Uploaddir = strings.Replace(uploadDir, "/", "\\", -1)
		}
	}
	// ip, _ := GetLocalIP()
	ip, iperr := util.GetIp()
	if iperr != nil {
		panic(iperr)
	}
	if ip == "" {
		fmt.Println("没有发现ip，请先连接网络再尝试")
		return
	}

	app.Port = xport
	//优雅关闭,非nil，初始化不会引发panic
	app.Stopchannel = make(chan bool)
	app.AutoCloseUi = make(chan int)
	app.Signalchannel = make(chan os.Signal)
	signal.Notify(app.Signalchannel, os.Interrupt)

	port := xport
	newaddr := ip + ":" + port
	log.Println("Current ip+port is:", newaddr)
	sendurl := "http://" + newaddr + "/send/sea/"
	receiveurl := "http://" + newaddr + "/upload"
	urlparms := &server.Urlparms{
		Sendip:     ip,
		Sendurl:    sendurl,
		Receiveurl: receiveurl,
	}
	if lastline == "2" {
		urlparms.Checkupload = true
	}
	http.HandleFunc("/send/sea/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("downloading")

		app.DownloadHandler(w, r)

	})
	// outputdir := "/Users/cj/Downloads"
	outputdir := uploadDir
	//outputdir static should be the current one
	//receive file
	http.HandleFunc("/upload/sea/", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "POST":
			// defer waitgroup.Done()
			filenames := util.ReadFilenames(outputdir)
			reader, err := r.MultipartReader()
			if err != nil {
				log.Println("Upload failed", err)
				app.Stopchannel <- true
				return
			}

			transferedfiles := []string{}

			for {
				// waitgroup.Add(1)
				part, err := reader.NextPart()
				if err == io.EOF {
					break
				}
				//filename is empty skip
				if part.FileName() == "" {
					continue
				}
				// n++
				// waitgroup.Add(1)
				//prepare destination filename
				fileName := util.GetFileName(part.FileName(), filenames)

				out, err := os.Create(filepath.Join(outputdir, fileName))
				if err != nil {
					log.Println("unable create file path", err)
					app.Stopchannel <- true
					return
				}
				defer out.Close()
				//Add name info filename
				filenames = append(filenames, fileName)

				// log.Println("output filename is :", out.Name())

				/*****************************************/
				// waitgroup.Add(1)
				//start read and write chunk
				//create a buf

				buf := make([]byte, 4096)
				for {
					//read a chunk
					b, err := part.Read(buf)
					if err != nil && err != io.EOF {
						log.Println("can not read file into disk", err)
						app.Stopchannel <- true
						return
					}
					//this part already finished
					if b == 0 {
						break
					}
					//write into a chunk
					if _, err := out.Write(buf[:b]); err != nil {
						log.Println("can not write file into disk", err)
						app.Stopchannel <- true
						return
					}

				}
				// go ReadBuff(&waitgroup, app, part, out)
				transferedfiles = append(transferedfiles, out.Name())

				//wait group problem
				// defer waitgroup.Done()
				/*****************************************/

			}
			// defer waitgroup.Done()
			// progressBar.FinishPrint("File transfer completed")
			// app.Stopchannel <- true
			//layui will call two request
			// WriteResponse(http.StatusAccepted, nil, w)
			WriteResponse(http.StatusAccepted, "get it", w)
		}
	})
	//later change it to be self router
	//load js,css static resources
	http.Handle("/static/css/", http.StripPrefix("/static/css/", http.FileServer(http.Dir("public/assets/css/"))))
	http.Handle("/static/js/", http.StripPrefix("/static/js/", http.FileServer(http.Dir("public/js/"))))
	http.Handle("/static/js/libs/", http.StripPrefix("/static/js/libs/", http.FileServer(http.Dir("public/js/libs/"))))
	http.Handle("/static/images/", http.StripPrefix("/static/images/", http.FileServer(http.Dir("public/assets/images/"))))
	http.Handle("/static/fonts/", http.StripPrefix("/static/fonts/", http.FileServer(http.Dir("public/assets/fonts/"))))

	//wait for index page
	http.HandleFunc("/", urlparms.IndexTmpl)
	//qrcode
	http.HandleFunc("/qrcode", urlparms.QrcodeTmpl)
	//upload
	http.HandleFunc("/upload", urlparms.UploadTmpl)
	//api sip get ip string
	http.HandleFunc("/api/sip", urlparms.OnSip)
	//layui demo
	http.HandleFunc("/homelay", urlparms.LayDemoTmpl)

	go func() {
		ChromeExe := lorca.ChromeExecutable()
		if ChromeExe != "" {
			//打开UI界面
			app.ExecUI()
		} else {
			addr := "http://127.0.0.1"
			server.Open(addr + ":" + port)
		}
	}()
	httpserver := &http.Server{Addr: ":" + port}

	app.Instance = httpserver

	app.Send(payload)
	go app.Wait()

	if err := httpserver.ListenAndServe(); err != nil {
		log.Printf("Httpserver: ListenAndServe() error: %s", err)
	}

}

// ReadBuff is a function
func ReadBuff(wg *sync.WaitGroup, s *server.Server, part *multipart.Part, out *os.File) {
	defer wg.Done()
	buf := make([]byte, 4096)
	for {
		//read a chunk
		b, err := part.Read(buf)
		if err != nil && err != io.EOF {
			log.Println("can not read file into disk", err)
			s.Stopchannel <- true
			return
		}
		//this part already finished
		if b == 0 {
			break
		}
		//write into a chunk
		if _, err := out.Write(buf[:b]); err != nil {
			log.Println("can not write file into disk", err)
			s.Stopchannel <- true
			return
		}

	}
}

// WriteResponse is a function
func WriteResponse(code int, jsonres interface{}, w http.ResponseWriter) {
	b, err := json.Marshal(jsonres)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write(b)
	}
}

// 打开系统默认浏览器

// GetSessionID returns a base64 encoded string of 40 random characters
func GetSessionID() (string, error) {
	randbytes := make([]byte, 40)
	if _, err := io.ReadFull(rand.Reader, randbytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(randbytes), nil
}

// GetLocalIP is a function
func GetLocalIP() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet
		IsIpNet bool
	)
	//获取所有网卡
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	//取第一个非lo的网卡IP
	for _, addr = range addrs {
		if ipNet, IsIpNet = addr.(*net.IPNet); IsIpNet && !ipNet.IP.IsLoopback() {
			//跳过ipv6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}
	return "", nil
}

// RangeRand 生成区间[-m, n]的安全随机数
func RangeRand(min, max int64) int64 {
	if min > max {
		panic("the min is greater than max!")
	}

	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, _ := rand.Int(rand.Reader, big.NewInt(max+1+i64Min))

		return result.Int64() - i64Min
	} else {
		result, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
		return min + result.Int64()
	}
}
