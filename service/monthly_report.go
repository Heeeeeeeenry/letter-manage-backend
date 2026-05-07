package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"

	"github.com/go-pdf/fpdf"
	"github.com/xuri/excelize/v2"
)

// LetterExport 导出数据结构
type LetterExport struct {
	LetterNo    string
	CitizenName string
	Phone       string
	ReceivedAt  string
	ChannelName string
	StatusName  string
	StatusCode  int
	Content     string
	CatL1       string
	CatL2       string
	CatL3       string
	UnitName    string
	CountyName  string
	StationName string
}

// ─── Main Entry ───

func GenerateMonthlyReportZip(permLevel string, unitID *uint, period string) (string, error) {
	now := time.Now()
	var startTime, endTime time.Time
	var periodLabel string

	switch period {
	case "day":
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endTime = startTime.Add(24 * time.Hour)
		periodLabel = "今日"
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		startTime = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
		endTime = startTime.Add(7 * 24 * time.Hour)
		periodLabel = "本周"
	case "month":
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endTime = startTime.AddDate(0, 1, 0)
		periodLabel = fmt.Sprintf("%d月", int(now.Month()))
	case "year":
		startTime = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		endTime = startTime.AddDate(1, 0, 0)
		periodLabel = fmt.Sprintf("%d年", now.Year())
	default:
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endTime = startTime.AddDate(0, 1, 0)
		periodLabel = fmt.Sprintf("%d月", int(now.Month()))
	}

	var realLetters []model.Letter
	dao.DB.Preload("Category").Preload("CurrentUnitObj").
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Order("created_at DESC").Limit(9999).Find(&realLetters)

	var realData []LetterExport
	for _, l := range realLetters {
		realData = append(realData, buildExport(l))
	}

	fake := NewFakeDataSet()
	fakeData := fake.GenerateFakeLetters(100)

	tmpDir, err := os.MkdirTemp("", "monthly_report_")
	if err != nil {
		return "", err
	}

	files := make(map[string]string)

	path1, err := genMonthlySummary(tmpDir, periodLabel, realData, fakeData)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}
	files[fmt.Sprintf("%s数据汇总.xlsx", periodLabel)] = path1

	path2, err := genStatisticsChart(tmpDir, periodLabel, periodLabel, realData, fakeData)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}
	files[fmt.Sprintf("%s通报数图统计.xlsx", periodLabel)] = path2

	for _, rt := range []string{"通报", "质态分析报告"} {
		p, err := genPDFStub(tmpDir, periodLabel, rt)
		if err != nil {
			os.RemoveAll(tmpDir)
			return "", err
		}
		files[fmt.Sprintf("%s%s.pdf", periodLabel, rt)] = p
	}

	path5, err := copyClassificationStandard(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}
	files["20260310全平台数据分类标准.xlsx"] = path5

	zipPath := filepath.Join(tmpDir, fmt.Sprintf("%s导出.zip", periodLabel))
	zipFile, err := os.Create(zipPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}
	zipWriter := zip.NewWriter(zipFile)
	for name, path := range files {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		w, err := zipWriter.Create(name)
		if err != nil {
			f.Close()
			continue
		}
		io.Copy(w, f)
		f.Close()
	}
	zipWriter.Close()
	zipFile.Close()
	return zipPath, nil
}

func buildExport(l model.Letter) LetterExport {
	ed := LetterExport{
		LetterNo:    l.LetterNo,
		CitizenName: l.CitizenName,
		Phone:       l.Phone,
		ReceivedAt:  l.ReceivedAt.Format("2006-01-02 15:04:05"),
		ChannelName: l.GetChannelName(),
		StatusName:  l.GetStatusName(),
		StatusCode:  int(l.CurrentStatus),
		Content:     l.Content,
	}
	if l.Category != nil {
		ed.CatL1 = l.Category.Level1
		ed.CatL2 = l.Category.Level2
		ed.CatL3 = l.Category.Level3
	}
	if l.CurrentUnitObj != nil {
		parts := []string{l.CurrentUnitObj.Level1, l.CurrentUnitObj.Level2, l.CurrentUnitObj.Level3}
		var nonEmpty []string
		for _, p := range parts {
			if p != "" {
				nonEmpty = append(nonEmpty, p)
			}
		}
		ed.UnitName = strings.Join(nonEmpty, " / ")
	}
	return ed
}

