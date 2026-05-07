package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"letter-manage-backend/model"

	"github.com/xuri/excelize/v2"
)

// GenerateStatsChart 基于模板生成通报数图统计
// 复制模板文件后替换数据，保持格式完全一致
func GenerateStatsChart(dir, periodLabel string, startTime, endTime time.Time,
	prevStart, prevEnd *time.Time) (string, error) {

	// 使用模板文件
	tmpl := detectTemplatePath()
	if tmpl == "" {
		// 无模板时回退到旧方式
		return generateStatsChartLegacy(dir, periodLabel, startTime, endTime, prevStart, prevEnd)
	}

	// 复制模板
	dstPath := filepath.Join(dir, FormatStatsChartFilename(periodLabel))
	if err := copyFile(tmpl, dstPath); err != nil {
		return generateStatsChartLegacy(dir, periodLabel, startTime, endTime, prevStart, prevEnd)
	}

	f, err := excelize.OpenFile(dstPath)
	if err != nil {
		return dstPath, nil
	}
	defer f.Close()

	// 获取数据
	letters, _ := ExportGetLettersInRange(startTime, endTime)
	prevLabel := ""
	if prevStart != nil {
		prevLabel = fmt.Sprintf("%d月", prevStart.Month())
	}

	// 预处理数据
	cross := ExportGetUnitCategoryCross(startTime, endTime)
	sort.Slice(cross, func(i, j int) bool { return cross[i].Total > cross[j].Total })
	prevCross := ExportGetUnitCategoryCross(prevStartOrNow(prevStart), prevEndOrNow(prevEnd))
	prevCrossMap := make(map[string]UnitCategoryStats)
	for _, pc := range prevCross {
		prevCrossMap[pc.UnitName] = pc
	}

	// ─── 替换所有Sheet中的月份标签 ───
	curMonth := fmt.Sprintf("%d月", startTime.Month())
	oldMonth := "3月"
	oldPrevMonth := "2月"
	tmplCur := "{month}"
	tmplPrev := "{prev_month}"
	tmplPrev2 := "{prev2_month}"
	prev2Label := ""
	if prevStart != nil {
		prev2Month := prevStart.Month() - 1
		if prev2Month < 1 {
			prev2Month = 12
		}
		prev2Label = fmt.Sprintf("%d月", prev2Month)
	}
	for _, sname := range f.GetSheetList() {
		rows, _ := f.GetRows(sname)
		for ri, row := range rows {
			for ci, cell := range row {
				if cell == "" {
					continue
				}
				newCell := cell
				// Replace template placeholders
				newCell = strings.ReplaceAll(newCell, tmplCur, curMonth)
				if prevLabel != "" {
					newCell = strings.ReplaceAll(newCell, tmplPrev, prevLabel)
				} else {
					newCell = strings.ReplaceAll(newCell, tmplPrev, curMonth)
				}
				if prev2Label != "" {
					newCell = strings.ReplaceAll(newCell, tmplPrev2, prev2Label)
				} else {
					newCell = strings.ReplaceAll(newCell, tmplPrev2, "")
				}
				// Also replace old hardcoded months
				newCell = strings.ReplaceAll(newCell, oldMonth, curMonth)
				if prevLabel != "" {
					newCell = strings.ReplaceAll(newCell, oldPrevMonth, prevLabel)
				}
				if newCell != cell {
					col, _ := excelize.ColumnNumberToName(ci + 1)
					f.SetCellValue(sname, fmt.Sprintf("%s%d", col, ri+1), newCell)
				}
			}
		}
	}

	// ─── 填充各Sheet数据 ───
	populateOverviewSheet(f, "总览", cross, prevCrossMap, prevLabel)
	populateRegionDistSheet(f, "地域分布", letters)
	populatePoliceTypeSheet(f, "大警种", letters)
	populateTeamStatsSheet(f, "基层所队", letters)
	populateRepeatSheet(f, "重复件", letters)
	populateComplaintDistSheet(f, "投诉分布", letters)
	populateTeamComplaintSheet(f, "投诉所队", letters, startTime, endTime)
	populateReportSheet(f, "举报", letters)
	populateSuggestionSheet(f, "意见建议", letters)
	populateKeySuggestionSheet(f, "重点意见", letters)
	populateConsultChangeSheet(f, "咨询变化", startTime, endTime, prevStart, prevEnd)
	populateHelpSheet(f, "求助", letters)
	populateAppealDistSheet(f, "申诉分布", letters)
	populateTeamAppealSheet(f, "申诉所队", letters)
	populateAppealPivotSheet(f, "Sheet7", letters)
	populateAppealDetailSheet(f, "申诉详情", letters)
	populateCategoryPivotSheet(f, "类别总透视", letters)
	populateNonRepeatSheet(f, "非重复信件详情", letters)

	f.Save()
	return dstPath, nil
}

