package service

import (
	"fmt"
	"sync"
	"time"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"

)

// ─── 请求级缓存：避免同一导出请求内重复查询 letters ───

var (
	exportLettersCache    []model.Letter
	exportLettersCacheKey string
	exportCacheMu         sync.Mutex
)

// FlushExportCache 清空导出缓存（每次导出完成后调用）
func FlushExportCache() {
	exportCacheMu.Lock()
	defer exportCacheMu.Unlock()
	exportLettersCache = nil
	exportLettersCacheKey = ""
}

// ExportGetLettersInRangeCached 带缓存的版本：同一时间范围内只查一次
func ExportGetLettersInRangeCached(startTime, endTime time.Time) ([]model.Letter, error) {
	key := fmt.Sprintf("%d-%d", startTime.Unix(), endTime.Unix())

	exportCacheMu.Lock()
	if key == exportLettersCacheKey && exportLettersCache != nil {
		cached := exportLettersCache
		exportCacheMu.Unlock()
		return cached, nil
	}
	exportCacheMu.Unlock()

	letters, err := ExportGetLettersInRange(startTime, endTime)
	if err != nil {
		return nil, err
	}

	exportCacheMu.Lock()
	exportLettersCache = letters
	exportLettersCacheKey = key
	exportCacheMu.Unlock()

	return letters, nil
}

// ─── 公共查询结构 ───

// LetterStatsRow 单行统计结果
type LetterStatsRow struct {
	Name  string
	Count int
}

// UnitCategoryStats 单位×类别交叉统计
type UnitCategoryStats struct {
	UnitName     string
	Complaint    int // 投诉
	Report       int // 举报
	Suggestion   int // 意见建议
	Consult      int // 咨询
	Appeal       int // 申诉
	Help         int // 求助
	Thank        int // 感谢
	DirectorMail int // 局长信箱
	Sub12389     int // 12389子系统
	VisitBJ      int // 进京到部赴省访
	WorkOrder    int // 12345工单
	PoliceDesk   int // 12345公安专席
	Total        int // 合计
}

// UnitChannelStats 单位×渠道交叉统计
type UnitChannelStats struct {
	UnitName     string
	DirectorMail int
	Sub12389     int
	VisitBJ      int
	WorkOrder    int
	PoliceDesk   int
}

// TeamStats 基层所队统计
type TeamStats struct {
	TeamName   string
	Complaint  int
	Report     int
	Suggestion int
	Consult    int
	Appeal     int
	Help       int
	Thank      int
	Total      int
}

// RepeatStats 重复件统计
type RepeatStats struct {
	TeamName     string
	RepeatCount  int
	CountyName   string
	TotalLetters int
}

// ComplaintDist 投诉分布（细类×单位）
type ComplaintDist struct {
	Index int
	Name  string
	// 接处警投诉
	NoPolice     int // 不出警出警慢
	NoCase       int // 有案不立
	Delay        int // 压案不查
	IllegalProc  int // 不按法定程序
	Shirk        int // 推诿扯皮
	OtherCompl   int // 其他投诉
	ComplTotal   int // 合计
	// 队伍纪律
	IllegalSeize int // 违规查扣冻
	EconDispute  int // 插手经济纠纷
	Arbitrary    int // 乱罚款乱收费
	FavorCase    int // 办人情案
	Cumbersom    int // 程序繁琐
	Bribe        int // 吃拿卡要
	Arrogance    int // 耍官威搞特权
	PoliceViol   int // 民警违纪违法
	OtherDiscip  int // 其他投诉
	AccidentSlow int // 事故处理慢不公正
}

// TeamComplaintStats 投诉所队分布
type TeamComplaintStats struct {
	TeamName    string
	NoPolice    int
	NoCase      int
	Delay       int
	IllegalProc int
	Shirk       int
	Total       int
}

// AppealDist 申诉分布
type AppealDist struct {
	Index            int
	Name             string
	ProcNotUnderstand int // 案件办理程序不理解
	PenaltyDisagree  int // 行政处罚决定有异议
	NotQualified     int // 不符合业务办理条件
	AccidentDisagree int // 事故认定结论不认同
	Total            int
}