// ─── 1. 数据汇总 ───

func genMonthlySummary(dir, periodLabel string, realData, fakeData []LetterExport) (string, error) {
	f := excelize.NewFile()
	defer f.Close()
	f.SetSheetName("Sheet1", "信箱")

	headers26 := []string{
		"序号", "信件编号", "信件状态", "来信时间", "来信渠道", "群众姓名", "手机号码",
		"信件类别", "信件细类", "简要诉求", "分县局", "主办单位",
		"重复信件", "退回信件", "县局签收", "所队签收",
		"回访满意", "回访时群众态度", "是否逾期",
		"办结后初次信访", "初次满意", "核查结论",
		"向您通报", "扫黑六霸", "典型案例", "弄虚作假",
	}
	for c, h := range headers26 {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		f.SetCellValue("信箱", cell, h)
	}

	all := append(append([]LetterExport{}, realData...), fakeData...)
	for i, e := range all {
		if i >= 250 {
			break
		}
		row := i + 2
		setVal(f, "信箱", 1, row, i+1)
		setVal(f, "信箱", 2, row, e.LetterNo)
		setVal(f, "信箱", 3, row, e.StatusName)
		setVal(f, "信箱", 4, row, trunc(e.ReceivedAt, 10))
		setVal(f, "信箱", 5, row, e.ChannelName)
		setVal(f, "信箱", 6, row, e.CitizenName)
		setVal(f, "信箱", 7, row, e.Phone)
		setVal(f, "信箱", 8, row, e.CatL1)
		setVal(f, "信箱", 9, row, first(e.CatL3, e.CatL2))
		setVal(f, "信箱", 10, row, trunc(e.Content, 80))
		setVal(f, "信箱", 11, row, e.CountyName)
		setVal(f, "信箱", 12, row, e.StationName)
		setVal(f, "信箱", 13, row, "否")
		setVal(f, "信箱", 14, row, "否")
		setVal(f, "信箱", 15, row, "已签收")
		setVal(f, "信箱", 16, row, "已签收")
		setVal(f, "信箱", 17, row, "未回访")
		setVal(f, "信箱", 18, row, "")
		setVal(f, "信箱", 19, row, "否")
		setVal(f, "信箱", 20, row, "")
		setVal(f, "信箱", 21, row, "")
		setVal(f, "信箱", 22, row, "")
		setVal(f, "信箱", 23, row, "")
		setVal(f, "信箱", 24, row, "")
		setVal(f, "信箱", 25, row, "")
		setVal(f, "信箱", 26, row, "")
	}

	for _, sname := range []string{"工单", "专席", "信访子系统"} {
		f.NewSheet(sname)
		for c, h := range headers26 {
			cell, _ := excelize.CoordinatesToCellName(c+1, 1)
			f.SetCellValue(sname, cell, h)
		}
		for i, e := range fakeData {
			if i >= 80 {
				break
			}
			row := i + 2
			setVal(f, sname, 1, row, i+1)
			setVal(f, sname, 2, row, e.LetterNo)
			setVal(f, sname, 3, row, e.StatusName)
			setVal(f, sname, 4, row, trunc(e.ReceivedAt, 10))
			setVal(f, sname, 5, row, e.ChannelName)
			setVal(f, sname, 6, row, e.CitizenName)
			setVal(f, sname, 7, row, e.Phone)
			setVal(f, sname, 8, row, e.CatL1)
			setVal(f, sname, 9, row, first(e.CatL3, e.CatL2))
			setVal(f, sname, 10, row, trunc(e.Content, 80))
			setVal(f, sname, 11, row, e.CountyName)
			setVal(f, sname, 12, row, e.StationName)
		}
	}

	f.NewSheet("日报")
	title := fmt.Sprintf("局长信箱日报重点信件反馈统计表（%s）", periodLabel)
	f.MergeCell("日报", "A1", "K1")
	setVal(f, "日报", 1, 1, title)
	dailyHeaders := []string{"序号", "上报日期", "批示日期", "批示内容", "重点信件基本情况", "核查结论", "处置进展", "反馈情况", "批示逾期", "分县局", "类别"}
	for c, h := range dailyHeaders {
		cell, _ := excelize.CoordinatesToCellName(c+1, 2)
		f.SetCellValue("日报", cell, h)
	}
	for i := 0; i < 15; i++ {
		row := i + 3
		setVal(f, "日报", 1, row, i+1)
		setVal(f, "日报", 2, row, time.Now().AddDate(0, 0, -i).Format("2006-01-02"))
	}

	f.NewSheet("日报数据统计")
	statsHeaders := []string{"日期", "局长信箱总数", "局长信箱有效件", "12345工单总数", "12345工单有效件",
		"12345自接总数", "12345自接有效件", "12389子系统", "12337政法委",
		"信访件总数", "信访件重复件", "局长收信总数", "局长收信有效件",
		"市委书记直通车", "群众来信", "总数合计"}
	for c, h := range statsHeaders {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		f.SetCellValue("日报数据统计", cell, h)
	}
	for i := 0; i < 25; i++ {
		row := i + 2
		setVal(f, "日报数据统计", 1, row, time.Now().AddDate(0, 0, -i).Format("01-02"))
		for c := 2; c <= 16; c++ {
			setVal(f, "日报数据统计", c, row, fmt.Sprintf("%d", (25-i)*3+c-2))
		}
	}

	path := filepath.Join(dir, fmt.Sprintf("%s数据汇总.xlsx", periodLabel))
	return path, f.SaveAs(path)
}

