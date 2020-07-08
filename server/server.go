package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"qrcp_pass/payload"
	"runtime"
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
