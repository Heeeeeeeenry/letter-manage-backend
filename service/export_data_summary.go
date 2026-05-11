package service

import (
	"fmt"
	"path/filepath"
	"time"

	"letter-manage-backend/model"

	"github.com/xuri/excelize/v2"
)

// GenerateDataSummary 生成 {月份}数据汇总.xlsx
// 参考 4月数据汇总.xlsx 的结构：信箱/工单/专席/信访子系统/日报/日报数据统计
func GenerateDataSummary(dir, periodLabel string, startTime, endTime time.Time) (string, error) {
	f := excelize.NewFile()
	defer f.Close()

	// ─── Sheet 1: 信箱（局长信箱渠道） ───
	allLetters, _ := ExportGetLettersInRangeCached(startTime, endTime)

	// 按渠道分组
	var mailboxLetters []model.Letter   // 局长信箱
	var workOrderLetters []model.Letter // 12345工单
	var deskLetters []model.Letter     // 12345公安专席
	var petitionLetters []model.Letter // 信访件/12389子系统

	for _, l := range allLetters {
		switch l.Channel {
		case model.ChannelDirectorMail, 1: // 局长信箱、市民上报
			mailboxLetters = append(mailboxLetters, l)
		case 10: // 12345工单
			workOrderLetters = append(workOrderLetters, l)
		case 11: // 12345公安专席
			deskLetters = append(deskLetters, l)
		case 8: // 12389子系统
			petitionLetters = append(petitionLetters, l)
		default:
			mailboxLetters = append(mailboxLetters, l)
		}
	}

	// 信箱sheet
	createLetterSheet(f, "信箱", mailboxLetters)
	// 工单sheet
	if len(workOrderLetters) > 0 {
		createLetterSheet(f, "工单", workOrderLetters)
	}
	// 专席sheet
	if len(deskLetters) > 0 {
		createLetterSheet(f, "专席", deskLetters)
	}
	// 信访子系统sheet
	if len(petitionLetters) > 0 {
		createLetterSheet(f, "信访子系统", petitionLetters)
	}

	// ─── 日报 sheet ───
	createDailyReportSheet(f, "日报", periodLabel, allLetters)

	// ─── 日报数据统计 sheet ───
	createDailyStatsSheet(f, "日报数据统计", startTime, endTime)

	// 删除默认的 Sheet1（如果还存在）
	for _, sn := range f.GetSheetList() {
		if sn == "Sheet1" || sn == "Sheet" {
			f.DeleteSheet(sn)
		}
	}

	// 保存
	path := filepath.Join(dir, FormatDataSummaryFilename(periodLabel))
	if err := f.SaveAs(path); err != nil {
		return "", fmt.Errorf("save data summary: %w", err)
	}
	return path, nil
}

// createLetterSheet 创建通用的信件列表sheet（26列格式）
func createLetterSheet(f *excelize.File, sheetName string, letters []model.Letter) {
	// 创建新sheet（如果已存在则重命名，不存在则创建）
	existingIdx, _ := f.GetSheetIndex(sheetName)
	if existingIdx > 0 {
		// sheet已存在，不重复创建
	} else {
		// 检查是否有默认Sheet1
		idx1, _ := f.GetSheetIndex("Sheet1")
		if idx1 > 0 {
			f.SetSheetName("Sheet1", sheetName)
		} else {
			f.NewSheet(sheetName)
		}
	}

	// 标题行样式
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Family: FontFamilyDefault, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{HeaderBgColor}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    getDefaultBorder(),
	})

	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border:    getDefaultBorder(),
	})

	// 写表头
	for i, h := range DataSummaryHeaders26 {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 写数据
	for i, l := range letters {
		row := i + 2
		le := BuildLetterExport(l)

		f.SetCellValue(sheetName, coordToCell(1, row), i+1)
		f.SetCellValue(sheetName, coordToCell(2, row), le.LetterNo)
		f.SetCellValue(sheetName, coordToCell(3, row), le.StatusName)
		f.SetCellValue(sheetName, coordToCell(4, row), le.CreatedAt)
		f.SetCellValue(sheetName, coordToCell(5, row), le.ChannelName)
		f.SetCellValue(sheetName, coordToCell(6, row), le.CitizenName)
		f.SetCellValue(sheetName, coordToCell(7, row), le.Phone)
		f.SetCellValue(sheetName, coordToCell(8, row), le.CatL1)
		f.SetCellValue(sheetName, coordToCell(9, row), le.CatL3)
		if le.CatL2 != "" && le.CatL3 != "" {
			f.SetCellValue(sheetName, coordToCell(9, row), le.CatL2+" / "+le.CatL3)
		}
		f.SetCellValue(sheetName, coordToCell(10, row), truncateStr(le.Content, 80))
		f.SetCellValue(sheetName, coordToCell(11, row), le.CountyName)
		f.SetCellValue(sheetName, coordToCell(12, row), le.StationName)
		// 这些字段旧API用固定值填充（参考旧代码 genMonthlySummary）
		f.SetCellValue(sheetName, coordToCell(13, row), "否")        // 重复信件
		f.SetCellValue(sheetName, coordToCell(14, row), "否")        // 退回信件
		f.SetCellValue(sheetName, coordToCell(15, row), "已签收")    // 县局签收
		f.SetCellValue(sheetName, coordToCell(16, row), "已签收")    // 所队签收
		f.SetCellValue(sheetName, coordToCell(17, row), "未回访")    // 回访满意
		f.SetCellValue(sheetName, coordToCell(18, row), "")          // 回访时群众态度
		f.SetCellValue(sheetName, coordToCell(19, row), getOverdueStatus(l)) // 是否逾期
		f.SetCellValue(sheetName, coordToCell(20, row), "")          // 办结后初次信访
		f.SetCellValue(sheetName, coordToCell(21, row), "")          // 初次满意
		f.SetCellValue(sheetName, coordToCell(22, row), "")          // 核查结论
		f.SetCellValue(sheetName, coordToCell(23, row), "")          // 向您通报
		f.SetCellValue(sheetName, coordToCell(24, row), "")          // 扫黑六霸
		f.SetCellValue(sheetName, coordToCell(25, row), "")          // 典型案例
		f.SetCellValue(sheetName, coordToCell(26, row), "")          // 弄虚作假

		// 应用样式
		for c := 1; c <= 26; c++ {
			cell, _ := excelize.CoordinatesToCellName(c, row)
			f.SetCellStyle(sheetName, cell, cell, cellStyle)
		}
	}

	// 列宽
	for col, w := range DefaultColWidths {
		f.SetColWidth(sheetName, col, col, w)
	}
}

