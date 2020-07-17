package util

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//Ips 得到网卡名加ip
func Ips() ([]string, error) {
	//过滤掉这些网卡
	var re = regexp.MustCompile(`^(veth|br\-|docker|lo|EHC|XHC|bridge|gif|stf|p2p|awdl|utun|tun|tap|VirtualBox)`)
	//获取192开头的
	var ipre = regexp.MustCompile(`^(192)`)
	var ips []string
	// ips := make(map[string]string)

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		//byName 是网卡
		byName, err := net.InterfaceByName(i.Name)
		if err != nil {
			return nil, err
		}
		if re.MatchString(byName.Name) {
			continue
		}
		addresses, err := byName.Addrs()
		for _, v := range addresses {
			//去除本机地址 以及不是ipv4的地址
			if ipnet, ok := v.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ipstring := ipnet.IP.String()
				if ipre.MatchString(ipstring) {
					ips = append(ips, ipnet.IP.String())
				}
			}

		}
	}
	return ips, nil
}

//GetIfaceIps 得到网卡名加ip
func GetIfaceIps() (map[string]string, error) {
	//过滤掉这些网卡
	var re = regexp.MustCompile(`^(veth|br\-|docker|lo|EHC|XHC|bridge|gif|stf|p2p|awdl|utun|tun|tap|VirtualBox)`)
	//获取192开头的
	var ipre = regexp.MustCompile(`^(192)`)
	ips := make(map[string]string)

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		//byName 是网卡
		byName, err := net.InterfaceByName(i.Name)
		if err != nil {
			return nil, err
		}
		if re.MatchString(byName.Name) {
			continue
		}
		addresses, err := byName.Addrs()
		for _, v := range addresses {
			//去除本机地址 以及不是ipv4的地址
			if ipnet, ok := v.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ipstrig := ipnet.IP.String()
				if ipre.MatchString(ipstrig) {
					ips[byName.Name] = ipnet.IP.String()
				}
			}

		}
	}
	return ips, nil
}

//GetIp is a function
func GetIp() (localip string, err error) {
	ips, err := Ips()
	if err != nil {
		panic(err)
	}
	if len(ips) == 0 {
		fmt.Println("没有发现ip，请先连接网络再尝试")
		return "", err
	}
	initlocalip := ips[0]
	return initlocalip, err
}

//ReadFilenames from dir
func ReadFilenames(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	// Create array of names of files which are stored in dir
	// used later to set valid name for received files
	filenames := make([]string, len(files))
	for _, fi := range files {
		filenames = append(filenames, fi.Name())
	}
	return filenames
}

// GetFileName generates a file name based on the existing files in the directory
// if name isn't taken leave it unchanged
// else change name to format "name(number).ext"
func GetFileName(newFilename string, fileNamesInTargetDir []string) string {
	fileExt := filepath.Ext(newFilename)
	fileName := strings.TrimSuffix(newFilename, fileExt)
	number := 1
	i := 0
	for i < len(fileNamesInTargetDir) {
		if newFilename == fileNamesInTargetDir[i] {
			newFilename = fmt.Sprintf("%s(%v)%s", fileName, number, fileExt)
			number++
			i = 0
		}
		i++
	}
	return newFilename
}

//PathExists check file/folder exists
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