// ─── 查询函数 ───

// ExportGetLettersInRange 获取时间范围内所有信件（预加载Category和Unit）
func ExportGetLettersInRange(startTime, endTime time.Time) ([]model.Letter, error) {
	var letters []model.Letter
	err := dao.DB.Preload("Category").Preload("CurrentUnitObj").
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Order("received_at DESC").
		Find(&letters).Error
	return letters, err
}

// ExportCountLettersInRange 统计时间范围内信件总数
func ExportCountLettersInRange(startTime, endTime time.Time) int64 {
	var count int64
	dao.DB.Model(&model.Letter{}).
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Count(&count)
	return count
}

// ExportGetCategoryStats 按分类统计
func ExportGetCategoryStats(startTime, endTime time.Time) []LetterStatsRow {
	var results []LetterStatsRow
	dao.DB.Model(&model.Letter{}).
		Select("COALESCE(categories.level1, '未分类') as name, count(*) as count").
		Joins("LEFT JOIN categories ON categories.id = letters.category_id").
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Group("categories.level1").
		Order("count DESC").
		Scan(&results)
	return results
}

// ExportGetChannelStats 按渠道统计
func ExportGetChannelStats(startTime, endTime time.Time) []LetterStatsRow {
	var results []LetterStatsRow
	rows, _ := dao.DB.Model(&model.Letter{}).
		Select("channel, count(*) as count").
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Group("channel").
		Order("count DESC").
		Rows()
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var ch int
			var cnt int
			rows.Scan(&ch, &cnt)
			name := model.ChannelToName[model.ChannelCode(ch)]
			if name == "" {
				name = "未知渠道"
			}
			results = append(results, LetterStatsRow{Name: name, Count: cnt})
		}
	}
	return results
}

// ExportGetStatusStats 按状态统计
func ExportGetStatusStats(startTime, endTime time.Time) []LetterStatsRow {
	var results []LetterStatsRow
	rows, _ := dao.DB.Model(&model.Letter{}).
		Select("current_status, count(*) as count").
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Group("current_status").
		Order("current_status").
		Rows()
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var st int
			var cnt int
			rows.Scan(&st, &cnt)
			name := model.StatusCodeToName[model.StatusCode(st)]
			if name == "" {
				name = "未知状态"
			}
			results = append(results, LetterStatsRow{Name: name, Count: cnt})
		}
	}
	return results
}

// ExportGetUnitLevel1Stats 按分县局（unit level1）统计
func ExportGetUnitLevel1Stats(startTime, endTime time.Time) []LetterStatsRow {
	var results []LetterStatsRow
	dao.DB.Model(&model.Letter{}).
		Select("COALESCE(units.level2, units.level1, '未知单位') as name, count(*) as count").
		Joins("LEFT JOIN units ON units.id = letters.current_unit_id").
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Group("units.level2, units.level1").
		Order("count DESC").
		Scan(&results)
	return results
}