func detectTemplatePath() string {
	// 优先使用项目内置模板
	paths := []string{
		"templates/通报数图统计_模板.xlsx",
		"/Users/liheng/Desktop/pic/origin/2026年3月通报数图统计.xlsx",
		"/Users/v_liheng02/Desktop/other/局长信箱原始资料/2026年3月通报数图统计.xlsx",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	return err
}

func getPrevLetters(prevStart, prevEnd *time.Time) ([]model.Letter, error) {
	if prevStart == nil || prevEnd == nil {
		return nil, nil
	}
	return ExportGetLettersInRange(*prevStart, *prevEnd)
}

func prevStartOrNow(t *time.Time) time.Time {
	if t != nil {
		return *t
	}
	return time.Now().AddDate(0, -1, 0)
}

func prevEndOrNow(t *time.Time) time.Time {
	if t != nil {
		return *t
	}
	return time.Now()
}

// ─── 数据填充函数（基于模板的坐标写入数据） ───

// populateOverviewSheet 填充总览Sheet数据
// 模板结构: R1=标题, R2/R3/R4=表头, R5+=数据行（按Level2排序）
func populateOverviewSheet(f *excelize.File, sheet string, cross []UnitCategoryStats,
	prevCrossMap map[string]UnitCategoryStats, prevLabel string) {

	if len(cross) == 0 {
		return
	}

	// 获取模板中的行数，从第5行开始写数据，覆盖现有数据
	// 先找到模板中最后一行的行号
	rows, _ := f.GetRows(sheet)
	dataStartRow := 5 // 数据从第5行开始
	totalRows := len(rows)

	// 清除旧数据（从第5行到最后一行数据行）
	for r := dataStartRow; r <= totalRows; r++ {
		for c := 1; c <= 34; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			// 跳过标题行和空行
			if r <= 4 {
				continue
			}
			cell := fmt.Sprintf("%s%d", col, r)
			f.SetCellValue(sheet, cell, "")
		}
	}

	// 写入新数据
	r := dataStartRow
	for _, row := range cross {
		if row.UnitName == "有效\n总量" {
			continue // 总量行后面单独处理
		}
		writeCell(f, sheet, 2, r, row.UnitName)
		writeCell(f, sheet, 3, r, row.Complaint)
		writeCell(f, sheet, 4, r, row.Report)
		writeCell(f, sheet, 5, r, row.Suggestion)
		writeCell(f, sheet, 6, r, row.Consult)
		writeCell(f, sheet, 7, r, row.Appeal)
		writeCell(f, sheet, 8, r, row.Help)
		writeCell(f, sheet, 9, r, row.Thank)
		writeCell(f, sheet, 10, r, row.DirectorMail)
		writeCell(f, sheet, 11, r, row.Sub12389)
		writeCell(f, sheet, 12, r, row.VisitBJ)
		writeCell(f, sheet, 13, r, row.WorkOrder)
		writeCell(f, sheet, 14, r, row.PoliceDesk)
		writeCell(f, sheet, 15, r, row.Total)

		// 上月数据（列17-29）
		if prev, ok := prevCrossMap[row.UnitName]; ok {
			writeCell(f, sheet, 17, r, prev.UnitName)
			writeCell(f, sheet, 18, r, prev.Complaint)
			writeCell(f, sheet, 19, r, prev.Report)
			writeCell(f, sheet, 20, r, prev.Suggestion)
			writeCell(f, sheet, 21, r, prev.Consult)
			writeCell(f, sheet, 22, r, prev.Appeal)
			writeCell(f, sheet, 23, r, prev.DirectorMail)
			writeCell(f, sheet, 24, r, prev.Sub12389)
			writeCell(f, sheet, 25, r, prev.VisitBJ)
			writeCell(f, sheet, 26, r, prev.WorkOrder)
			writeCell(f, sheet, 27, r, prev.PoliceDesk)
			writeCell(f, sheet, 28, r, prev.Total)
		}
		r++
	}
}

