package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// findPython 查找 Python3 可执行文件
func findPython() string {
	// 1. 先查 PATH
	if p, err := exec.LookPath("python3"); err == nil {
		return p
	}
	// 2. 常见 conda 路径
	for _, p := range []string{
		os.ExpandEnv("$HOME/work/conda/anaconda3/bin/python3"),
		os.ExpandEnv("$HOME/anaconda3/bin/python3"),
		os.ExpandEnv("$HOME/miniconda3/bin/python3"),
		"/opt/homebrew/bin/python3",
		"/usr/local/bin/python3",
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// findSoffice 查找 soffice 可执行文件
// macOS 上 LibreOffice 安装在 App Bundle 内，不在默认 PATH 中
func findSoffice() string {
	if p, err := exec.LookPath("soffice"); err == nil {
		return p
	}
	if runtime.GOOS == "darwin" {
		macPaths := []string{
			"/Applications/LibreOffice.app/Contents/MacOS/soffice",
			"/Applications/LibreOffice-still.app/Contents/MacOS/soffice",
		}
		for _, p := range macPaths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

// ─── PDF 报告生成 ───

// ReportPDFs 一对 PDF 报告文件路径
type ReportPDFs struct {
	Bulletin string // 通报.pdf
	Quality  string // 质态分析报告.pdf
}

// GenerateReportPDFs 用 Python reportlab 生成月报 PDF（格式匹配原始通报.pdf + 质态分析报告.pdf）
// startTime/endTime: 时间范围，periodLabel: 如 "5月"
func GenerateReportPDFs(summaryXLSX, chartXLSX string, startTime, endTime time.Time, periodLabel string, outputDir string) (*ReportPDFs, error) {
	python := findPython()
	if python == "" {
		return nil, fmt.Errorf("Python3 not found (install: pip3 install pymysql openpyxl reportlab)")
	}

	startStr := startTime.Format("2006-01-02")
	endStr := endTime.Format("2006-01-02")
	script := filepath.Join("scripts", "gen_report_pdfs.py")
	bulletinName := FormatBulletinFilename(periodLabel)
	qualityName := FormatAnalysisFilename(periodLabel)

	cmd := exec.Command(python, script, startStr, endStr, outputDir, periodLabel, bulletinName, qualityName)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Python report PDF generation failed: %w\nstderr may be in server log", err)
	}

	lines := splitLines(string(out))
	if len(lines) < 2 {
		return nil, fmt.Errorf("Python report PDF unexpected output: %s", string(out))
	}

	bulletin := strings.TrimSpace(lines[len(lines)-2])
	quality := strings.TrimSpace(lines[len(lines)-1])

	if _, err := os.Stat(bulletin); os.IsNotExist(err) {
		return nil, fmt.Errorf("通报.pdf not generated: %s", bulletin)
	}
	if _, err := os.Stat(quality); os.IsNotExist(err) {
		return nil, fmt.Errorf("质态分析报告.pdf not generated: %s", quality)
	}

	return &ReportPDFs{Bulletin: bulletin, Quality: quality}, nil
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// convertViaLibreOffice 用 LibreOffice 将 xlsx 转为 pdf
func convertViaLibreOffice(xlsxPath string) (string, error) {
	soffice := findSoffice()
	if soffice == "" {
		return "", fmt.Errorf("LibreOffice not installed")
	}

	outputDir := filepath.Dir(xlsxPath)
	cmd := exec.Command(soffice,
		"--headless",
		"--norestore",
		"-env:UserInstallation=file:///tmp/libreoffice_profile",
		"--convert-to", "pdf",
		"--outdir", outputDir,
		xlsxPath,
	)
	// 限制 LibreOffice 内存，避免 OOM
	cmd.Env = append(os.Environ(), "SAL_DISABLE_OPENCL=true")
	// LibreOffice first run may be slow (profile init), allow up to 60s
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("LibreOffice convert failed: %w", err)
	}

	baseName := filepath.Base(xlsxPath)
	ext := filepath.Ext(baseName)
	pdfName := baseName[:len(baseName)-len(ext)] + ".pdf"
	pdfPath := filepath.Join(outputDir, pdfName)

	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return "", fmt.Errorf("PDF not generated: %s", pdfPath)
	}
	return pdfPath, nil
}

// ─── 可用性检查 ───

// LibreOfficeAvailable 检查 LibreOffice 是否已安装
func LibreOfficeAvailable() bool {
	return findSoffice() != ""
}

// PDFConversionAvailable 检查 Python 是否可用（用于生成 PDF 报告）
func PDFConversionAvailable() bool {
	return findPython() != ""
}
