package ip

import (
	"net"
	"strings"
)

//获取本地对外的IP
func GetOutboundIP() (ip string, err error) {

	//通过Google的dns服务器8.8.8.8:80获取使用的ip
	//用udp的方式拨号
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
	}
	defer conn.Close()

	//获取本地网络地址
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	//将ip从地址中分离出来
	ip = strings.Split(localAddr.IP.String(), ":")[0]

	return
}
