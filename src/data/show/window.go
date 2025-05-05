package show

import (
	"gping/src/data/ping"
	"gping/src/utils"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 定义一个窗口模型,其由2个表, 1个文本框和命令输入框组成
type windowM struct {
	table      table.Model
	ipIputArea textarea.Model
	focused    string
	tableData  [][]string
	quitting   bool
	statusMsg  string
	pingContr  *ping.PingController
	result     chan ping.PingResult
}

// 定义一个通道信息
type pingResultMsg ping.PingResult

// 初始化窗口
func initialWindowM() *windowM {
	// 初始化PASS表格
	tableHeader := []table.Column{
		{Title: "OD", Width: 6},
		{Title: "Ipaddr", Width: 40},
		{Title: "Sent", Width: 8},
		{Title: "Recev", Width: 8},
		{Title: "Loss(%)", Width: 8},
		{Title: "AVG RTT", Width: 8},
		{Title: "MIN RTT", Width: 8},
		{Title: "MAX RTT", Width: 8},
		{Title: "Status", Width: 6},
		{Title: "Last Update", Width: 20},
	}

	// 创建表
	t := table.New(
		table.WithColumns(tableHeader),
		table.WithHeight(50),
		table.WithStyles(styles),
	)

	// 初始化文本区域
	ipArea := textarea.New()
	ipArea.Placeholder = "👉 Input IP or domain name"
	ipArea.SetWidth(45)
	ipArea.SetHeight(50)
	ipArea.Focus()

	return &windowM{
		table:      t,
		ipIputArea: ipArea,
		focused:    "ipArea",
		tableData:  nil,
		statusMsg:  "Ready",
		result:     make(chan ping.PingResult, 100),
		pingContr:  nil,
	}
}

// 辅助函数：将二维字符串切片转换为表格行
func convertToRows(data [][]string) []table.Row {
	rows := make([]table.Row, len(data))
	for i, item := range data {
		rows[i] = table.Row(item)
	}
	return rows
}

// 为窗口模型实现Init()
func (w *windowM) Init() tea.Cmd {
	return nil
}

// 为窗口实现等待数据刷新
func (w *windowM) waitForResult() tea.Msg {
	result, ok := <-w.result
	if !ok {
		return nil
	}
	return pingResultMsg(result)
}

// 为窗口实现Update()
func (w *windowM) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c":
			w.quitting = true
			if w.pingContr != nil {
				w.pingContr.StopAll()
				w.pingContr.Wait()
			}
			close(w.result)
			return w, tea.Quit

		case "tab":
			// 按 'tab' 切换焦点
			switch w.focused {
			case "table":
				w.focused = "ipArea"
				w.table.Blur()
				w.ipIputArea.Focus()
			case "ipArea":
				w.focused = "table"
				w.ipIputArea.Blur()
				w.table.Focus()
			}

		case "ctrl+b":
			if w.focused == "ipArea" {
				iplist := utils.RemoveDuplicates(strings.Split(w.ipIputArea.Value(), "\n"))
				if utils.IsNonEmptyStrings(iplist) {
					ips := utils.FilterLines(iplist)
					if w.pingContr == nil {
						w.pingContr = ping.NewPingController()
						w.pingContr.StartPing(ips, w.result)
					} else {
						for _, ip := range ips {
							w.pingContr.AddPing(ip, w.result)
						}
					}

					w.tableData = w.pingContr.GetResultsStringSlice()
					w.table.SetRows(convertToRows(w.tableData))
					// 重置表格光标位置
					if len(w.tableData) > 0 {
						w.table.SetCursor(0)
					}
					return w, tea.Batch(
						w.waitForResult,               // 开始监听结果
						func() tea.Msg { return nil }, // 确保UI更新
					)
				}
				return w, nil
			}

		case "ctrl+s": // 停止协程中的ping
			if w.focused == "table" {
				if len(w.tableData) > 0 && w.table.Cursor() < len(w.tableData) {
					// 删除数据
					w.tableData = append(w.tableData[:w.table.Cursor()], w.tableData[w.table.Cursor()+1:]...)
					// 更新表格
					w.table.SetRows(convertToRows(w.tableData))
					// 调整光标位置
					if w.table.Cursor() >= len(w.tableData) && len(w.tableData) > 0 {
						w.table.SetCursor(len(w.tableData) - 1)
					}
					w.statusMsg = "deleted from Ping queue"
				}
			}
		}

	case pingResultMsg:
		// 更新数据
		w.tableData = w.pingContr.GetResultsStringSlice()

		// 更新表格行
		w.table.SetRows(convertToRows(w.tableData))

		// 保持当前焦点和光标位置
		switch w.focused {
		case "table":
			w.table.Focus()
		}

		w.statusMsg = "Data updated"
		return w, w.waitForResult
	}

	// 更新当前聚焦组件
	switch w.focused {
	case "table":
		w.table, cmd = w.table.Update(msg)
		cmds = append(cmds, cmd)
	case "ipArea":
		w.ipIputArea, cmd = w.ipIputArea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return w, tea.Batch(cmds...)
}

// 为窗口实现view
func (w *windowM) View() string {
	if w.quitting {
		return "Goodbye! \n"
	}

	// 渲染table
	tV := w.table.View()
	if w.focused == "table" {
		tV = activeBorderStyle.Render(tV)
	} else {
		tV = borderStyle.Render(tV)
	}
	tV = lipgloss.NewStyle().MarginBottom(1).Render(tV)

	tablesView := lipgloss.NewStyle().MarginBottom(2).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			"Running Ping",
			tV,
		),
	)

	// 渲染文本区域
	textareaView := w.ipIputArea.View()
	if w.focused == "ipArea" {
		textareaView = activeBorderStyle.Render(textareaView)
	} else {
		textareaView = borderStyle.Render(textareaView)
	}

	textareaView = lipgloss.NewStyle().MarginBottom(2).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			"IP ADDRESSES",
			textareaView,
		),
	)

	// 上部分
	uppart := lipgloss.JoinHorizontal(lipgloss.Top,
		tablesView,
		"  ",
		textareaView,
	)

	// 状态信息
	status := statusStyle.Render(w.statusMsg)
	help := "Tab: Switch focus | ↑/↓: Navigate | focus ip addresss ctrl + b: start ping | ctrl + c: Quit\n"
	return lipgloss.JoinVertical(
		lipgloss.Left,
		uppart,
		status,
		help,
	)
}

// 项目对外窗口
func NewWindowProgram() *tea.Program {
	return tea.NewProgram(initialWindowM(), tea.WithAltScreen())
}