// ExportGetUnitCategoryCross 单位×类别交叉统计（总览表用）
func ExportGetUnitCategoryCross(startTime, endTime time.Time) []UnitCategoryStats {
	letters, err := ExportGetLettersInRangeCached(startTime, endTime)
	if err != nil || len(letters) == 0 {
		return nil
	}

	// 按单位Level2+类别+渠道聚合
	type key struct {
		unit     string
		catL1    string
		chanCode model.ChannelCode
	}
	agg := make(map[key]int)
	unitSet := make(map[string]bool)
	totals := UnitCategoryStats{UnitName: "有效\n总量"}

	for _, l := range letters {
		unitName := ""
		if l.CurrentUnitObj != nil {
			if l.CurrentUnitObj.Level2 != "" {
				unitName = l.CurrentUnitObj.Level2
			} else if l.CurrentUnitObj.Level1 == "市局" {
				unitName = "市本级"
			} else {
				unitName = l.CurrentUnitObj.Level1
			}
		}
		if unitName == "" {
			unitName = "未知"
		}
		unitSet[unitName] = true

		catL1 := ""
		if l.Category != nil {
			catL1 = l.Category.Level1
		}
		k := key{unitName, catL1, l.Channel}
		agg[k]++
	}

	// 构建结果
	var result []UnitCategoryStats
	for un := range unitSet {
		row := UnitCategoryStats{UnitName: un}
		for k, v := range agg {
			if k.unit != un {
				continue
			}
			switch k.catL1 {
			case "投诉举报类":
				row.Complaint += v
			case "提供社会违法线索类":
				row.Report += v
			case "意见建议类":
				row.Suggestion += v
			case "咨询政策类":
				row.Consult += v
			case "申诉类":
				row.Appeal += v
			case "求助类":
				row.Help += v
			case "表扬肯定类":
				row.Thank += v
			}
			switch k.chanCode {
			case model.ChannelDirectorMail:
				row.DirectorMail += v
			case 8: // 12389子系统
				row.Sub12389 += v
			case 9: // 进京到部赴省访
				row.VisitBJ += v
			case 10: // 12345工单
				row.WorkOrder += v
			case 11: // 12345公安专席
				row.PoliceDesk += v
			}
		}
		row.Total = row.Complaint + row.Report + row.Suggestion + row.Consult + row.Appeal + row.Help + row.Thank
		result = append(result, row)
	}

	// 计算总量行
	for _, row := range result {
		totals.Complaint += row.Complaint
		totals.Report += row.Report
		totals.Suggestion += row.Suggestion
		totals.Consult += row.Consult
		totals.Appeal += row.Appeal
		totals.Help += row.Help
		totals.Thank += row.Thank
		totals.DirectorMail += row.DirectorMail
		totals.Sub12389 += row.Sub12389
		totals.VisitBJ += row.VisitBJ
		totals.WorkOrder += row.WorkOrder
		totals.PoliceDesk += row.PoliceDesk
		totals.Total += row.Total
	}
	result = append(result, totals)

	return result
}

// ExportGetTeamStats 基层所队统计
func ExportGetTeamStats(startTime, endTime time.Time) []TeamStats {
	letters, err := ExportGetLettersInRangeCached(startTime, endTime)
	if err != nil || len(letters) == 0 {
		return nil
	}

	type teamKey struct {
		unitName string // level2 / level3
	}
	teamAgg := make(map[string]*TeamStats)
	for _, l := range letters {
		name := ""
		if l.CurrentUnitObj != nil {
			if l.CurrentUnitObj.Level3 != "" {
				name = l.CurrentUnitObj.Level3
			} else if l.CurrentUnitObj.Level2 != "" {
				name = l.CurrentUnitObj.Level2
			} else {
				name = l.CurrentUnitObj.Level1
			}
		}
		if name == "" {
			continue
		}
		if _, ok := teamAgg[name]; !ok {
			teamAgg[name] = &TeamStats{TeamName: name}
		}
		ts := teamAgg[name]
		catL1 := ""
		if l.Category != nil {
			catL1 = l.Category.Level1
		}
		switch catL1 {
		case "投诉举报类":
			ts.Complaint++
		case "提供社会违法线索类":
			ts.Report++
		case "意见建议类":
			ts.Suggestion++
		case "咨询政策类":
			ts.Consult++
		case "申诉类":
			ts.Appeal++
		case "求助类":
			ts.Help++
		case "表扬肯定类":
			ts.Thank++
		}
		ts.Total++
	}

	var result []TeamStats
	for _, v := range teamAgg {
		result = append(result, *v)
	}
	return result
}

// ExportGetDailyStats 日报数据统计（每日各渠道统计）
type DailyStats struct {
	Date          string
	DirectorTotal int
	DirectorValid int
	WorkOrderTotal int
	WorkOrderValid int
	SelfTotal     int
	SelfValid     int
	Sub12389      int
	Sub12337      int
	PetitionTotal int
	PetitionRepeat int
	ChiefTotal    int
	ChiefValid    int
	MayorDirect   int
	CitizenMail   int
	GrandTotal    int
}

