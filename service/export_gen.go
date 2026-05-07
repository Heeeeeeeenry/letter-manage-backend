package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type namedFile struct {
	Filename string
	Path     string
}

// GenerateFullExport 生成完整的月度导出ZIP包
func GenerateFullExport(permLevel string, unitID *uint, period string) (string, []ExportDataGap, error) {
	now := time.Now()
	startTime, endTime, periodLabel := calcPeriod(period, now)

	EnsureExportData(period)

	// 提取签收数据（从letter_flows JSON → letter_signoffs结构化表）
	InitSignoffExtraction()

	gaps := ExportCheckDataGaps(startTime, endTime)

	tmpDir, err := os.MkdirTemp("", "full_export_")
	if err != nil {
		return "", gaps, fmt.Errorf("create temp dir: %w", err)
	}

	var files []namedFile

	summaryPath, err := GenerateDataSummary(tmpDir, periodLabel, startTime, endTime)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", gaps, fmt.Errorf("generate data summary: %w", err)
	}
	files = append(files, namedFile{Filename: FormatDataSummaryFilename(periodLabel), Path: summaryPath})

	prevStart, prevEnd := getPrevMonthRange(startTime)
	chartPath, err := GenerateStatsChart(tmpDir, periodLabel, startTime, endTime, prevStart, prevEnd)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", gaps, fmt.Errorf("generate stats chart: %w", err)
	}
	files = append(files, namedFile{Filename: FormatStatsChartFilename(periodLabel), Path: chartPath})

	pdfConverted := false
	if LibreOfficeAvailable() {
		bulletinPath, err := ConvertXLSXToPDF(summaryPath)
		if err == nil {
			files = append(files, namedFile{Filename: FormatBulletinFilename(periodLabel), Path: bulletinPath})
			pdfConverted = true
		}
		analysisPath, err := ConvertXLSXToPDF(chartPath)
		if err == nil {
			files = append(files, namedFile{Filename: FormatAnalysisFilename(periodLabel), Path: analysisPath})
		}
	}

	if !pdfConverted {
		gaps = append(gaps, ExportDataGap{
			Field:    "PDF转换",
			Affected: "通报.pdf / 质态分析报告.pdf",
			Status:   "empty",
			Advice:   "需要安装 LibreOffice 才能将xlsx转换为PDF。安装命令: brew install --cask libreoffice",
		})
	}

	zipPath := filepath.Join(tmpDir, FormatZipFilename(periodLabel))
	if err := createZip(zipPath, files); err != nil {
		os.RemoveAll(tmpDir)
		return "", gaps, fmt.Errorf("create zip: %w", err)
	}

	return zipPath, gaps, nil
}

func calcPeriod(period string, now time.Time) (startTime, endTime time.Time, periodLabel string) {
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
	return
}

func getPrevMonthRange(currentStart time.Time) (*time.Time, *time.Time) {
	prevStart := currentStart.AddDate(0, -1, 0)
	prevEnd := currentStart
	return &prevStart, &prevEnd
}

func createZip(zipPath string, files []namedFile) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, f := range files {
		src, err := os.Open(f.Path)
		if err != nil {
			continue
		}
		w, err := zipWriter.Create(f.Filename)
		if err != nil {
			src.Close()
			continue
		}
		io.Copy(w, src)
		src.Close()
	}
	return nil
}

func ExportPermissionFilter(permLevel string, unitID *uint) {
	_ = permLevel
	_ = unitID
}
