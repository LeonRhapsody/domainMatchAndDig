package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func main() {
	var (
		mode    string
		threads int
		isDig   bool
	)

	if runtime.GOOS == "darwin" {
		mode = "dig"
		threads = 10
	} else {
		mode = os.Args[1]
		threads, _ = strconv.Atoi(os.Args[2])
	}

	if mode == "dig" {
		isDig = true
	}

	// （解析次数 IP）
	ipListFile := "1.txt"

	// IP范围文件路径(200.12.0.0/15 代播_)
	ipRangeFile := "vip.txt"

	// 输出文件路径
	outputFile := "output.txt"

	// 打开 IP 清单文件
	ipList, err := os.Open(ipListFile)
	if err != nil {
		log.Fatalf("Failed to open IP list file: %v", err)
	}
	defer ipList.Close()
	// 打开输出文件
	output, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open output file: %v", err)
	}
	defer output.Close()

	// 读取 IP 地址段文件，构建 IP 范围列表
	_, map1, err := readIPRanges(ipRangeFile)
	if err != nil {
		log.Fatalf("Failed to read IP range file: %v", err)
	}

	ipList, err = os.Open(ipListFile)
	if err != nil {
		log.Fatalf("Failed to open IP list file %s: %v", ipListFile, err)
	}
	defer ipList.Close()

	var lines []string
	scanner := bufio.NewScanner(ipList)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error scanning file: %v", err)
	}

	// 计算需要启动的线程数
	linesPerThread := len(lines) / threads
	if len(lines)%linesPerThread != 0 {
		linesPerThread++
	}

	// 使用 WaitGroup 来等待所有文件处理完成
	var wg sync.WaitGroup

	// 并行处理每个 IP 清单文件
	for i := 0; i < threads; i++ {
		fmt.Println("start thread ", i)
		wg.Add(1)
		go func(startIdx int) {
			defer wg.Done()

			// 计算本线程的结束索引
			endIdx := startIdx + linesPerThread
			if endIdx > len(lines) {
				endIdx = len(lines)
			}

			// 处理本线程的行
			for j := startIdx; j < endIdx; j++ {
				result := processLine(isDig, lines[j], &map1)
				fmt.Printf("%s\n", result)
				output.WriteString(fmt.Sprintf("%s\n", result))
			}
		}(i * linesPerThread)

		// 等待所有文件处理完成
	}
	wg.Wait()

}

// 读取 IP 地址段文件，返回 IP 范围列表
func readIPRanges(filename string) ([]net.IPNet, map[string][]net.IPNet, error) {
	var ipRanges []net.IPNet
	map1 := make(map[string][]net.IPNet)

	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		s1 := strings.Split(line, " ")
		client := s1[len(s1)-1]
		ip1 := s1[0]

		_, ipNet, err := net.ParseCIDR(ip1)
		if err != nil {
			ips := strings.Split(ip1, "-")
			if len(ips) != 2 {
				continue
			}
			startIP := net.ParseIP(strings.TrimSpace(ips[0]))
			endIP := net.ParseIP(strings.TrimSpace(ips[1]))

			// 将起始 IP 地址和结束 IP 地址之间的每个 IP 地址都添加到 IP 范围列表中
			for ip := startIP; compareIP(ip, endIP) <= 0; ip = incrementIP(ip) {
				map1[client] = append(map1[client], net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)})
				ipRanges = append(ipRanges, net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)})
			}
		} else {
			map1[client] = append(map1[client], *ipNet)
			ipRanges = append(ipRanges, *ipNet)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return ipRanges, map1, nil
}

// 增加 IP 地址
func incrementIP(ip net.IP) net.IP {
	nextIP := make(net.IP, len(ip))
	copy(nextIP, ip)

	for i := len(nextIP) - 1; i >= 0; i-- {
		nextIP[i]++
		if nextIP[i] > 0 {
			break
		}
	}

	return nextIP
}

// 比较两个 IP 地址
func compareIP(ip1, ip2 net.IP) int {
	for i := 0; i < len(ip1) && i < len(ip2); i++ {
		if ip1[i] != ip2[i] {
			return int(ip1[i]) - int(ip2[i])
		}
	}
	return len(ip1) - len(ip2)
}

// 判断 IP 是否在 IP 范围内
func isInIPRanges(ip string, map1 *map[string][]net.IPNet) (bool, string) {
	ipAddr := net.ParseIP(ip)
	for client, value := range *map1 {
		for _, ipNet := range value {
			if ipNet.Contains(ipAddr) {
				return true, client
			}
		}
	}
	return false, ""
}

// 处理每一行的逻辑
func processLine(isDig bool, line string, map1 *map[string][]net.IPNet) string {

	fields := strings.Fields(line)
	var nums string
	var ip string
	var status string

	if isDig {
		nums = fields[0]
		ip = fields[1]
		status, _ = getDNSStatus("www.qq.com", ip)
	} else {
		ip = fields[0]
	}

	var result string
	if ok, client := isInIPRanges(ip, map1); ok {
		result = fmt.Sprintf("%s %s %s %s", nums, ip, status, client)
	} else {
		result = fmt.Sprintf("%s %s %s %s", nums, ip, status, "未知")
	}

	return result

}
