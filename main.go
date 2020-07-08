package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"qrcp_pass/payload"
	"qrcp_pass/server"
	"qrcp_pass/util"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/skip2/go-qrcode"
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

func main() {

	payload, _ := payload.FromArgs(os.Args[1:])
	// fmt.Println(payload)
	randomport := RangeRand(0, 40000)
	xport := strconv.FormatInt(randomport, 10)
	log.Println("Port is:", xport)
	app := &server.Server{}
	// ip, _ := GetLocalIP()
	ips, iperr := util.Ips()
	if iperr != nil {
		panic(iperr)
	}
	if len(ips) == 0 {
		fmt.Println("没有发现ip，请先连接网络再尝试")
		return
	}
	ip := ips[0]
	log.Println("Current ip is:", ip)
	app.Port = xport
	//优雅关闭
	app.Stopchannel = make(chan bool)
	app.Uistopchannel = make(chan bool)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		app.Stopchannel <- true
	}()
	// Create cookie used to verify request is coming from first client to connect
	cookie := &http.Cookie{
		Name:  "qrcp",
		Value: "",
	}
	port := xport
	newaddr := ip + ":" + port
	sendurl := "http://" + newaddr + "/send/sea/"
	var waitgroup sync.WaitGroup
	waitgroup.Add(1)
	var initCookie sync.Once
	http.HandleFunc("/send/sea/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("1111")
		if cookie.Value == "" {
			if !strings.HasPrefix(r.Header.Get("User-Agent"), "Mozilla") {
				http.Error(w, "", http.StatusOK)
				return
			}
			initCookie.Do(func() {
				value, err := GetSessionID()
				if err != nil {
					log.Print("unable to get sessionid", err)
					app.Stopchannel <- true
					return
				}
				cookie.Value = value
				http.SetCookie(w, cookie)
			})
		} else {
			rcookie, err := r.Cookie(cookie.Name)
			if err != nil || rcookie.Value != cookie.Value {
				http.Error(w, "", http.StatusNotFound)
				return
			}
			// if cookie exists and add
			waitgroup.Add(1)
		}
		//remove connnection when waitgroup done
		defer waitgroup.Done()
		w.Header().Set("Content-Disposition", "attachment; filename="+app.Payload.Filename)
		http.ServeFile(w, r, app.Payload.Path)

	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("222")
		f, err := qrcode.Encode(sendurl, qrcode.Highest, 300)
		if err != nil {
			log.Println(err.Error())
			return
		}
		w.Write(f)
	})
	//wait for all wait done
	go func() {
		waitgroup.Wait()
		app.Stopchannel <- true
	}()

	httpserver := &http.Server{Addr: ":" + port}
	// listener, err := net.Listen("tcp", newaddr)
	// if err != nil {
	// 	log.Fatalln("error get ")
	// }
	//go open 必须在有网的情况下才能调取成功
	addr := "http://127.0.0.1"
	go func() {

		// open(addr + ":" + port)
		if err := httpserver.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	go func() {
		ChromeExe := lorca.ChromeExecutable()
		if ChromeExe != "" {
			//打开UI界面
			app.ExecUI()
		} else {
			open(addr + ":" + port)
		}
	}()

	// go func() {
	// 	if err := httpserver.ListenAndServe(); err != nil {
	// 		// cannot panic, because this probably is an intentional close
	// 		log.Printf("Httpserver: ListenAndServe() error: %s", err)
	// 	}
	// 	// if err := (httpserver.Serve(server.TcpKeepAliveListener{listener.(*net.TCPListener)})); err != http.ErrServerClosed {
	// 	// 	log.Fatalln(err)
	// 	// }
	// }()

	app.Instance = httpserver

	// qr.RenderString(sendurl)

	app.Send(payload)
	xerr := app.Wait()
	if xerr != nil {
		log.Fatalln("error is :", xerr)
	}

}

// 打开系统默认浏览器

// 目录
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// GetSessionID returns a base64 encoded string of 40 random characters
func GetSessionID() (string, error) {
	randbytes := make([]byte, 40)
	if _, err := io.ReadFull(rand.Reader, randbytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(randbytes), nil
}

//GetLocalIP is a function
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

//RangeRand 生成区间[-m, n]的安全随机数
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