// createDailyReportSheet 创建日报sheet
func createDailyReportSheet(f *excelize.File, sheetName, periodLabel string, letters []model.Letter) {
	f.NewSheet(sheetName)

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Family: FontFamilyDefault},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Family: FontFamilyDefault, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{HeaderBgColor}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    getDefaultBorder(),
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border:    getDefaultBorder(),
	})

	// 标题行
	title := fmt.Sprintf("局长信箱日报重点信件反馈统计表（%s）", periodLabel)
	f.MergeCell(sheetName, "A1", "K1")
	f.SetCellValue(sheetName, "A1", title)
	f.SetCellStyle(sheetName, "A1", "K1", titleStyle)

	// 表头
	for i, h := range DailyReportHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheetName, cell, h)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 数据行（取最近的15条作为示例）
	maxRows := len(letters)
	if maxRows > 15 {
		maxRows = 15
	}
	for i := 0; i < maxRows; i++ {
		row := i + 3
		l := letters[i]
		le := BuildLetterExport(l)

		f.SetCellValue(sheetName, cellName(1, row), i+1)
		f.SetCellValue(sheetName, cellName(2, row), l.CreatedAt.Format("2006-01-02"))
		f.SetCellValue(sheetName, cellName(3, row), PlaceholderNoData) // 批示日期
		f.SetCellValue(sheetName, cellName(4, row), PlaceholderNoData) // 批示内容
		f.SetCellValue(sheetName, cellName(5, row), le.Content)        // 基本情况
		f.SetCellValue(sheetName, cellName(6, row), PlaceholderNoData) // 核查结论
		f.SetCellValue(sheetName, cellName(7, row), PlaceholderNoData) // 处置进展
		f.SetCellValue(sheetName, cellName(8, row), PlaceholderNoData) // 反馈情况
		f.SetCellValue(sheetName, cellName(9, row), PlaceholderNoData) // 批示逾期
		f.SetCellValue(sheetName, cellName(10, row), le.CountyName)
		f.SetCellValue(sheetName, cellName(11, row), le.CatL1)

		for c := 1; c <= 11; c++ {
			cell, _ := excelize.CoordinatesToCellName(c, row)
			f.SetCellStyle(sheetName, cell, cell, cellStyle)
		}
	}

	// 列宽
	for col, w := range map[string]float64{"A": 6, "B": 14, "C": 14, "D": 20, "E": 30, "F": 10, "G": 30, "H": 14, "I": 10, "J": 10, "K": 10} {
		f.SetColWidth(sheetName, col, col, w)
	}
}

