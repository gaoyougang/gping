package show

import (
	// "strconv"
	// "strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var (
	// redStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	// orangeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	// greenStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	defaultStyle = lipgloss.NewStyle()

	// styledCell = func(value string, rowData []string) lipgloss.Style {
	// 	// 状态列样式
	// 	if len(rowData) > 8 && value == rowData[8] { // Status列
	// 		switch value {
	// 		case "LOST", "FAILED":
	// 			return redStyle
	// 		case "PASS":
	// 			return greenStyle
	// 		}
	// 	}

	// 	// Loss列样式
	// 	if len(rowData) > 4 && value == rowData[4] { // Loss列
	// 		lossStr := strings.TrimSuffix(value, "%")
	// 		if loss, err := strconv.ParseFloat(lossStr, 64); err == nil {
	// 			switch {
	// 			case loss >= 80 && loss <= 100:
	// 				return redStyle
	// 			case loss >= 60 && loss < 80:
	// 				return orangeStyle
	// 			}
	// 		}
	// 	}
	// 	return defaultStyle
	// }

	// 自定义样式
	styles = table.Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1),
		Cell: lipgloss.NewStyle().Padding(0, 1),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#A550DF")).
			Padding(0, 0),
	}
)
