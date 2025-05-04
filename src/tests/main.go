package tests

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	ping "github.com/prometheus-community/pro-bing"
	"github.com/olekukonko/tablewriter"
)

type PingResult struct {
	IP           string
	Sent         int
	Received     int
	Loss         float64
	AvgRTT       time.Duration
	MinRTT       time.Duration
	MaxRTT       time.Duration
	LastStatus   string
	LastUpdate   time.Time
	Error        error
}

func Tests() {
	ips := []string{
		"8.8.8.8",
		"1.1.1.1",
		"114.114.114.114",
		"223.5.5.5",
		"example.com",
	}

	results := make(chan PingResult, len(ips))
	resultMap := sync.Map{}

	// 初始化表格
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"IP", "Sent", "Received", "Loss", "Avg RTT", "Min RTT", "Max RTT", "Status", "Last Update"})
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	// 启动 goroutine 更新表格
	go func() {
		for res := range results {
			// 加载现有结果
			if existing, ok := resultMap.Load(res.IP); ok {
				existingRes := existing.(PingResult)
				res.Sent += existingRes.Sent
				res.Received += existingRes.Received
			}
			resultMap.Store(res.IP, res)
			renderTable(&resultMap, table)
		}
	}()

	// 启动 ping goroutines
	var wg sync.WaitGroup
	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			pingIP(ip, results)
		}(ip)
	}

	// 处理退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		close(results)
		os.Exit(0)
	}()

	wg.Wait()
	close(results)
}

func pingIP(ip string, results chan<- PingResult) {
	var totalSent, totalReceived int
	
	for {
		pinger, err := ping.NewPinger(ip)
		if err != nil {
			results <- PingResult{
				IP:         ip,
				Sent:       totalSent,
				Received:   totalReceived,
				Loss:       100.0,
				LastStatus: "Error",
				Error:      err,
				LastUpdate: time.Now(),
			}
			time.Sleep(2 * time.Second)
			continue
		}

		pinger.Count = 1 
		pinger.Timeout = time.Second * 1
		pinger.SetPrivileged(true)

		pinger.OnFinish = func(stats *ping.Statistics) {
			totalSent += stats.PacketsSent
			totalReceived += stats.PacketsRecv
			
			loss := float64(totalSent-totalReceived) / float64(totalSent) * 100

			results <- PingResult{
				IP:         ip,
				Sent:       totalSent,
				Received:   totalReceived,
				Loss:       loss,
				AvgRTT:     stats.AvgRtt,
				MinRTT:     stats.MinRtt,
				MaxRTT:     stats.MaxRtt,
				LastStatus: "OK",
				LastUpdate: time.Now(),
			}
		}

		err = pinger.Run()
		if err != nil {
			totalSent += pinger.Count // 即使失败也计入发送包数
			results <- PingResult{
				IP:         ip,
				Sent:       totalSent,
				Received:   totalReceived,
				Loss:       float64(totalSent-totalReceived)/float64(totalSent)*100,
				LastStatus: "Error",
				Error:      err,
				LastUpdate: time.Now(),
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func renderTable(resultMap *sync.Map, table *tablewriter.Table) {
	var data [][]string

	resultMap.Range(func(key, value interface{}) bool {
		res := value.(PingResult)
		row := []string{
			res.IP,
			fmt.Sprintf("%d", res.Sent),
			fmt.Sprintf("%d", res.Received),
			fmt.Sprintf("%.1f%%", res.Loss),
			res.AvgRTT.String(),
			res.MinRTT.String(),
			res.MaxRTT.String(),
			res.LastStatus,
			res.LastUpdate.Format("15:04:05"),
		}
		data = append(data, row)
		return true
	})

	fmt.Print("\033[H\033[2J") // 清屏
	table.ClearRows()
	table.AppendBulk(data)
	table.Render()
}