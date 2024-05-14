package main

import (
	"github.com/miekg/dns"
)

func getDNSStatus(hostname string, dnsAddress string) (string, error) {
	// 创建 DNS 客户端
	client := dns.Client{}
	message := dns.Msg{}
	message.SetQuestion(dns.Fqdn(hostname), dns.TypeA)

	// 执行 DNS 查询
	response, _, err := client.Exchange(&message, dnsAddress+":53")
	if err != nil {
		return "无响应", err
	}

	// 获取状态码
	status := dns.RcodeToString[response.Rcode]
	return status, nil
}