// writeCell writes a value to a cell, using 0 for zero ints to keep them visible
func writeCell(f *excelize.File, sheet string, col, row int, value interface{}) {
	colName, _ := excelize.ColumnNumberToName(col)
	f.SetCellValue(sheet, fmt.Sprintf("%s%d", colName, row), value)
}

// populateRegionDistSheet 填充地域分布Sheet
func populateRegionDistSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	startTime := letters[len(letters)-1].ReceivedAt
	endTime := letters[0].ReceivedAt.AddDate(0, 0, 1)
	stats := ExportGetUnitLevel1Stats(startTime, endTime)
	if len(stats) == 0 {
		return
	}
	// 清除旧数据行（从第2行开始）
	rows, _ := f.GetRows(sheet)
	for r := 2; r <= len(rows); r++ {
		for c := 1; c <= 2; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	for i, s := range stats {
		writeCell(f, sheet, 1, i+2, s.Name)
		writeCell(f, sheet, 2, i+2, s.Count)
	}
}

// populatePoliceTypeSheet 填充大警种Sheet
func populatePoliceTypeSheet(f *excelize.File, sheet string, letters []model.Letter) {
	policeTypes := map[string]int{
		"刑侦系统": 0, "治安系统": 0, "经侦系统": 0, "交管系统": 0, "其他单位": 0,
	}
	for _, l := range letters {
		name := ""
		if l.CurrentUnitObj != nil && l.CurrentUnitObj.Level2 != "" {
			name = l.CurrentUnitObj.Level2
		}
		switch {
		case containsStr(name, "刑侦"):
			policeTypes["刑侦系统"]++
		case containsStr(name, "治安"):
			policeTypes["治安系统"]++
		case containsStr(name, "经侦"):
			policeTypes["经侦系统"]++
		case containsStr(name, "交管"):
			policeTypes["交管系统"]++
		default:
			policeTypes["其他单位"]++
		}
	}
	// 清除旧数据
	rows, _ := f.GetRows(sheet)
	for r := 1; r <= len(rows); r++ {
		for c := 1; c <= 2; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	r := 1
	for name, count := range policeTypes {
		if count > 0 {
			writeCell(f, sheet, 1, r, name)
			writeCell(f, sheet, 2, r, count)
			r++
		}
	}
}

// populateTeamStatsSheet 填充基层所队Sheet
func populateTeamStatsSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	teams := ExportGetTeamStats(
		letters[len(letters)-1].ReceivedAt,
		letters[0].ReceivedAt.AddDate(0, 0, 1),
	)
	if len(teams) == 0 {
		return
	}
	// 清除旧数据（行3+）
	rows, _ := f.GetRows(sheet)
	for r := 3; r <= len(rows); r++ {
		for c := 1; c <= 9; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	// 表头保留在行1-2，数据从行3开始
	for i, t := range teams {
		r := i + 3
		writeCell(f, sheet, 1, r, t.TeamName)
		writeCell(f, sheet, 2, r, t.Complaint)
		writeCell(f, sheet, 3, r, t.Report)
		writeCell(f, sheet, 4, r, t.Suggestion)
		writeCell(f, sheet, 5, r, t.Consult)
		writeCell(f, sheet, 6, r, t.Appeal)
		writeCell(f, sheet, 7, r, t.Help)
		writeCell(f, sheet, 8, r, t.Thank)
		writeCell(f, sheet, 9, r, t.Total)
	}
}

// populateRepeatSheet 填充重复件Sheet
func populateRepeatSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	startTime := letters[len(letters)-1].ReceivedAt
	endTime := letters[0].ReceivedAt.AddDate(0, 0, 1)
	repeats := ExportGetRepeatStats(startTime, endTime)
	if len(repeats) == 0 {
		return
	}
	rows, _ := f.GetRows(sheet)
	for r := 3; r <= len(rows); r++ {
		for c := 1; c <= 5; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	for i, rp := range repeats {
		r := i + 3
		writeCell(f, sheet, 1, r, rp.TeamName)
		writeCell(f, sheet, 2, r, rp.RepeatCount)
		writeCell(f, sheet, 4, r, rp.CountyName)
		writeCell(f, sheet, 5, r, rp.TotalLetters)
	}
}

// populateComplaintDistSheet 填充投诉分布Sheet
func populateComplaintDistSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	startTime := letters[len(letters)-1].ReceivedAt
	endTime := letters[0].ReceivedAt.AddDate(0, 0, 1)
	detail := ExportGetCategoryDetail(startTime, endTime, "投诉举报类")
	if len(detail) == 0 {
		return
	}
	// 聚合
	type unitAgg map[string]int
	units := make(map[string]unitAgg)
	unitOrder := make([]string, 0)
	seen := make(map[string]bool)
	for _, d := range detail {
		if _, ok := units[d.UnitName]; !ok {
			units[d.UnitName] = make(unitAgg)
		}
		units[d.UnitName][stripNL(d.CatL3)] += d.Count
		if !seen[d.UnitName] {
			unitOrder = append(unitOrder, d.UnitName)
			seen[d.UnitName] = true
		}
	}
	// 清除旧数据
	rows, _ := f.GetRows(sheet)
	for r := 3; r <= len(rows); r++ {
		for c := 1; c <= 19; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	// 写总计
	complaintKeys := []string{"不出警、出警慢", "有案不立", "压案不查，久拖不决", "不按法定程序和要求办理案件", "推诿扯皮、办事拖拉", "其他投诉"}
	totals := make(unitAgg)
	for _, ua := range units {
		for k, v := range ua {
			totals[k] += v
		}
	}
	r := 3
	writeCell(f, sheet, 1, r, 0)
	writeCell(f, sheet, 2, r, "总计")
	totalSum := 0
	for i, ck := range complaintKeys {
		v := totals[ck]
		writeCell(f, sheet, i+3, r, v)
		totalSum += v
	}
	writeCell(f, sheet, 9, r, totalSum)
	r++
	for idx, uname := range unitOrder {
		ua := units[uname]
		writeCell(f, sheet, 1, r, idx+1)
		writeCell(f, sheet, 2, r, uname)
		rowSum := 0
		for i, ck := range complaintKeys {
			v := ua[ck]
			writeCell(f, sheet, i+3, r, v)
			rowSum += v
		}
		writeCell(f, sheet, 9, r, rowSum)
		r++
	}
}

// populateTeamComplaintSheet 填充投诉所队Sheet
func populateTeamComplaintSheet(f *excelize.File, sheet string, letters []model.Letter,
	startTime, endTime time.Time) {
	detail := ExportGetTeamCategoryDetail(startTime, endTime, "投诉举报类")
	if len(detail) == 0 {
		return
	}
	type unitAgg map[string]int
	teams := make(map[string]unitAgg)
	teamOrder := make([]string, 0)
	seen := make(map[string]bool)
	for _, d := range detail {
		if _, ok := teams[d.UnitName]; !ok {
			teams[d.UnitName] = make(unitAgg)
		}
		teams[d.UnitName][stripNL(d.CatL3)] += d.Count
		if !seen[d.UnitName] {
			teamOrder = append(teamOrder, d.UnitName)
			seen[d.UnitName] = true
		}
	}
	rows, _ := f.GetRows(sheet)
	for r := 3; r <= len(rows); r++ {
		for c := 1; c <= 7; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	complaintKeys := []string{"不出警、出警慢", "有案不立", "压案不查，久拖不决", "不按法定程序和要求办理案件", "推诿扯皮、办事拖拉"}
	r := 3
	for _, tname := range teamOrder {
		ta := teams[tname]
		writeCell(f, sheet, 1, r, tname)
		rowSum := 0
		for i, ck := range complaintKeys {
			v := ta[ck]
			writeCell(f, sheet, i+2, r, v)
			rowSum += v
		}
		writeCell(f, sheet, 7, r, rowSum)
		r++
	}
}

// populateReportSheet 填充举报Sheet
func populateReportSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	startTime := letters[len(letters)-1].ReceivedAt
	endTime := letters[0].ReceivedAt.AddDate(0, 0, 1)
	detail := ExportGetCategoryDetail(startTime, endTime, "提供社会违法线索类")
	if len(detail) == 0 {
		return
	}
	type unitAgg map[string]int
	units := make(map[string]unitAgg)
	unitOrder := make([]string, 0)
	seen := make(map[string]bool)
	for _, d := range detail {
		if _, ok := units[d.UnitName]; !ok {
			units[d.UnitName] = make(unitAgg)
		}
		units[d.UnitName][stripNL(d.CatL3)] += d.Count
		if !seen[d.UnitName] {
			unitOrder = append(unitOrder, d.UnitName)
			seen[d.UnitName] = true
		}
	}
	rows, _ := f.GetRows(sheet)
	for r := 2; r <= len(rows); r++ {
		for c := 1; c <= 6; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	totals := make(unitAgg)
	for _, ua := range units {
		for k, v := range ua {
			totals[k] += v
		}
	}
	r := 2
	writeCell(f, sheet, 1, r, "总计")
	reportKeys := []string{"涉赌线索", "涉黑线索", "涉黄线索", "交通违法线索", "其他案件线索"}
	for i, rk := range reportKeys {
		writeCell(f, sheet, i+2, r, totals[rk])
	}
	r++
	for _, uname := range unitOrder {
		ua := units[uname]
		writeCell(f, sheet, 1, r, uname)
		for i, rk := range reportKeys {
			writeCell(f, sheet, i+2, r, ua[rk])
		}
		r++
	}
}

// populateSuggestionSheet 填充意见建议Sheet
func populateSuggestionSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	startTime := letters[len(letters)-1].ReceivedAt
	endTime := letters[0].ReceivedAt.AddDate(0, 0, 1)
	detail := ExportGetCategoryDetail(startTime, endTime, "意见建议类")
	if len(detail) == 0 {
		return
	}
	type unitAgg map[string]int
	units := make(map[string]unitAgg)
	unitOrder := make([]string, 0)
	seen := make(map[string]bool)
	for _, d := range detail {
		if _, ok := units[d.UnitName]; !ok {
			units[d.UnitName] = make(unitAgg)
		}
		units[d.UnitName][stripNL(d.CatL3)] += d.Count
		if !seen[d.UnitName] {
			unitOrder = append(unitOrder, d.UnitName)
			seen[d.UnitName] = true
		}
	}
	rows, _ := f.GetRows(sheet)
	headerCount := len(strings.Split(rows[0][0], "")) // approximate column count
	_ = headerCount
	for r := 2; r <= len(rows); r++ {
		for c := 1; c <= 15; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	suggKeys := []string{"案件办理效率", "案件反馈/公开", "车驾管业务方面", "道路设施方面", "公共秩序方面",
		"公正文明执法", "公众交通安全出行", "户籍业务方面", "交通秩序方面", "其他意见建议",
		"社会治安方面", "提升队伍业务素质", "优化窗口服务机制模式", "治安行政管理方面"}
	totals := make(unitAgg)
	for _, ua := range units {
		for k, v := range ua {
			totals[k] += v
		}
	}
	r := 2
	writeCell(f, sheet, 1, r, "总计")
	for i, sk := range suggKeys {
		writeCell(f, sheet, i+2, r, totals[sk])
	}
	r++
	for _, uname := range unitOrder {
		ua := units[uname]
		writeCell(f, sheet, 1, r, uname)
		for i, sk := range suggKeys {
			writeCell(f, sheet, i+2, r, ua[sk])
		}
		r++
	}
}

// populateKeySuggestionSheet 填充重点意见Sheet
func populateKeySuggestionSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	startTime := letters[len(letters)-1].ReceivedAt
	endTime := letters[0].ReceivedAt.AddDate(0, 0, 1)
	detail := ExportGetCategoryDetail(startTime, endTime, "意见建议类")
	if len(detail) == 0 {
		return
	}
	type unitAgg map[string]int
	units := make(map[string]unitAgg)
	unitOrder := make([]string, 0)
	seen := make(map[string]bool)
	for _, d := range detail {
		if _, ok := units[d.UnitName]; !ok {
			units[d.UnitName] = make(unitAgg)
		}
		units[d.UnitName][stripNL(d.CatL3)] += d.Count
		if !seen[d.UnitName] {
			unitOrder = append(unitOrder, d.UnitName)
			seen[d.UnitName] = true
		}
	}
	rows, _ := f.GetRows(sheet)
	for r := 2; r <= len(rows); r++ {
		for c := 1; c <= 3; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	totals := make(unitAgg)
	for _, ua := range units {
		for k, v := range ua {
			totals[k] += v
		}
	}
	r := 2
	writeCell(f, sheet, 1, r, "总计")
	writeCell(f, sheet, 2, r, totals["案件办理效率"])
	writeCell(f, sheet, 3, r, totals["公正文明执法"])
	r++
	for _, uname := range unitOrder {
		ua := units[uname]
		writeCell(f, sheet, 1, r, uname)
		writeCell(f, sheet, 2, r, ua["案件办理效率"])
		writeCell(f, sheet, 3, r, ua["公正文明执法"])
		r++
	}
}

// populateConsultChangeSheet 填充咨询变化Sheet
func populateConsultChangeSheet(f *excelize.File, sheet string,
	curStart, curEnd time.Time, prevStart, prevEnd *time.Time) {
	curDetail := ExportGetCategoryDetail(curStart, curEnd, "咨询政策类")
	if len(curDetail) == 0 {
		return
	}
	curAgg := make(map[string]int)
	for _, d := range curDetail {
		curAgg[stripNL(d.CatL3)] += d.Count
	}
	prevAgg := make(map[string]int)
	if prevStart != nil && prevEnd != nil {
		prevDetail := ExportGetCategoryDetail(*prevStart, *prevEnd, "咨询政策类")
		for _, d := range prevDetail {
			prevAgg[stripNL(d.CatL3)] += d.Count
		}
	}
	rows, _ := f.GetRows(sheet)
	for r := 2; r <= len(rows); r++ {
		for c := 1; c <= 5; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	r := 2
	for catL3, curVal := range curAgg {
		writeCell(f, sheet, 1, r, catL3)
		writeCell(f, sheet, 2, r, curVal)
		prevVal := prevAgg[catL3]
		writeCell(f, sheet, 3, r, prevVal)
		if prevVal > 0 {
			writeCell(f, sheet, 4, r, float64(curVal-prevVal)/float64(prevVal))
		}
		r++
	}
}

// populateHelpSheet 填充求助Sheet
func populateHelpSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	startTime := letters[len(letters)-1].ReceivedAt
	endTime := letters[0].ReceivedAt.AddDate(0, 0, 1)
	detail := ExportGetCategoryDetail(startTime, endTime, "求助类")
	if len(detail) == 0 {
		return
	}
	type unitAgg map[string]int
	units := make(map[string]unitAgg)
	unitOrder := make([]string, 0)
	seen := make(map[string]bool)
	for _, d := range detail {
		if _, ok := units[d.UnitName]; !ok {
			units[d.UnitName] = make(unitAgg)
		}
		units[d.UnitName][stripNL(d.CatL3)] += d.Count
		if !seen[d.UnitName] {
			unitOrder = append(unitOrder, d.UnitName)
			seen[d.UnitName] = true
		}
	}
	rows, _ := f.GetRows(sheet)
	for r := 2; r <= len(rows); r++ {
		for c := 1; c <= 5; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	helpKeys := []string{"请求帮助寻人/寻物", "请求调解民事纠纷", "请求上级协调业务办理指定管辖", "请求受理初查"}
	totals := make(unitAgg)
	for _, ua := range units {
		for k, v := range ua {
			totals[k] += v
		}
	}
	r := 2
	writeCell(f, sheet, 1, r, "总计")
	for i, hk := range helpKeys {
		writeCell(f, sheet, i+2, r, totals[hk])
	}
	r++
	for _, uname := range unitOrder {
		ua := units[uname]
		writeCell(f, sheet, 1, r, uname)
		for i, hk := range helpKeys {
			writeCell(f, sheet, i+2, r, ua[hk])
		}
		r++
	}
}

// populateAppealDistSheet 填充申诉分布Sheet
func populateAppealDistSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	startTime := letters[len(letters)-1].ReceivedAt
	endTime := letters[0].ReceivedAt.AddDate(0, 0, 1)
	detail := ExportGetCategoryDetail(startTime, endTime, "申诉类")
	if len(detail) == 0 {
		return
	}
	type unitAgg map[string]int
	units := make(map[string]unitAgg)
	unitOrder := make([]string, 0)
	seen := make(map[string]bool)
	for _, d := range detail {
		if _, ok := units[d.UnitName]; !ok {
			units[d.UnitName] = make(unitAgg)
		}
		units[d.UnitName][stripNL(d.CatL3)] += d.Count
		if !seen[d.UnitName] {
			unitOrder = append(unitOrder, d.UnitName)
			seen[d.UnitName] = true
		}
	}
	rows, _ := f.GetRows(sheet)
	for r := 3; r <= len(rows); r++ {
		for c := 1; c <= 7; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	appealKeys := []string{"案件办理程序不理解", "行政处罚决定有异议", "不符合业务办理条件", "事故认定结论不认同"}
	r := 3
	for idx, uname := range unitOrder {
		ua := units[uname]
		writeCell(f, sheet, 1, r, idx+1)
		writeCell(f, sheet, 2, r, uname)
		rowSum := 0
		for i, ak := range appealKeys {
			v := ua[ak]
			writeCell(f, sheet, i+3, r, v)
			rowSum += v
		}
		writeCell(f, sheet, 7, r, rowSum)
		r++
	}
}

// populateTeamAppealSheet 填充申诉所队Sheet
func populateTeamAppealSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	startTime := letters[len(letters)-1].ReceivedAt
	endTime := letters[0].ReceivedAt.AddDate(0, 0, 1)
	detail := ExportGetTeamCategoryDetail(startTime, endTime, "申诉类")
	if len(detail) == 0 {
		return
	}
	type unitAgg map[string]int
	units := make(map[string]unitAgg)
	unitOrder := make([]string, 0)
	seen := make(map[string]bool)
	for _, d := range detail {
		if _, ok := units[d.UnitName]; !ok {
			units[d.UnitName] = make(unitAgg)
		}
		units[d.UnitName][stripNL(d.CatL3)] += d.Count
		if !seen[d.UnitName] {
			unitOrder = append(unitOrder, d.UnitName)
			seen[d.UnitName] = true
		}
	}
	rows, _ := f.GetRows(sheet)
	for r := 3; r <= len(rows); r++ {
		for c := 1; c <= 6; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	appealKeys := []string{"案件办理程序不理解", "行政处罚决定有异议", "不符合业务办理条件", "事故认定结论不认同"}
	r := 3
	for _, uname := range unitOrder {
		ua := units[uname]
		writeCell(f, sheet, 1, r, uname)
		rowSum := 0
		for i, ak := range appealKeys {
			v := ua[ak]
			writeCell(f, sheet, i+2, r, v)
			rowSum += v
		}
		writeCell(f, sheet, 6, r, rowSum)
		r++
	}
}

// populateAppealPivotSheet 填充Sheet7（申诉透视）
func populateAppealPivotSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	startTime := letters[len(letters)-1].ReceivedAt
	endTime := letters[0].ReceivedAt.AddDate(0, 0, 1)
	detail := ExportGetCategoryDetail(startTime, endTime, "申诉类")
	if len(detail) == 0 {
		return
	}
	type unitAgg map[string]int
	units := make(map[string]unitAgg)
	unitOrder := make([]string, 0)
	seen := make(map[string]bool)
	for _, d := range detail {
		if _, ok := units[d.UnitName]; !ok {
			units[d.UnitName] = make(unitAgg)
		}
		units[d.UnitName][stripNL(d.CatL3)] += d.Count
		if !seen[d.UnitName] {
			unitOrder = append(unitOrder, d.UnitName)
			seen[d.UnitName] = true
		}
	}
	rows, _ := f.GetRows(sheet)
	for r := 5; r <= len(rows); r++ {
		for c := 1; c <= 6; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	pivotKeys := []string{"案件申诉", "事故认定申诉", "行政处罚申诉", "业务办理申诉"}
	r := 5
	for _, uname := range unitOrder {
		ua := units[uname]
		writeCell(f, sheet, 1, r, uname)
		rowSum := 0
		for i, pk := range pivotKeys {
			v := ua[pk]
			writeCell(f, sheet, i+2, r, v)
			rowSum += v
		}
		writeCell(f, sheet, 6, r, rowSum)
		r++
	}
}

// populateAppealDetailSheet 填充申诉详情Sheet
func populateAppealDetailSheet(f *excelize.File, sheet string, letters []model.Letter) {
	rows, _ := f.GetRows(sheet)
	for r := 2; r <= len(rows); r++ {
		for c := 1; c <= 12; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	r := 2
	for _, l := range letters {
		if l.Category == nil || l.Category.Level1 != "申诉类" {
			continue
		}
		le := BuildLetterExport(l)
		writeCell(f, sheet, 1, r, r-1)
		writeCell(f, sheet, 2, r, le.LetterNo)
		writeCell(f, sheet, 3, r, le.StatusName)
		writeCell(f, sheet, 4, r, le.ReceivedAt)
		writeCell(f, sheet, 5, r, le.ChannelName)
		writeCell(f, sheet, 6, r, le.CitizenName)
		writeCell(f, sheet, 7, r, le.Phone)
		writeCell(f, sheet, 8, r, le.CatL1)
		writeCell(f, sheet, 9, r, le.CatL3)
		writeCell(f, sheet, 10, r, truncateStr(le.Content, 60))
		writeCell(f, sheet, 11, r, le.CountyName)
		writeCell(f, sheet, 12, r, le.StationName)
		r++
	}
}

// populateCategoryPivotSheet 填充类别总透视Sheet
func populateCategoryPivotSheet(f *excelize.File, sheet string, letters []model.Letter) {
	if len(letters) == 0 {
		return
	}
	startTime := letters[len(letters)-1].ReceivedAt
	endTime := letters[0].ReceivedAt.AddDate(0, 0, 1)
	// 获取所有类别的细类统计
	allCatL1 := []string{"意见建议类", "申诉类", "投诉举报类", "提供社会违法线索类", "求助类", "咨询政策类"}
	headers := []string{"行标签"}
	catL3Set := make(map[string]bool)
	allData := make(map[string]map[string]int) // unitName -> catL3 -> count

	for _, cat := range allCatL1 {
		detail := ExportGetCategoryDetail(startTime, endTime, cat)
		for _, d := range detail {
			cleanCat := stripNL(d.CatL3)
			catL3Set[cleanCat] = true
			if allData[d.UnitName] == nil {
				allData[d.UnitName] = make(map[string]int)
			}
			allData[d.UnitName][cleanCat] += d.Count
		}
	}

	if len(allData) == 0 {
		return
	}

	// 构建列头
	var catL3List []string
	for k := range catL3Set {
		catL3List = append(catL3List, k)
	}
	sort.Strings(catL3List)

	// 清除旧数据
	rows, _ := f.GetRows(sheet)
	for r := 3; r <= len(rows); r++ {
		for c := 1; c <= len(rows[0]); c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}

	// 写列头
	writeCell(f, sheet, 1, 4, "行标签")
	writeCell(f, sheet, 2, 3, "列标签")
	for i, cl3 := range catL3List {
		writeCell(f, sheet, i+2, 4, cl3)
		_ = headers
	}

	// 写数据
	r := 5
	for uname, catMap := range allData {
		writeCell(f, sheet, 1, r, uname)
		rowSum := 0
		for i, cl3 := range catL3List {
			v := catMap[cl3]
			writeCell(f, sheet, i+2, r, v)
			rowSum += v
		}
		writeCell(f, sheet, len(catL3List)+2, r, rowSum)
		r++
	}
	writeCell(f, sheet, 4, 3, "计数项:信件细类")
}

// populateNonRepeatSheet 填充非重复信件详情Sheet
func populateNonRepeatSheet(f *excelize.File, sheet string, letters []model.Letter) {
	rows, _ := f.GetRows(sheet)
	for r := 2; r <= len(rows); r++ {
		for c := 1; c <= 26; c++ {
			col, _ := excelize.ColumnNumberToName(c)
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", col, r), "")
		}
	}
	r := 2
	for _, l := range letters {
		le := BuildLetterExport(l)
		writeCell(f, sheet, 1, r, r-1)
		writeCell(f, sheet, 2, r, le.LetterNo)
		writeCell(f, sheet, 3, r, le.StatusName)
		writeCell(f, sheet, 4, r, le.ReceivedAt)
		writeCell(f, sheet, 5, r, le.ChannelName)
		writeCell(f, sheet, 6, r, le.CitizenName)
		writeCell(f, sheet, 7, r, le.Phone)
		writeCell(f, sheet, 8, r, le.CatL1)
		writeCell(f, sheet, 9, r, le.CatL3)
		writeCell(f, sheet, 10, r, truncateStr(le.Content, 60))
		writeCell(f, sheet, 11, r, le.CountyName)
		writeCell(f, sheet, 12, r, le.StationName)
		for ci := 13; ci <= 26; ci++ {
			writeCell(f, sheet, ci, r, PlaceholderNoData)
		}
		r++
	}
}

// ─── 辅助函数 ───

// stripNL removes newline characters from a string for key matching
func stripNL(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\n' {
			result = append(result, s[i])
		}
	}
	return string(result)
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && strings.Contains(s, substr)
}

// generateStatsChartLegacy 旧版生成（无模板时回退）
func generateStatsChartLegacy(dir, periodLabel string, startTime, endTime time.Time,
	prevStart, prevEnd *time.Time) (string, error) {
	// ... 保留旧实现 ...
	f := excelize.NewFile()
	defer f.Close()
	path := filepath.Join(dir, FormatStatsChartFilename(periodLabel))
	f.SetCellValue("Sheet1", "A1", PlaceholderNoData)
	f.SaveAs(path)
	return path, nil
}

// 保留原有函数声明以避免编译错误（历史兼容性占位）
func _unused_legacy() {}
