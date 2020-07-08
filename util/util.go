package util

import (
	"net"
	"regexp"
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

// func main() {
// 	ips, err := Ips()
// 	if err != nil {
// 		panic(err)
// 	}
// 	if len(ips) == 0 {
// 		fmt.Println("没有发现ip，请先连接网络再尝试")
// 		return
// 	}
// 	fmt.Println(ips[0])
// }
