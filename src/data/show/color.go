package show

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var (
	// 头部样式
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	// 选中的样式
	selectStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#A550DF")).
			Padding(0, 0)

	// 关键字颜色的映射
	// typeColors = map[string]lipgloss.Color{
	// 	"PASS":         lipgloss.Color("#75FBAB"),
	// 	"FAILED":       lipgloss.Color("#FF7698"),
	// 	"UNSTABLE":     lipgloss.Color("#00E2C7"),
	// 	"HIGH_LATENCY": lipgloss.Color("#D7FF87"),
	// }

	// 自定义样式
	styles = table.Styles{
		Header:   headerStyle,
		Cell:     lipgloss.NewStyle().Padding(0, 1),
		Selected: selectStyle,
	}
)

// func getCellStyle() func(row []table.Row, col int) lipgloss.Style {
// 	baseStyle := lipgloss.NewStyle().Padding(0, 1)

// 	return func(row []table.Row, col int) lipgloss.Style {
// 		// 检查 row 是否为空 或 col 越界
// 		if len(row) == 0 || col < 0 || col >= len(row) {
// 			return baseStyle
// 		}
// 		for _, val := range row {
// 			key := val[col]
// 			if color, ok := typeColors[key]; ok {
// 				return baseStyle.Foreground(color).Bold(true)
// 			}
// 		}
// 		return baseStyle
// 	}
// }
