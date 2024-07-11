package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Result struct {
	InputDir      string
	FoundFilePath chan string
	IPMap         map[string]int
	mapLock       sync.Mutex
	Threads       int
	discreteMap   map[string]*DiversityCalculator
	mu            sync.Mutex
}

type DiversityCalculator struct {
	ipCounts        map[string]int
	totalCount      int
	shannonIndex    float64
	maxShannonIndex float64
	mu              sync.Mutex
}

func (R *Result) offlineWatch() {
	defer close(R.FoundFilePath)

	err := filepath.WalkDir(R.InputDir, func(root string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err) // 可能会有访问权限等错误
			return nil
		}

		if d.IsDir() {
			return nil // 跳过目录
		}

		R.FoundFilePath <- path.Join(root)

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
}

func (R *Result) execTransfer() {
	var wg sync.WaitGroup

	for i := 0; i < R.Threads; i++ {

		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for filename := range R.FoundFilePath {
				//R.cleanIpList(filename)
				R.clientIPDiscreteAnalyze(filename)
			}

		}(i)
	}
	wg.Wait()

	//R.Output()
	R.OutputDiscreteIndex()
}

// 违规DNS监测
func (R *Result) cleanIpList(fileName string) {

	start := time.Now()
	counts := 0

	file, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		log.Println(err)
		return
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		counts++
		arr := strings.Split(scanner.Text(), "|")
		if len(arr) < 4 {
			continue
		}
		domain := arr[2]
		if domain == "" {
			continue
		}
		R.mapLock.Lock()
		R.IPMap[domain]++
		R.mapLock.Unlock()

	}

	execTime := time.Since(start).Seconds()
	qps := int(float64(counts) / execTime)
	fmt.Printf("%s qps %d \n", fileName, qps)

}

func (R *Result) clientIPDiscreteAnalyze(fileName string) {

	start := time.Now()
	counts := 0
	discreteMap := make(map[string][]string)

	file, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		log.Println(err)
		return
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		counts++
		arr := strings.Split(scanner.Text(), "|")
		if len(arr) < 4 {
			continue
		}
		clientIP := arr[2]
		dnsIP := arr[3]
		if clientIP == "" {
			continue
		}

		discreteMap[dnsIP] = append(discreteMap[dnsIP], clientIP)

	}

	for dnsIP, clientIPs := range discreteMap {
		R.mu.Lock()
		if _, exists := R.discreteMap[dnsIP]; !exists {
			R.discreteMap[dnsIP] = NewDiversityCalculator()
		}
		R.discreteMap[dnsIP].Update(clientIPs)
		R.mu.Unlock()
	}

	execTime := time.Since(start).Seconds()
	qps := int(float64(counts) / execTime)
	fmt.Printf("%s qps %d \n", fileName, qps)

}

func (R *Result) Output() {
	type kv struct {
		Key   string
		Value int
	}
	var kvList []kv
	for s, i := range R.IPMap {
		kvList = append(kvList, kv{Key: s, Value: i})
	}
	sort.Slice(kvList, func(i, j int) bool {
		return kvList[i].Value > kvList[j].Value
	})

	// 打开输出文件
	output, err := os.OpenFile("result.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open output file: %v", err)
	}
	defer output.Close()
	for _, kv := range kvList {
		output.WriteString(strconv.Itoa(kv.Value) + " " + kv.Key + "\n")
	}
}

func (R *Result) OutputDiscreteIndex() {
	type kv struct {
		Key   string
		Value float64
	}
	var kvList []kv
	for s, i := range R.discreteMap {
		kvList = append(kvList, kv{Key: s, Value: i.GetShannonIndex()})
	}

	sort.Slice(kvList, func(i, j int) bool {
		return kvList[i].Value < kvList[j].Value
	})

	// 打开输出文件
	output, err := os.OpenFile("result.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open output file: %v", err)
	}
	defer output.Close()
	for _, kv := range kvList {
		fmt.Printf("%f %s\n", kv.Value, kv.Key)
		//output.WriteString()
	}
}