func ExportGetDailyStats(startTime, endTime time.Time) []DailyStats {
	var results []DailyStats
	// 按天分组统计
	type dayAgg struct {
		date string
		total int
		channelCounts map[int]int
	}
	dao.DB.Model(&model.Letter{}).
		Select("DATE(received_at) as date, channel, count(*) as cnt").
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Group("DATE(received_at), channel").
		Order("DATE(received_at) DESC").
		Scan(&results) // simplified, will refine

	// 更多细化查询…先简化处理
	return results
}

// ExportGetLetterByChannel 按渠道获取信件（用于分sheet导出）
func ExportGetLetterByChannel(startTime, endTime time.Time, channelCodes ...model.ChannelCode) ([]model.Letter, error) {
	var letters []model.Letter
	query := dao.DB.Preload("Category").Preload("CurrentUnitObj").
		Where("received_at >= ? AND received_at < ?", startTime, endTime)
	if len(channelCodes) > 0 {
		query = query.Where("channel IN ?", channelCodes)
	}
	err := query.Order("received_at DESC").Find(&letters).Error
	return letters, err
}

// ExportGetRepeatLetters 获取重复信件统计（按同一人+同一内容近似判断）
func ExportGetRepeatLetters(startTime, endTime time.Time) []RepeatStats {
	var results []RepeatStats
	// 简化：按phone+citizen_name分组计数>1的视为重复
	dao.DB.Model(&model.Letter{}).
		Select("COALESCE(citizen_name,'') as name, COALESCE(phone,'') as phone, count(*) as cnt").
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Group("citizen_name, phone").
		Having("count(*) > 1").
		Scan(&results)
	return results
}

