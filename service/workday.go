package service

import (
	"time"
)

// 中国法定节假日（2026年）
// 注意：这里列举的是固定日期 + 调休信息，需要每年更新
// 数据来源：国务院办公厅关于2026年部分节假日安排的通知
var statutoryHolidays2026 = map[string]bool{
	// 元旦
	"2026-01-01": true, // 元旦
	// 春节
	"2026-02-17": true, // 除夕
	"2026-02-18": true, // 春节
	"2026-02-19": true, // 初二
	"2026-02-20": true, // 初三
	"2026-02-21": true, // 初四
	"2026-02-22": true, // 初五
	"2026-02-23": true, // 初六
	// 清明节
	"2026-04-04": true, // 清明节
	"2026-04-05": true, // 清明假期
	"2026-04-06": true, // 清明假期
	// 劳动节
	"2026-05-01": true, // 劳动节
	"2026-05-02": true,
	"2026-05-03": true,
	"2026-05-04": true,
	"2026-05-05": true,
	// 端午节
	"2026-06-19": true, // 端午节
	"2026-06-20": true,
	"2026-06-21": true,
	// 中秋节+国庆节
	"2026-10-01": true,
	"2026-10-02": true,
	"2026-10-03": true,
	"2026-10-04": true,
	"2026-10-05": true,
	"2026-10-06": true,
	"2026-10-07": true,
	"2026-10-08": true,
}

// 调休上班日（周六日需要上班补班）
var makeupWorkdays2026 = map[string]bool{
	// 元旦调休
	"2026-01-04": true, // 周日补班
	// 春节调休
	"2026-02-15": true, // 周日补班
	"2026-02-28": true, // 周六补班
	// 劳动节调休
	"2026-04-26": true, // 周日补班
	"2026-05-09": true, // 周六补班
	// 端午节调休
	"2026-06-14": true, // 周日补班
	// 中秋节+国庆节调休
	"2026-09-27": true, // 周日补班
	"2026-10-10": true, // 周六补班
}

// isHoliday 判断某天是否为法定节假日
func isHoliday(t time.Time) bool {
	dateStr := t.Format("2006-01-02")
	return statutoryHolidays2026[dateStr]
}

// isMakeupWorkday 判断某天是否为调休上班日
func isMakeupWorkday(t time.Time) bool {
	dateStr := t.Format("2006-01-02")
	return makeupWorkdays2026[dateStr]
}

// isWorkday 判断某天是否为工作日
// 工作日 = 周一至周五 + 调休上班日 - 法定节假日
func isWorkday(t time.Time) bool {
	weekday := t.Weekday()
	// 法定节假日不算工作日
	if isHoliday(t) {
		return false
	}
	// 调休上班日算工作日
	if isMakeupWorkday(t) {
		return true
	}
	// 周末不算工作日
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	// 周一到周五算工作日
	return true
}

// AddWorkdays 从 start 开始加上 n 个工作日，返回结果日期（不含起始日）
// 例如 start=周一, n=1 → 周二; start=周五, n=1 → 下周一
func AddWorkdays(start time.Time, n int) time.Time {
	current := start.AddDate(0, 0, 1) // 从第二天开始计算
	added := 0
	for added < n {
		if isWorkday(current) {
			added++
		}
		current = current.AddDate(0, 0, 1)
	}
	// 退回多加的一天
	return current.AddDate(0, 0, -1)
}

// GetWorkdayDeadline 计算从 start 开始加上 n 个工作日的截止日期时间点
// 返回截止日当天 23:59:59
func GetWorkdayDeadline(start time.Time, n int) time.Time {
	deadlineDate := AddWorkdays(start, n)
	return time.Date(
		deadlineDate.Year(), deadlineDate.Month(), deadlineDate.Day(),
		23, 59, 59, 0, deadlineDate.Location(),
	)
}