// ─── 2. 通报数图统计 ───

func genStatisticsChart(dir, yearMonth, monthName string, realData, fakeData []LetterExport) (string, error) {
	tmpl := "/Users/v_liheng02/Desktop/other/局长信箱原始资料/2026年3月通报数图统计.xlsx"
	if _, err := os.Stat(tmpl); os.IsNotExist(err) {
		return genEmptyExcel(dir, fmt.Sprintf("%s通报数图统计.xlsx", yearMonth))
	}
	src, _ := os.Open(tmpl)
	dstPath := filepath.Join(dir, fmt.Sprintf("%s通报数图统计.xlsx", yearMonth))
	dst, _ := os.Create(dstPath)
	io.Copy(dst, src)
	src.Close()
	dst.Close()

	f, err := excelize.OpenFile(dstPath)
	if err != nil {
		return dstPath, nil
	}
	defer f.Close()
	for _, sname := range f.GetSheetList() {
		rows, _ := f.GetRows(sname)
		for ri, row := range rows {
			for ci, cell := range row {
				if strings.Contains(cell, "3月") {
					col, _ := excelize.ColumnNumberToName(ci + 1)
					f.SetCellValue(sname, fmt.Sprintf("%s%d", col, ri+1), strings.ReplaceAll(cell, "3月", monthName))
				}
			}
		}
	}
	f.Save()
	return dstPath, nil
}

// ─── 3. PDF stub ───

func genPDFStub(dir, yearMonth, reportType string) (string, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(0, 10, fmt.Sprintf("%s %s", yearMonth, reportType), "", 1, "C", false, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 7, "This PDF is a placeholder. CJK text requires font preprocessing.", "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 7, "See the Excel files for detailed Chinese content.", "", 1, "L", false, 0, "")
	path := filepath.Join(dir, fmt.Sprintf("%s%s.pdf", yearMonth, reportType))
	return path, pdf.OutputFileAndClose(path)
}

// ─── 4. 分类标准复制 ───

func copyClassificationStandard(dir string) (string, error) {
	src := "/Users/v_liheng02/Desktop/other/局长信箱原始资料/20260310全平台数据分类标准.xlsx"
	dst := filepath.Join(dir, "20260310全平台数据分类标准.xlsx")
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return genEmptyExcel(dir, "20260310全平台数据分类标准.xlsx")
	}
	s, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	return dst, err
}

// ─── helpers ───

func setVal(f *excelize.File, sheet string, col, row int, val interface{}) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	f.SetCellValue(sheet, cell, val)
}

func trunc(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func first(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func genEmptyExcel(dir, name string) (string, error) {
	f := excelize.NewFile()
	setVal(f, "Sheet1", 1, 1, "No data")
	path := filepath.Join(dir, name)
	return path, f.SaveAs(path)
}

func ExportMonthlyReport(permLevel string, unitID *uint, period string) (string, error) {
	return GenerateMonthlyReportZip(permLevel, unitID, period)
}
