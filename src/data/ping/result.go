package ping

import (
	"fmt"
	"gping/src/utils"
	"sync"
	"sync/atomic"
	"time"

	ping "github.com/prometheus-community/pro-bing"
)

var oid uint32 = 0

// 定义一个Result结构体用于存储ping的结果
type PingResult struct {
	id         uint32
	ip         string
	sent       uint32
	received   uint32
	loss       float64
	avgRTT     time.Duration
	minRTT     time.Duration
	maxRTT     time.Duration
	lastStatus string
	lastUpdate time.Time
	err        error
}

// 定义一个ping的控制器,用于管理多个ping任务
type PingController struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	stopChs map[string]chan struct{} // 每个IP对应的停止通道
	results map[string]PingResult    // 每个IP的最新结果
	allDone chan struct{}            // 所有任务完成的信号
}

// 创建一个新的ping控制
func NewPingController() *PingController {
	return &PingController{
		stopChs: make(map[string]chan struct{}, 0),
		results: make(map[string]PingResult, 0),
		allDone: make(chan struct{}),
	}
}

// updateResult 更新结果并发送到通道
func (pc *PingController) updateResult(ip string, result PingResult, results chan<- PingResult) {
	pc.mu.Lock()
	pc.results[ip] = result
	pc.mu.Unlock()

	select {
	case results <- result:
	default:
	}
}

// AddPing 在现有的控制器上增加新的 ping 协程
func (pc *PingController) AddPing(ip string, results chan<- PingResult) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// 检查是否已存在该IP的ping任务
	if _, exists := pc.stopChs[ip]; exists {
		return
	}

	// 创建新的停止通道
	stopCh := make(chan struct{})
	pc.stopChs[ip] = stopCh
	newID := atomic.AddUint32(&oid, 1)
	pc.wg.Add(1)

	// 启动新的ping协程
	go func(id uint32, ipaddr string) {
		defer pc.wg.Done()
		pc.pingIP(newID, ip, results, stopCh)
	}(newID, ip)
}

// StopPing 根据IP停止对应的ping任务
func (pc *PingController) StopPing(ip string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if stopCh, exists := pc.stopChs[ip]; exists {
		close(stopCh)
		delete(pc.stopChs, ip)
		delete(pc.results, ip)
	}
}

// StopAll停止所有ping任务
func (pc *PingController) StopAll() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	for ip, stopCh := range pc.stopChs {
		close(stopCh)
		delete(pc.stopChs, ip)
		delete(pc.results, ip)
	}
}

// Wait等待所有ping任务完成
func (pc *PingController) Wait() {
	<-pc.allDone
}

func (pc *PingController) GetResultsStringSlice() [][]string {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	var rest [][]string
	for _, res := range pc.results {
		row := []string{
			fmt.Sprintf("%d", res.id),
			res.ip,
			fmt.Sprintf("%d", res.sent),
			fmt.Sprintf("%d", res.received),
			fmt.Sprintf("%.2f", res.loss),
			fmt.Sprintf("%.2f", res.avgRTT.Seconds()*1000),
			fmt.Sprintf("%.2f", res.minRTT.Seconds()*1000),
			fmt.Sprintf("%.2f", res.maxRTT.Seconds()*1000),
			res.lastStatus,
			res.lastUpdate.Format("2006-01-02 15:04:05"),
		}
		rest = append(rest, row)
	}
	results := utils.SortByNumericColumn(rest, 0)
	return results
}

// 启动对多个ip进行ping任务
func (pc *PingController) StartPing(ips []string, results chan<- PingResult) {
	for _, ip := range ips {
		pc.mu.Lock()                             // 对创建的ip进行加锁,确保同一时间只有一个操作,防止判断错误()
		if _, exists := pc.stopChs[ip]; exists { // 检查IP是否存在ping的操作(这里也可以对ips提前去重)
			pc.mu.Unlock() // 解锁
			continue
		}

		// 如果不存在,则创建ping协程
		stopCh := make(chan struct{}) // 在stopChs创建一个ip对应的空结构体,用于监听停止信号
		pc.stopChs[ip] = stopCh
		newid := atomic.AddUint32(&oid, 1)
		pc.mu.Unlock() // 解锁
		pc.wg.Add(1)   // wg计数 +1

		go func(index uint32, ipaddr string) {
			defer pc.wg.Done()
			pc.pingIP(newid, ip, results, stopCh)
		}(newid, ip)
	}

	go func() {
		pc.wg.Wait() // 等待关闭
		close(pc.allDone)
	}()
}

// pingIP,对单个地址进行ping操作
func (pc *PingController) pingIP(id uint32, ip string, results chan<- PingResult, stopChan <-chan struct{}) {
	// 定义两个uint32整数,用于存储总共发送和接收的数据包
	var totalSent, totalReceived uint32

	for {
		select {
		case <-stopChan:
			pc.mu.Lock()
			delete(pc.stopChs, ip)
			delete(pc.results, ip)
			pc.mu.Unlock()
			return
		default:
			pinger, err := ping.NewPinger(ip)
			if err != nil {
				pc.updateResult(ip, PingResult{
					id:         id,
					ip:         ip,
					sent:       totalSent,
					received:   totalReceived,
					loss:       100.0,
					lastStatus: "FAILED",
					err:        err,
					lastUpdate: time.Now(),
				}, results)
				time.Sleep(2 * time.Second)
				continue
			}

			pinger.Count = 1
			pinger.Timeout = time.Second * 1
			pinger.SetPrivileged(true)

			pinger.OnFinish = func(status *ping.Statistics) {
				totalSent += uint32(status.PacketsSent)
				totalReceived += uint32(status.PacketsRecv)
				loss := float64(totalSent-totalReceived) / float64(totalSent) * 100

				lastStatus := "PASS"
				if status.PacketsRecv == 0 { // 没有收到任何回复包
					lastStatus = "FAILED"
				} else if loss > 50 { // 丢包率超过50%认为不正常
					lastStatus = "UNSTABLE"
				} else if status.AvgRtt > time.Millisecond*100 { // 平均延迟超过100ms认为不理想
					lastStatus = "HIGH_LATENCY"
				}
				pc.updateResult(ip, PingResult{
					id:         id,
					ip:         ip,
					sent:       totalSent,
					received:   totalReceived,
					loss:       loss,
					avgRTT:     status.AvgRtt,
					minRTT:     status.MinRtt,
					maxRTT:     status.MaxRtt,
					lastStatus: lastStatus,
					lastUpdate: time.Now(),
				}, results)
			}

			err = pinger.Run()
			if err != nil {
				totalSent += uint32(pinger.Count)
				pc.updateResult(ip, PingResult{
					id:         id,
					ip:         ip,
					sent:       totalSent,
					received:   totalReceived,
					loss:       float64(totalSent-totalReceived) / float64(totalSent) * 100,
					lastStatus: "LOST",
					err:        err,
					lastUpdate: time.Now(),
				}, results)
			}
			time.Sleep(1 * time.Second)
		}
	}
}