// createDailyStatsSheet 创建日报数据统计sheet
func createDailyStatsSheet(f *excelize.File, sheetName string, startTime, endTime time.Time) {
	f.NewSheet(sheetName)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 10, Family: FontFamilyDefault, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{HeaderBgColor}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border:    getDefaultBorder(),
	})
	cellStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border:    getDefaultBorder(),
	})

	// 写表头（单行）
	for i := range DailyStatsHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, DailyStatsHeaders[i])
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 按天统计
	stats := aggregateDailyStats(startTime, endTime)
	for i, s := range stats {
		row := i + 2
		// 日期格式：旧API用 MM-DD 格式
		dateStr := ""
		if len(s.date) >= 10 {
			dateStr = s.date[5:10] // "2026-05-07" -> "05-07"
		} else {
			dateStr = s.date
		}
		f.SetCellValue(sheetName, coordToCell(1, row), dateStr)
		f.SetCellValue(sheetName, coordToCell(2, row), s.directorTotal)
		f.SetCellValue(sheetName, coordToCell(3, row), s.directorValid)
		f.SetCellValue(sheetName, coordToCell(4, row), s.workOrderTotal)
		f.SetCellValue(sheetName, coordToCell(5, row), s.workOrderValid)
		f.SetCellValue(sheetName, coordToCell(6, row), s.selfTotal)
		f.SetCellValue(sheetName, coordToCell(7, row), s.selfValid)
		f.SetCellValue(sheetName, coordToCell(8, row), s.sub12389)
		f.SetCellValue(sheetName, coordToCell(9, row), s.sub12337)
		f.SetCellValue(sheetName, coordToCell(10, row), s.petitionTotal)
		f.SetCellValue(sheetName, coordToCell(11, row), s.petitionRepeat)
		f.SetCellValue(sheetName, coordToCell(12, row), s.chiefTotal)
		f.SetCellValue(sheetName, coordToCell(13, row), s.chiefValid)
		f.SetCellValue(sheetName, coordToCell(14, row), "") // 市委书记直通车
		f.SetCellValue(sheetName, coordToCell(15, row), "") // 群众来信
		total := s.directorTotal + s.workOrderTotal + s.selfTotal + s.petitionTotal + s.chiefTotal
		f.SetCellValue(sheetName, coordToCell(16, row), total)

		for c := 1; c <= 16; c++ {
			cell, _ := excelize.CoordinatesToCellName(c, row)
			f.SetCellStyle(sheetName, cell, cell, cellStyle)
		}
	}

	// 列宽
	for col, w := range map[string]float64{"A": 14, "B": 8, "C": 8, "D": 8, "E": 8, "F": 8, "G": 8, "H": 8, "I": 8, "J": 8, "K": 8, "L": 8, "M": 8, "N": 10, "O": 10, "P": 10} {
		f.SetColWidth(sheetName, col, col, w)
	}
}

// ─── 辅助函数 ───

type dailyAgg struct {
	date           string
	directorTotal  int
	directorValid  int
	workOrderTotal int
	workOrderValid int
	selfTotal      int
	selfValid      int
	sub12389       int
	sub12337       int
	petitionTotal  int
	petitionRepeat int
	chiefTotal     int
	chiefValid     int
}

func aggregateDailyStats(startTime, endTime time.Time) []dailyAgg {
	letters, _ := ExportGetLettersInRangeCached(startTime, endTime)
	dayMap := make(map[string]*dailyAgg)

	// 构建日期列表
	current := startTime
	for current.Before(endTime) {
		dateStr := current.Format("2006-01-02")
		dayMap[dateStr] = &dailyAgg{date: dateStr}
		current = current.AddDate(0, 0, 1)
	}

	for _, l := range letters {
		dateStr := l.CreatedAt.Format("2006-01-02")
		d, ok := dayMap[dateStr]
		if !ok {
			continue
		}
		switch l.Channel {
		case 1, model.ChannelDirectorMail: // 市民上报、局长信箱 → 局长信箱列
			d.directorTotal++
			d.directorValid++
		case 10: // 12345工单
			d.workOrderTotal++
			d.workOrderValid++
		case 11: // 12345公安专席 → 12345自接列
			d.selfTotal++
			d.selfValid++
		case 8: // 12389子系统
			d.sub12389++
		}
		// 所有来信都计入信访件和局长收信
		d.petitionTotal++
		if l.Channel == model.ChannelDirectorMail || l.Channel == 1 {
			d.chiefTotal++
			d.chiefValid++
		}
	}

	var result []dailyAgg
	for _, d := range dayMap {
		result = append(result, *d)
	}
	// 日期排序
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].date > result[j].date {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

func getOverdueStatus(l model.Letter) string {
	if l.DeadlineAt != nil && time.Now().After(*l.DeadlineAt) && l.CurrentStatus != model.StatusCodeDone {
		return "是"
	}
	return "否"
}

func getDefaultBorder() []excelize.Border {
	return []excelize.Border{
		{Type: "left", Color: BorderColor, Style: BorderStyle},
		{Type: "right", Color: BorderColor, Style: BorderStyle},
		{Type: "top", Color: BorderColor, Style: BorderStyle},
		{Type: "bottom", Color: BorderColor, Style: BorderStyle},
	}
}

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// coordToCell 列号转单元格名
func coordToCell(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}
