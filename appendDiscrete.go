package main

import (
	"math"
)

// NewDiversityCalculator 初始化多样性计算器
func NewDiversityCalculator() *DiversityCalculator {
	return &DiversityCalculator{
		ipCounts: make(map[string]int),
	}
}

// Update 更新IP计数和多样性指数
func (dc *DiversityCalculator) Update(ips []string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	for _, ip := range ips {
		if _, exists := dc.ipCounts[ip]; exists {
			dc.ipCounts[ip]++
		} else {
			dc.ipCounts[ip] = 1
		}
		dc.totalCount++
	}
	dc.calculateShannonIndex()
}

// 计算Shannon多样性指数
func (dc *DiversityCalculator) calculateShannonIndex() {
	dc.shannonIndex = 0.0
	for _, count := range dc.ipCounts {
		proportion := float64(count) / float64(dc.totalCount)
		dc.shannonIndex -= proportion * math.Log2(proportion)
	}
	dc.maxShannonIndex = math.Log2(float64(len(dc.ipCounts)))
}

// GetShannonIndex 获取当前的Shannon多样性指数
func (dc *DiversityCalculator) GetShannonIndex() float64 {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	return dc.shannonIndex
}

// GetEvennessIndex 获取当前的均一性指数
func (dc *DiversityCalculator) GetEvennessIndex() float64 {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if len(dc.ipCounts) == 0 || dc.maxShannonIndex == 0 {
		return 1.0
	}
	return dc.shannonIndex / dc.maxShannonIndex
}

//func main() {
//	// 创建多样性计算器
//	dc := NewDiversityCalculator()
//
//	// 初始IP地址列表
//	str1 := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.1", "192.168.1.2", "192.168.1.1"}
//
//	// 更新初始IP列表并计算多样性指数
//	dc.Update(str1)
//	fmt.Printf("初始IP列表的Shannon多样性指数: %f\n", dc.GetShannonIndex())
//	fmt.Printf("初始IP列表的均一性指数: %f\n\n", dc.GetEvennessIndex())
//
//	// 新增IP地址列表
//	str2 := []string{"192.168.1.4", "192.168.1.5", "192.168.1.6", "192.168.1.1", "192.168.1.2"}
//
//	// 更新新的IP列表并重新计算多样性指数
//	dc.Update(str2)
//	fmt.Printf("更新后IP列表的Shannon多样性指数: %f\n", dc.GetShannonIndex())
//	fmt.Printf("更新后IP列表的均一性指数: %f\n", dc.GetEvennessIndex())
//}
