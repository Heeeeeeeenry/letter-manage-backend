package service

// ─── XLSX 常量 ───

// DataSummaryHeaders26 数据汇总26列表头（信箱/工单/专席/信访子系统通用）
var DataSummaryHeaders26 = []string{
	"序号", "信件编号", "信件状态", "来信时间", "来信渠道",
	"群众姓名", "手机号码", "信件类别", "信件细类", "简要诉求",
	"分县局", "主办单位", "重复信件", "退回信件",
	"县局签收", "所队签收", "回访满意", "回访时群众态度",
	"是否逾期", "办结后初次信访", "初次满意", "核查结论",
	"向您通报", "扫黑六霸", "典型案例", "弄虚作假",
}

// DailyReportHeaders 日报表头
var DailyReportHeaders = []string{
	"序号", "上报日期", "批示日期", "批示内容", "重点信件基本情况",
	"核查结论", "处置进展", "反馈情况", "批示逾期", "分县局", "类别",
}

// DailyStatsHeaders 日报数据统计表头
var DailyStatsHeaders = []string{
	"日期", "局长信箱总数", "局长信箱有效件", "12345工单总数", "12345工单有效件",
	"12345自接总数", "12345自接有效件",
	"12389子系统", "12337政法委",
	"信访件总数", "信访件重复件", "局长收信总数", "局长收信有效件", "市委书记直通车", "群众来信", "总数合计",
}

// DailyStatsSubHeaders 日报数据统计子表头
var DailyStatsSubHeaders = []string{
	"", "总数", "有效件", "总数", "有效件", "总数", "有效件",
	"", "", "总数", "重复件", "总数", "有效件", "", "", "",
}

// ─── 通报数图统计 Sheet 名称 ───

var StatsChartSheets = []string{
	"总览", "地域分布", "大警种", "基层所队", "重复件",
	"投诉分布", "投诉所队", "举报", "意见建议", "重点意见",
	"咨询变化", "求助", "申诉分布", "申诉所队", "Sheet7",
	"申诉详情", "推广注册率", "签收办理", "实质解决",
	"回访满意率", "各县民警底数", "非重复信件详情", "类别总透视",
}

// ─── 列宽常量 ───

var DefaultColWidths = map[string]float64{
	"A": 6, "B": 22, "C": 12, "D": 18, "E": 10,
	"F": 10, "G": 13, "H": 14, "I": 20, "J": 40,
	"K": 16, "L": 16, "M": 8, "N": 8,
}

// ─── 缺失数据占位符 ───

const (
	PlaceholderNoData     = "暂无数据"
	PlaceholderNoFeedback = "暂无回访记录"
	PlaceholderNoPromote  = "系统尚未接入推广数据"
	PlaceholderNoStaff    = "数据待补充"
)

// ─── XLSX 样式常量 ───

const (
	FontFamilyDefault = "微软雅黑"
	HeaderBgColor     = "#4472C4"
	BorderStyle       = 1
	BorderColor       = "000000"
)

// ─── 文件名格式 ───

func FormatDataSummaryFilename(periodLabel string) string {
	return periodLabel + "数据汇总.xlsx"
}

func FormatStatsChartFilename(periodLabel string) string {
	return periodLabel + "通报数图统计.xlsx"
}

func FormatBulletinFilename(periodLabel string) string {
	return periodLabel + "通报.pdf"
}

func FormatAnalysisFilename(periodLabel string) string {
	return periodLabel + "质态分析报告.pdf"
}

func FormatZipFilename(periodLabel string) string {
	return periodLabel + "导出.zip"
}
