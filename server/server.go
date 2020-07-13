package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"qrcp_pass/payload"
	"runtime"
	"strconv"
	"time"

	"github.com/zserge/lorca"
)

//Server is a struct
type Server struct {
	Instance               *http.Server
	Stopchannel            chan bool
	Uistopchannel          chan bool
	Payload                payload.Payload
	ExpectParallelRequests bool
	Port                   string
}

//Urlparms is a struct
type Urlparms struct {
	Sendip  string
	Sendurl string
}

// Send adds a handler for sending the file
func (s *Server) Send(p payload.Payload) {
	s.Payload = p
	s.ExpectParallelRequests = true
}

// Wait for transfer to be completed, it waits forever if kept awlive
func (s *Server) Wait() error {
	s.Uistopchannel = make(chan bool)
	<-s.Stopchannel
	s.Uistopchannel <- true
	//这里出错了
	if err := s.Instance.Shutdown(context.Background()); err != nil {
		// fmt.Println("xxxxxxxxxxxx")
		log.Println(err)
	}
	//这里出错了
	if s.Payload.DeleteAfterTransfer {
		s.Payload.Delete()
	}
	return nil
}

//ExecUI is a function
func (s *Server) ExecUI() {
	// Wait Server Run
	time.Sleep(3 * time.Second)

	// Cli Args
	var args []string
	if runtime.GOOS == "linux" {
		args = append(args, "--class=Lorca")
	}
	if runtime.GOOS == "windows" {
		args = append(args, "-ldflags '-H windowsgui'")
	}

	// New Lorca UI
	ui, err := lorca.New(
		`data:text/html,
		<html><head><title>Lovesrr</title></head></html>`,
		"", 360, 640, args...,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = ui.Close()
	}()

	// Load url
	_ = ui.Load(fmt.Sprintf(
		"http://%s",
		"127.0.0.1:"+s.Port),
	)

	// Wait until the interrupt signal arrives
	// or browser window is closed
	// sigc := make(chan os.Signal)
	// signal.Notify(sigc, os.Interrupt)
	// select {
	// case <-sigc:
	// case <-ui.Done():
	// }
	fmt.Println(s.Uistopchannel)
	select {
	case <-s.Uistopchannel:
	case <-ui.Done():
	}
	// Close UI
	log.Println("exiting...")
}

//IndexTmpl is a function
func (url *Urlparms) IndexTmpl(w http.ResponseWriter, r *http.Request) {
	t1, err := template.ParseFiles("template/test.html")
	if err != nil {
		panic(err)
	}
	t1.Execute(w, nil)

	// sendurl := url.Sendurl
	// f, err := qrcode.Encode(sendurl, qrcode.Highest, 300)
	// if err != nil {
	// 	log.Println(err.Error())
	// 	return
	// }
	// w.Write(f)
}

//QrcodeTmpl is a function
func (url *Urlparms) QrcodeTmpl(w http.ResponseWriter, r *http.Request) {
	t1, err := template.ParseFiles("template/page/qrcode.html")
	if err != nil {
		panic(err)
	}
	t1.Execute(w, nil)

	// sendurl := url.Sendurl
	// f, err := qrcode.Encode(sendurl, qrcode.Highest, 300)
	// if err != nil {
	// 	log.Println(err.Error())
	// 	return
	// }
	// w.Write(f)
}

//OnSip is a function
func (url *Urlparms) OnSip(res http.ResponseWriter, req *http.Request) {

	ipmap := make(map[string]string)
	ipmap["ip"] = url.Sendurl
	jsonips, _ := json.Marshal(ipmap)
	//返回的这个是给json用的，需要去掉
	res.Header().Set("Content-Length", strconv.Itoa(len(jsonips)))
	io.WriteString(res, string(jsonips))
}
