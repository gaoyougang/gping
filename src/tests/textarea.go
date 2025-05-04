package tests

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 定义一个文本输入模型
type model struct {
	rows     []string
	allData  [][]interface{} // 存储所有数据
	textarea textarea.Model
	ready    bool
	viewport viewport.Model
}

func NewProgram() *tea.Program {
	return tea.NewProgram(initModel())
}

// 为模型增加一个初始化函数
func initModel() model {
	// 定义列规格
	columns := []struct {
		title  string
		width  int
		format string
	}{
		{"Date", 20, "%-20s"},
		{"ID", 5, "%-5d"},
		{"Ipaddress", 15, "%-15s"},
		{"Status", 10, "%-10s"},
		{"Packet", 20, "%-20s"},
		{"Quality", 20, "%-20s"},
	}

	// 生成测试数据(100行)
	allData := generateTableData(100)

	// 生成初始表格行
	var rows []string
	header := generateHeader(columns)
	rows = append(rows, header)

	// 只添加前20行数据
	for i := 0; i < 10 && i < len(allData); i++ {
		row := generateRow(allData[i], columns)
		rows = append(rows, row)
	}

	ti := textarea.New()
	// 设置提示
	ti.Placeholder = "👉 Input IP or Domain name"
	// 设置宽
	ti.SetWidth(45)
	// 设置高
	ti.SetHeight(32)
	// 选择时定位
	ti.Focus()

	return model{
		textarea: ti,
		allData:  allData,
		rows:     rows,
	}
}

/*
 -> 为模型实现bubbletea.Model的接口:
	1. Init() Cmd
	2. Update(Msg) (Model, Cmd)
	3. View() string
*/

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// 更新视口显示的内容
func (m *model) updateViewportContent() {
	// 计算当前滚动位置对应的行范围
	start := m.viewport.YOffset
	end := start + 10 // 显示20行
	if end > len(m.allData) {
		end = len(m.allData)
	}

	// 生成表头
	columns := []struct {
		title  string
		width  int
		format string
	}{
		{"Date", 20, "%-20s"},
		{"ID", 5, "%-5d"},
		{"Ipaddress", 15, "%-15s"},
		{"Status", 10, "%-10s"},
		{"Packet", 20, "%-20s"},
		{"Quality", 20, "%-20s"},
	}
	header := generateHeader(columns)

	// 生成当前可见的行
	var visibleRows []string
	visibleRows = append(visibleRows, header)
	for i := start; i < end; i++ {
		row := generateRow(m.allData[i], columns)
		visibleRows = append(visibleRows, row)
	}

	// 设置视口内容
	fullContent := strings.Join(m.generateAllRows(), "\n")
	m.viewport.SetContent(fullContent)
	
	// 但只显示部分内容(由视口的YOffset和高度决定)
	m.rows = visibleRows
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

		switch msg.String() {
		case "pgup":
			m.viewport.ScrollUp(10)
			m.updateViewportContent()
		case "pgdown":
			m.viewport.ScrollDown(10)
			m.updateViewportContent()
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			// 初始化视口
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	tableStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2)

	// 只显示当前视口范围内的行
	content := strings.Join(m.rows, "\n")

	// 添加滚动指示器
	scrollPercent := float64(m.viewport.YOffset) / float64(len(m.allData)-20)
	if scrollPercent < 0 {
		scrollPercent = 0
	}
	if scrollPercent > 1 {
		scrollPercent = 1
	}

	scrollIndicator := fmt.Sprintf("↓ %d/%d (%.0f%%)",
		m.viewport.YOffset,
		len(m.allData)-20,
		scrollPercent*100)

	// 创建紫色粗圆角边框样式
	textarea_borderStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder(), true).   // 使用粗边框
		BorderStyle(lipgloss.RoundedBorder()).  // 设置为圆角
		BorderForeground(lipgloss.Color("99")). // 紫色边框颜色
		Padding(0, 1)                           // 内边距

	// 应用样式
	styleTable := tableStyle.Render(content + "\n\n" + scrollIndicator)
	styleTextArea := textarea_borderStyle.Render(m.textarea.View())

	layout_one := lipgloss.JoinVertical(lipgloss.Left, styleTable ,styleTable)
	layout := lipgloss.JoinHorizontal(lipgloss.Top,layout_one, styleTextArea)

	// 应用边框到 TextArea
	return fmt.Sprintf("Please enter the address where you want to ping 👇\n%s\n\n%s\n", layout, "(ctrl+b to begin)")
}

// 生成所有行的内容(用于滚动)
func (m *model) generateAllRows() []string {
	columns := []struct {
		title  string
		width  int
		format string
	}{
		{"Date", 20, "%-20s"},
		{"ID", 5, "%-5d"},
		{"Ipaddress", 15, "%-15s"},
		{"Status", 10, "%-10s"},
		{"Packet", 20, "%-20s"},
		{"Quality", 20, "%-20s"},
	}

	var allRows []string
	allRows = append(allRows, generateHeader(columns))
	for _, data := range m.allData {
		allRows = append(allRows, generateRow(data, columns))
	}
	return allRows
}

func generateHeader(columns []struct {
	title  string
	width  int
	format string
}) string {
	var headerParts []string
	for _, col := range columns {
		headerParts = append(headerParts, fmt.Sprintf("%-*s", col.width, col.title))
	}
	return strings.Join(headerParts, " ")
}

func generateRow(data []interface{}, columns []struct {
	title  string
	width  int
	format string
}) string {
	var rowParts []string
	for i, col := range columns {
		if i < len(data) {
			rowParts = append(rowParts, fmt.Sprintf(col.format, data[i]))
		} else {
			rowParts = append(rowParts, strings.Repeat(" ", col.width))
		}
	}
	return strings.Join(rowParts, " ")
}

func generateTableData(rowCount int) [][]interface{} {
	departments := []string{"Engineering", "Marketing", "Sales", "HR", "Finance"}
	statuses := []string{"Active", "Inactive", "On Leave", "Terminated"}

	var data [][]interface{}
	for i := 1; i <= rowCount; i++ {
		row := []interface{}{
			time.Now().Format("2006-01-02 15:04:05"),
			i, // ID
			fmt.Sprintf("%d", i), // Name
			statuses[i%len(statuses)], // Status
			departments[i%len(departments)], // Department
			"OK",
		}
		data = append(data, row)
	}
	return data
}
