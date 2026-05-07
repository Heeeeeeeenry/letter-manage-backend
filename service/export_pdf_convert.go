package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ConvertXLSXToPDF 使用 LibreOffice 将 xlsx 转换为 pdf
// 返回生成的文件路径
func ConvertXLSXToPDF(xlsxPath string) (string, error) {
	// 检查 LibreOffice 是否可用
	if _, err := exec.LookPath("soffice"); err != nil {
		return "", fmt.Errorf("LibreOffice not installed: %w", err)
	}

	// 输出目录与原文件同一目录
	outputDir := filepath.Dir(xlsxPath)

	// 执行转换
	cmd := exec.Command("soffice",
		"--headless",
		"--convert-to", "pdf",
		"--outdir", outputDir,
		xlsxPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("LibreOffice convert failed: %w", err)
	}

	// 生成的文件名：原文件名（不含扩展名）+ .pdf
	baseName := filepath.Base(xlsxPath)
	ext := filepath.Ext(baseName)
	pdfName := baseName[:len(baseName)-len(ext)] + ".pdf"
	pdfPath := filepath.Join(outputDir, pdfName)

	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		return "", fmt.Errorf("PDF not generated: %s", pdfPath)
	}

	return pdfPath, nil
}

// LibreOfficeAvailable 检查 LibreOffice 是否已安装
func LibreOfficeAvailable() bool {
	_, err := exec.LookPath("soffice")
	return err == nil
}