// ExportCheckDataSufficiency 检查当前月份数据是否足够
func ExportCheckDataSufficiency(period string) (int, bool) {
	now := time.Now()
	var startTime time.Time
	switch period {
	case "month":
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
	endTime := startTime.AddDate(0, 1, 0)

	var count int64
	dao.DB.Model(&model.Letter{}).
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Count(&count)

	threshold := 200
	return int(count), int(count) >= threshold
}

// EnsureExportData 确保导出数据充足
func EnsureExportData(period string) {
	count, sufficient := ExportCheckDataSufficiency(period)
	if sufficient {
		return
	}
	// 数据不足，触发mock
	GenerateExportMockData(period, count)
}

// ─── LetterExport 构建 ───

// BuildLetterExport 构建导出用的 LetterExport 结构
func BuildLetterExport(l model.Letter) *LetterExport2 {
	le := &LetterExport2{
		LetterNo:    l.LetterNo,
		CitizenName: l.CitizenName,
		Phone:       l.Phone,
		ReceivedAt:  l.ReceivedAt.Format("2006-01-02 15:04:05"),
		ChannelName: getChannelName(l.Channel),
		StatusName:  getStatusName(l.CurrentStatus),
		StatusCode:  int(l.CurrentStatus),
		Content:     l.Content,
	}
	if l.Category != nil {
		le.CatL1 = l.Category.Level1
		le.CatL2 = l.Category.Level2
		le.CatL3 = l.Category.Level3
	}
	if l.CurrentUnitObj != nil {
		parts := []string{l.CurrentUnitObj.Level1, l.CurrentUnitObj.Level2, l.CurrentUnitObj.Level3}
		var nonEmpty []string
		for _, p := range parts {
			if p != "" {
				nonEmpty = append(nonEmpty, p)
			}
		}
		le.UnitName = joinStrings(nonEmpty, " / ")
		le.CountyName = l.CurrentUnitObj.Level1
		le.StationName = l.CurrentUnitObj.Level3
		if le.StationName == "" {
			le.StationName = l.CurrentUnitObj.Level2
		}
	}
	return le
}

// LetterExport2 导出数据结构（独立版本，不改原有 LetterExport）
type LetterExport2 struct {
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

func getChannelName(ch model.ChannelCode) string {
	// 先查标准枚举
	if name, ok := model.ChannelToName[ch]; ok {
		return name
	}
	// 扩展渠道（未在model中定义，但业务中使用）
	extMap := map[model.ChannelCode]string{
		8:  "12389子系统",
		9:  "进京到部赴省访",
		10: "12345工单",
		11: "12345公安专席",
	}
	if name, ok := extMap[ch]; ok {
		return name
	}
	return "未知渠道"
}

func getStatusName(st model.StatusCode) string {
	if name, ok := model.StatusCodeToName[st]; ok {
		return name
	}
	return "未知状态"
}

// ─── DB 辅助 ───

func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

// ─── 导出数据不足检查 ───

// ExportDataGaps 检查并返回缺失的数据项
type ExportDataGap struct {
	Field     string `json:"field"`
	Affected  string `json:"affected"`
	Status    string `json:"status"` // "empty" / "partial"
	Advice    string `json:"advice"`
}

func ExportCheckDataGaps(startTime, endTime time.Time) []ExportDataGap {
	var gaps []ExportDataGap

	// 1. 检查feedback回访数据
	var fbCount int64
	dao.DB.Model(&model.Feedback{}).Count(&fbCount)
	if fbCount == 0 {
		gaps = append(gaps, ExportDataGap{
			Field:    "回访记录",
			Affected: "数据汇总:C17-C18(回访满意/态度), 通报数图统计:「回访满意率」sheet",
			Status:   "empty",
			Advice:   "建议新建回访反馈表(feedbacks)，记录每封信的回访结果、满意度和态度",
		})
	}

	// 2. 检查渠道多样性
	var channelCount int64
	dao.DB.Model(&model.Letter{}).
		Select("COUNT(DISTINCT channel)").
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Count(&channelCount)
	if channelCount < 3 {
		gaps = append(gaps, ExportDataGap{
			Field:    "渠道多样性",
			Affected: "通报数图统计:「总览」渠道分布列",
			Status:   "partial",
			Advice:   "当前仅支持市民上报和局长信箱2种渠道。建议添加渠道枚举:12389子系统(8)、12345工单(10)、12345公安专席(11)、信访件等其他渠道",
		})
	}

	// 3. 检查分类多样性
	var catCount int64
	dao.DB.Model(&model.Letter{}).
		Joins("LEFT JOIN categories ON categories.id = letters.category_id").
		Where("received_at >= ? AND received_at < ?", startTime, endTime).
		Where("categories.level1 != ''").
		Select("COUNT(DISTINCT categories.level1)").
		Count(&catCount)
	if catCount < 3 {
		gaps = append(gaps, ExportDataGap{
			Field:    "分类多样性",
			Affected: "通报数图统计:各类别分布sheet(投诉/举报/意见/咨询/申诉/求助)",
			Status:   "partial",
			Advice:   "建议信件数据覆盖更多分类。categories表已有7个大类(投诉举报类/咨询政策类/意见建议类/申诉类/求助类/提供社会违法线索类/表扬肯定类)",
		})
	}

	// 4. 检查推广注册率数据
	gaps = append(gaps, ExportDataGap{
		Field:    "推广注册率",
		Affected: "通报数图统计:「推广注册率」sheet",
		Status:   "empty",
		Advice:   "推广注册率数据需要从外部推广系统获取。当前无相关数据表，建议新建推广数据表存储各分县局注册人数和户籍人口数",
	})

	// 5. 检查民警底数
	gaps = append(gaps, ExportDataGap{
		Field:    "民警底数",
		Affected: "通报数图统计:「各县民警底数」sheet",
		Status:   "empty",
		Advice:   "各分县局民警人数需要从人事系统获取。当前无相关数据表，建议新建民警统计表",
	})

	// 6. 签收超时数据
	gaps = append(gaps, ExportDataGap{
		Field:    "签收超时详情",
		Affected: "数据汇总:C15-C16, 通报数图统计:「签收办理」sheet",
		Status:   "partial",
		Advice:   "签收时间数据可以从letter_flows的JSON中提取。建议将签收/超时信息结构化存储到单独的表",
	})

	return gaps
}

// ─── 细类统计查询 ───

// CatDetailResult 细类统计结果：单位 → 细类 → 计数
type CatDetailResult struct {
	UnitName string
	CatL3    string
	Count    int
}

// ExportGetCategoryDetail 获取指定大类下各单位细类统计
// catL1: 大类名称如"投诉举报类"、"意见建议类"、"申诉类"等
func ExportGetCategoryDetail(startTime, endTime time.Time, catL1 string) []CatDetailResult {
	letters, _ := ExportGetLettersInRangeCached(startTime, endTime)

	type key struct {
		unit  string
		catL3 string
	}
	agg := make(map[key]int)

	for _, l := range letters {
		if l.Category == nil || l.Category.Level1 != catL1 {
			continue
		}
		unitName := ""
		if l.CurrentUnitObj != nil {
			if l.CurrentUnitObj.Level2 != "" {
				unitName = l.CurrentUnitObj.Level2
			} else {
				unitName = l.CurrentUnitObj.Level1
			}
		}
		if unitName == "" {
			unitName = "未知"
		}
		catL3 := l.Category.Level3
		if catL3 == "" {
			catL3 = l.Category.Level2
		}
		agg[key{unitName, catL3}]++
	}

	var results []CatDetailResult
	for k, v := range agg {
		results = append(results, CatDetailResult{
			UnitName: k.unit,
			CatL3:    k.catL3,
			Count:    v,
		})
	}
	return results
}

// ExportGetTeamCategoryDetail 获取指定大类下各所队细类统计
func ExportGetTeamCategoryDetail(startTime, endTime time.Time, catL1 string) []CatDetailResult {
	letters, _ := ExportGetLettersInRangeCached(startTime, endTime)

	type key struct {
		team  string
		catL3 string
	}
	agg := make(map[key]int)

	for _, l := range letters {
		if l.Category == nil || l.Category.Level1 != catL1 {
			continue
		}
		teamName := ""
		if l.CurrentUnitObj != nil {
			if l.CurrentUnitObj.Level3 != "" {
				teamName = l.CurrentUnitObj.Level3
			} else if l.CurrentUnitObj.Level2 != "" {
				teamName = l.CurrentUnitObj.Level2
			} else {
				teamName = l.CurrentUnitObj.Level1
			}
		}
		if teamName == "" {
			continue
		}
		catL3 := l.Category.Level3
		if catL3 == "" {
			catL3 = l.Category.Level2
		}
		agg[key{teamName, catL3}]++
	}

	var results []CatDetailResult
	for k, v := range agg {
		results = append(results, CatDetailResult{
			UnitName: k.team,
			CatL3:    k.catL3,
			Count:    v,
		})
	}
	return results
}

// ExportGetRepeatStats 重复件统计（按所队+分县局）
func ExportGetRepeatStats(startTime, endTime time.Time) []RepeatStats {
	letters, _ := ExportGetLettersInRangeCached(startTime, endTime)

	// Group by citizen name/phone to find repeats
	type repeatKey struct {
		name  string
		phone string
	}
	repeatCount := make(map[repeatKey]int)
	teamForRepeat := make(map[repeatKey]string)
	countyForRepeat := make(map[repeatKey]string)

	for _, l := range letters {
		rk := repeatKey{l.CitizenName, l.Phone}
		repeatCount[rk]++
		if l.CurrentUnitObj != nil {
			if l.CurrentUnitObj.Level3 != "" {
				teamForRepeat[rk] = l.CurrentUnitObj.Level3
			} else if l.CurrentUnitObj.Level2 != "" {
				teamForRepeat[rk] = l.CurrentUnitObj.Level2
			}
			countyForRepeat[rk] = l.CurrentUnitObj.Level2
		}
	}

	// Aggregate by team
	teamAgg := make(map[string]*RepeatStats)
	for rk, cnt := range repeatCount {
		if cnt <= 1 {
			continue
		}
		team := teamForRepeat[rk]
		if team == "" {
			team = "未知"
		}
		if _, ok := teamAgg[team]; !ok {
			teamAgg[team] = &RepeatStats{TeamName: team}
		}
		ts := teamAgg[team]
		ts.RepeatCount += cnt - 1 // 重复次数=出现次数-1
		ts.CountyName = countyForRepeat[rk]
		ts.TotalLetters += cnt
	}

	var results []RepeatStats
	for _, ts := range teamAgg {
		results = append(results, *ts)
	}
	return results
}

// ─── 导出数据检查结束 ───
