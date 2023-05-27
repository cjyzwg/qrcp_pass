package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"qrcp_pass/payload"
	"runtime"
	"strconv"
	"time"

	"github.com/zserge/lorca"
)

// Server is a struct
type Server struct {
	Instance               *http.Server
	Stopchannel            chan bool
	AutoCloseUi            chan int
	Signalchannel          chan os.Signal
	Payload                payload.Payload
	ExpectParallelRequests bool
	Port                   string
	IsUpload               bool
	Uploaddir              string
}

// Urlparms is a struct
type Urlparms struct {
	Sendip      string
	Sendurl     string
	Receiveurl  string
	Checkupload bool
}

// Send adds a handler for sending the file
func (s *Server) Send(p payload.Payload) {
	s.Payload = p
	s.ExpectParallelRequests = true
}

// Open 目录
func Open(url string) error {
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

// Wait for transfer to be completed, it waits forever if kept awlive 拷贝
func (s *Server) Wait() {
	<-s.Stopchannel
	log.Println("检测到传输完成的信号")
	//DeleteAfterTransfer
	if s.Payload.DeleteAfterTransfer {
		s.Payload.Delete()
	}
	//open upload folder
	if s.IsUpload {
		Open(s.Uploaddir)
	}
	close(s.AutoCloseUi)

}

// ExecUI is a function
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
	// args = append(args, "--start-fullscreen")       // 起動時最大化（追加）
	args = append(args, "--remote-allow-origins=*") // websocket.Dial bad status 回避问题（追加）

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
	// fmt.Println(s.Uistopchannel)
	// sigc := make(chan os.Signal)
	// signal.Notify(sigc, os.Interrupt)
	select {
	// case <-sigc:
	case <-s.Signalchannel:
		log.Println("监听到signal channel down")
	case <-ui.Done():
		log.Println("手动关闭ui close")
	case <-s.AutoCloseUi:
		ui.Close()
		log.Println("传输完成，关闭ui")
	}
	// Close UI
	log.Println("exiting...")
	os.Exit(1)

}

// IndexTmpl is a function
func (url *Urlparms) IndexTmpl(w http.ResponseWriter, r *http.Request) {

	type IndexParms struct {
		Path        string
		Translation string
	}

	parms := &IndexParms{
		Path:        "download",
		Translation: "传文件",
	}
	if url.Checkupload {
		parms.Path = "upload"
		parms.Translation = "收文件"
	}

	t1, err := template.ParseFiles("template/index.html")
	if err != nil {
		panic(err)
	}
	t1.Execute(w, parms)

	// sendurl := url.Sendurl
	// f, err := qrcode.Encode(sendurl, qrcode.Highest, 300)
	// if err != nil {
	// 	log.Println(err.Error())
	// 	return
	// }
	// w.Write(f)
}

// QrcodeTmpl is a function
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

// UploadTmpl is a function
func (url *Urlparms) UploadTmpl(w http.ResponseWriter, r *http.Request) {
	t1, err := template.ParseFiles("template/page/upload.html")
	if err != nil {
		panic(err)
	}
	t1.Execute(w, nil)
}

// LayDemoTmpl is a function
func (url *Urlparms) LayDemoTmpl(w http.ResponseWriter, r *http.Request) {
	t1, err := template.ParseFiles("template/page/laydemo.html")
	if err != nil {
		panic(err)
	}
	t1.Execute(w, nil)
}

// OnSip is a function
func (url *Urlparms) OnSip(res http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	// log.Println("query", query)
	path := query["path"][0]
	// log.Println("path", path)
	ipmap := make(map[string]string)
	if path == "download" {
		ipmap["ip"] = url.Sendurl
	} else {
		ipmap["ip"] = url.Receiveurl
	}
	jsonips, _ := json.Marshal(ipmap)
	//返回的这个是给json用的，需要去掉
	res.Header().Set("Content-Length", strconv.Itoa(len(jsonips)))
	io.WriteString(res, string(jsonips))
}

func DownLoadFile(s *Server, w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Disposition", "attachment; filename="+app.Payload.Filename)
	// log.Println(filePath)
	// http.ServeFile(w, r, filePath)

	//大文件传输
	// file, err := os.Open(filePath)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// defer file.Close()

	// fileInfo, err := file.Stat()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// fileSize := fileInfo.Size()
	// // Create a new progress bar
	// progressBar := pb.New64(fileSize).SetUnits(pb.U_BYTES)
	// progressBar.Start()
	// defer progressBar.Finish()

	// // Create a new writer that updates the progress bar while writing
	// writer := progressBar.NewProxyWriter(w)

	// // Copy the file data to the writer with progress bar updates
	// _, err = io.Copy(writer, file)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// // Set the content type and size
	// w.Header().Set("Content-Type", "application/octet-stream")
	// w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	// // Serve the file
	// http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
	// close(ch)

}

func (s *Server) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Open the file
	file, err := os.Open(s.Payload.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set the content type and attachment header
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(s.Payload.Filename))

	// Get the file size
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fileSize := fileInfo.Size()

	// Set the content length header
	w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))

	// Create a buffer to store the file chunks
	buffer := make([]byte, 1024*1024) // 1MB chunks
	var bytesRead int64

	var percent float64
	// Stream the file to the client
	for {
		// Read a chunk from the file
		n, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			break
		}

		// Write the chunk to the response writer
		if n > 0 {
			_, err := w.Write(buffer[:n])
			if err != nil {
				return
			}

			// Flush the buffer
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}

			// Update the number of bytes read so far
			bytesRead += int64(n)

			// Calculate the percentage downloaded
			if fileSize > 0 {
				percent = float64(bytesRead) / float64(fileSize) * 100
				fmt.Printf("%.0f%% downloaded\n", percent)
			}

		}

		// Check if we have read the entire file
		if bytesRead >= fileSize {
			log.Println(bytesRead)

			close(s.Stopchannel)
			break
		}

	}

}
