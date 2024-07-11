package main

import "math"

// 计算Shannon多样性指数
func shannonDiversityIndex(ipList []string) float64 {
	// 统计每个IP的出现次数
	ipCounts := make(map[string]int)
	for _, ip := range ipList {
		ipCounts[ip]++
	}

	totalCount := len(ipList)
	shannonIndex := 0.0

	// 计算Shannon多样性指数
	for _, count := range ipCounts {
		proportion := float64(count) / float64(totalCount)
		shannonIndex -= proportion * math.Log2(proportion)
	}

	return shannonIndex
}

// 计算均一性（归一化的Shannon多样性指数）
// 当比较两组不同规模的IP数据时，直接比较Shannon指数可能会出现偏差，因为Shannon指数与样本大小有关。为了解决这一问题，可以使用归一化的Shannon指数，或称为均一性（evenness），它将Shannon指数标准化为一个0到1之间的值，使得不同规模的数据可以进行比较。
// 均一性计算公式如下：
// E=H/Hmax

func evennessIndex(ipList []string) float64 {
	ipCounts := make(map[string]int)
	for _, ip := range ipList {
		ipCounts[ip]++
	}

	uniqueIPs := len(ipCounts)
	if uniqueIPs == 1 {
		return 1.0
	}

	shannonIndex := shannonDiversityIndex(ipList)
	maxShannonIndex := math.Log2(float64(uniqueIPs))

	return shannonIndex / maxShannonIndex
}
