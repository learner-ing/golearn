package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var startIp, endIp, port, filePath string
var threads int

func Init() {
	flag.StringVar(&startIp, "s", "", "Start IP")
	flag.StringVar(&endIp, "e", "", "End IP")
	flag.StringVar(&port, "p", "445", "port")
	flag.StringVar(&filePath, "f", "ip.txt", "ip list")
	flag.IntVar(&threads, "t", 10, "Threads,Default 10")
}
func main() {
	Init()
	flag.Usage = func() {
		fmt.Println("portscan -s 192.168.1.1 -e 192.168.255.255 -t 10")
		fmt.Println("portscan -s 192.168.1.1 -p 1-65535 -t 10")
		fmt.Println("portscan -f ip.txt -p 1-65535 -t 10")
		flag.PrintDefaults()
	}
	flag.Parse()
	ports := getPort()
	var ips []string
	if startIp != "" {
		if endIp == "" {
			endIp = startIp
		}
		if !checkIp(startIp) && !checkIp(endIp) {
			log.Fatalln("[X]: Ip format error;")
		}
		ips = getIp()
	} else {
		ips = getIpWithFile()
	}
	var addr []string
	for _, i := range ips {
		for _, p := range ports {
			addr = append(addr, i+":"+p)
		}
	}
	var wg sync.WaitGroup
	for {
		if len(addr) <= 0 {
			break
		}
		for i := 0; i < threads; i++ {
			if len(addr) <= 0 {
				break
			}
			wg.Add(1)
			go scan(addr[0], &wg)
			addr = addr[1:]
		}
		wg.Wait()
	}

}

//判断ip格式
func checkIp(ip string) bool {
	ok, err := regexp.Match("^((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})(\\.((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})){3}$", []byte(ip))
	if err != nil {
		return false
	}
	return ok
}

//判断端口范围和线程
func checkNumber(number string) bool {
	p, err := strconv.Atoi(number)
	if err == nil && p <= 65536 && p > 0 {
		return true
	} else {
		return false
	}
}

//获取端口列表
func getPort() []string {
	var ports []string
	p := strings.Split(port, ",")
	for _, value := range p {
		if strings.Contains(value, "-") {
			start := strings.Split(value, "-")[0]
			end := strings.Split(value, "-")[1]
			if checkNumber(start) && checkNumber(end) {
				s, _ := strconv.Atoi(start)
				e, _ := strconv.Atoi(end)
				for i := s; i <= e; i++ {
					if !checkNumber(strconv.Itoa(i)) {
						log.Fatalln("[X] port out of range")
					}
					ports = append(ports, strconv.Itoa(i))
				}
			}
		} else {
			if !checkNumber(value) {
				log.Fatalln("[X] port out of range")
			}
			ports = append(ports, value)
		}
	}
	return ports
}

//获取ip列表
func getIp() []string {
	start := strings.Split(startIp, ".")
	end := strings.Split(endIp, ".")
	s, _ := strconv.Atoi(start[0])
	e, _ := strconv.Atoi(end[0])
	var ips []string
	var i1, i2, i3, i4 string
	for i := s; i <= e; i++ {
		i1 = strconv.Itoa(i) + "."
		s, _ := strconv.Atoi(start[1])
		e, _ := strconv.Atoi(end[1])
		for j := s; j <= e; j++ {
			i2 = strconv.Itoa(j) + "."
			s, _ := strconv.Atoi(start[2])
			e, _ := strconv.Atoi(end[2])
			for k := s; k <= e; k++ {
				i3 = strconv.Itoa(k) + "."
				s, _ := strconv.Atoi(start[3])
				e, _ := strconv.Atoi(end[3])
				for f := s; f <= e; f++ {
					i4 = strconv.Itoa(f)
					ips = append(ips, i1+i2+i3+i4)
				}
			}
		}
	}
	return ips
}
func getIpWithFile() (ips []string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("open " + filePath + "error")
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		ip, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		ips = append(ips, strings.TrimSpace(ip))
	}
	return ips
}
func scan(addr string, wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := net.DialTimeout("tcp", addr, time.Second*1)
	if err == nil {
		log.Println("[+] :" + addr + " is open")
		conn.Close()
	}
}
