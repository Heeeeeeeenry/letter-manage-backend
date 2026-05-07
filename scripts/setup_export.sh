#!/bin/bash
# 导出功能部署脚本
# 安装 LibreOffice + 中文字体支持

set -e

echo "=== 导出功能环境部署 ==="

# 检测操作系统
if [[ "$(uname)" == "Darwin" ]]; then
    echo "macOS 环境"
    if ! command -v soffice &> /dev/null; then
        echo "正在安装 LibreOffice..."
        if command -v brew &> /dev/null; then
            brew install --cask libreoffice
        else
            echo "请先安装 Homebrew，或手动下载 LibreOffice:"
            echo "  https://www.libreoffice.org/download/"
            exit 1
        fi
    else
        echo "✓ LibreOffice 已安装"
    fi
    
elif [[ "$(uname)" == "Linux" ]]; then
    echo "Linux 环境"
    if ! command -v soffice &> /dev/null; then
        echo "正在安装 LibreOffice..."
        if command -v apt-get &> /dev/null; then
            sudo apt-get update
            sudo apt-get install -y libreoffice-core libreoffice-writer libreoffice-calc libreoffice-impress
        elif command -v yum &> /dev/null; then
            sudo yum install -y libreoffice-core libreoffice-writer libreoffice-calc libreoffice-impress
        else
            echo "请手动安装 LibreOffice"
            exit 1
        fi
    else
        echo "✓ LibreOffice 已安装"
    fi
    
    # 安装中文字体
    echo "正在安装中文字体..."
    if command -v apt-get &> /dev/null; then
        sudo apt-get install -y fonts-wqy-zenhei fonts-wqy-microhei || true
    elif command -v yum &> /dev/null; then
        sudo yum install -y cjkuni-uming-fonts || true
    fi
fi

# 验证
echo ""
echo "=== 验证 ==="
if command -v soffice &> /dev/null; then
    echo "✓ LibreOffice: $(soffice --version | head -1)"
else
    echo "✗ LibreOffice: 未安装 - PDF转换功能不可用"
    echo "  导出的ZIP中将不包含PDF文件"
fi

echo ""
echo "=== 完成 ==="
echo "导出功能已就绪。API路由: POST /api/letter/export_report"
echo ""
echo "使用方式:"
echo "  curl -X POST http://localhost:8080/api/letter/export_report \\"
echo "    -H \"Content-Type: application/json\" \\"
echo "    -H \"Cookie: session_key=...\" \\"
echo "    -d '{\"period\":\"month\"}'"
echo ""
echo "参数说明:"
echo "  period: month(默认) | day | week | year"
