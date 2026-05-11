package service

import (
	"fmt"
	"math/rand"
	"time"

	"letter-manage-backend/dao"
	"letter-manage-backend/model"
)

// GenerateExportMockData 生成测试用假数据
// 当数据不足时自动触发，确保报表有足够样本
func GenerateExportMockData(period string, existingCount int) {
	now := time.Now()
	targetTotal := 300

	need := targetTotal - existingCount
	if need <= 0 {
		return
	}

	// 查询数据库中的真实units和categories用于mock
	allUnits := loadAllUnits()
	allCats := loadAllCategories()

	if len(allUnits) == 0 || len(allCats) == 0 {
		return
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 生成的数据分布到当月及前2个月
	months := []time.Time{
		time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()),
	}
	if now.Month() > 1 {
		months = append(months, time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location()))
	}
	if now.Month() > 2 {
		months = append(months, time.Date(now.Year(), now.Month()-2, 1, 0, 0, 0, 0, now.Location()))
	}

	perMonth := need / len(months)
	if perMonth < 50 {
		perMonth = 50
	}

	for _, monthStart := range months {
		monthEnd := monthStart.AddDate(0, 1, 0)
		daysInMonth := int(monthEnd.Sub(monthStart).Hours() / 24)

		// 检查这个月已有多少数据
		var existing int64
		dao.DB.Model(&model.Letter{}).
			Where("letters.updated_at >= ? AND letters.updated_at < ?", monthStart, monthEnd).
			Count(&existing)

		toCreate := perMonth - int(existing)
		if toCreate <= 0 {
			continue
		}
		if toCreate > 300 {
			toCreate = 300
		}

		// 权重分配：各分县局
		unitWeights := getUnitWeights()

		for i := 0; i < toCreate; i++ {
			letter := buildMockLetter(rng, monthStart, daysInMonth, allUnits, allCats, unitWeights)
			if err := dao.DB.Create(letter).Error; err != nil {
				continue
			}
		}
	}
}

// ─── 内部数据结构 ───

type mockUnit struct {
	ID     uint
	Level1 string
	Level2 string
	Level3 string
}

type mockCategory struct {
	ID     uint
	Level1 string
	Level2 string
	Level3 string
}

func loadAllUnits() []mockUnit {
	var dbUnits []model.Unit
	dao.DB.Where("level1 IN ('分局','市局')").Find(&dbUnits)
	var result []mockUnit
	for _, u := range dbUnits {
		result = append(result, mockUnit{
			ID:     u.ID,
			Level1: u.Level1,
			Level2: u.Level2,
			Level3: u.Level3,
		})
	}
	return result
}

func loadAllCategories() []mockCategory {
	var dbCats []model.Category
	dao.DB.Find(&dbCats)
	var result []mockCategory
	for _, c := range dbCats {
		result = append(result, mockCategory{
			ID:     c.ID,
			Level1: c.Level1,
			Level2: c.Level2,
			Level3: c.Level3,
		})
	}
	return result
}

func getUnitWeights() map[string]int {
	return map[string]int{
		"桃城": 20, "高新": 8, "滨湖": 4, "冀州": 5, "枣强": 6,
		"武邑": 5, "深州": 8, "武强": 3, "饶阳": 3, "安平": 4,
		"故城": 6, "景县": 6, "阜城": 4, "交管": 10, "市局": 5,
		"其他": 3,
	}
}

func buildMockLetter(rng *rand.Rand, monthStart time.Time, daysInMonth int,
	units []mockUnit, cats []mockCategory, unitWeights map[string]int) *model.Letter {

	// 随机时间
	dayOffset := rng.Intn(daysInMonth)
	hour := rng.Intn(14) + 8 // 8:00-22:00
	minute := rng.Intn(60)
	receivedAt := monthStart.AddDate(0, 0, dayOffset).Add(
		time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)

	// 随机单位（按权重）
	unit := weightedPickUnit(rng, units, unitWeights)

	// 随机分类（按分布）
	cat := weightedPickCategory(rng, cats)

	// 随机渠道
	channel := weightedPickChannel(rng)

	// 随机状态（按正常流转分布）
	status := weightedPickStatus(rng)

	// 随机内容
	citizenName := randomName(rng)
	phone := randomPhone(rng)
	letterNo := fmt.Sprintf("XJ%s%04d", receivedAt.Format("20060102150405"), rng.Intn(9999))
	content := randomContent(rng, cat.Level3, citizenName)

	letter := &model.Letter{
		LetterNo:      letterNo,
		CitizenName:   citizenName,
		Phone:         phone,
		IDCard:        "",
		Channel:       channel,
		CategoryID:    &cat.ID,
		Content:       content,
		CurrentUnitID: &unit.ID,
		HandlerUserID: nil,
		HandlerUnitID: nil,
		CurrentStatus: status,
		DeadlineAt:    nil,
	}

	// 30% 概率设置 deadline
	if rng.Intn(100) < 30 {
		dl := receivedAt.AddDate(0, 0, 15)
		letter.DeadlineAt = &dl
	}

	return letter
}

// ─── 加权选择器 ───

func weightedPickUnit(rng *rand.Rand, units []mockUnit, weights map[string]int) mockUnit {
	total := 0
	for _, w := range weights {
		total += w
	}
	r := rng.Intn(total)
	cum := 0
	pickedLevel1 := ""
	for name, w := range weights {
		cum += w
		if r < cum {
			pickedLevel1 = name
			break
		}
	}
	if pickedLevel1 == "" {
		pickedLevel1 = "桃城"
	}

	// 从匹配的units中随机选一个
	var matched []mockUnit
	for _, u := range units {
		if u.Level1 == pickedLevel1 || (pickedLevel1 == "交管" && u.Level1 == "分局" && u.Level2 == "交管支队") {
			matched = append(matched, u)
		}
	}
	if len(matched) == 0 {
		matched = units
	}
	return matched[rng.Intn(len(matched))]
}

func weightedPickCategory(rng *rand.Rand, cats []mockCategory) mockCategory {
	// 按权重选level1
	type catLevel1 struct {
		name     string
		weight   int
	}
	levels := []catLevel1{
		{"投诉举报类", 30},
		{"咨询政策类", 15},
		{"意见建议类", 20},
		{"申诉类", 10},
		{"求助类", 15},
		{"提供社会违法线索类", 5},
		{"表扬肯定类", 5},
		{"其他类", 5},
	}
	total := 0
	for _, l := range levels {
		total += l.weight
	}
	r := rng.Intn(total)
	cum := 0
	pickedL1 := levels[0].name
	for _, l := range levels {
		cum += l.weight
		if r < cum {
			pickedL1 = l.name
			break
		}
	}

	var matched []mockCategory
	for _, c := range cats {
		if c.Level1 == pickedL1 {
			matched = append(matched, c)
		}
	}
	if len(matched) == 0 {
		return cats[rng.Intn(len(cats))]
	}
	return matched[rng.Intn(len(matched))]
}

func weightedPickChannel(rng *rand.Rand) model.ChannelCode {
	type chW struct {
		code   model.ChannelCode
		weight int
	}
	channels := []chW{
		{2, 40},  // 局长信箱
		{10, 20}, // 12345工单
		{11, 15}, // 12345公安专席
		{8, 10},  // 12389子系统
		{1, 10},  // 市民上报
		{7, 5},   // 其他
	}
	total := 0
	for _, c := range channels {
		total += c.weight
	}
	r := rng.Intn(total)
	cum := 0
	for _, c := range channels {
		cum += c.weight
		if r < cum {
			return c.code
		}
	}
	return 2
}

func weightedPickStatus(rng *rand.Rand) model.StatusCode {
	type stW struct {
		code   model.StatusCode
		weight int
	}
	statuses := []stW{
		{1, 5},   // 预处理
		{2, 8},   // 待区县局下发
		{3, 5},   // 已下发至分县局
		{4, 3},   // 市局越级下发
		{5, 8},   // 已下发至处理单位
		{6, 20},  // 处理中
		{7, 5},   // 待核查
		{8, 10},  // 待分县局审核
		{9, 5},   // 待市局审核
		{10, 25}, // 已办结
		{11, 3},  // 无效
		{12, 2},  // 已退回
		{13, 1},  // 已延期
	}
	total := 0
	for _, s := range statuses {
		total += s.weight
	}
	r := rng.Intn(total)
	cum := 0
	for _, s := range statuses {
		cum += s.weight
		if r < cum {
			return s.code
		}
	}
	return 10
}

// ─── 名称/电话/内容生成 ───

var surnames = []string{"张", "李", "王", "刘", "陈", "杨", "赵", "黄", "周", "吴",
	"徐", "孙", "马", "朱", "胡", "郭", "何", "高", "林", "罗"}
var givenNames = []string{"伟", "芳", "娜", "秀英", "敏", "静", "丽", "强", "磊", "军",
	"洋", "勇", "艳", "杰", "娟", "涛", "明", "超", "秀兰", "霞"}

func randomName(rng *rand.Rand) string {
	return surnames[rng.Intn(len(surnames))] + givenNames[rng.Intn(len(givenNames))]
}

var phonePrefixes = []string{"138", "139", "150", "151", "152", "188", "187", "135", "136", "133", "156", "176"}

func randomPhone(rng *rand.Rand) string {
	return phonePrefixes[rng.Intn(len(phonePrefixes))] + fmt.Sprintf("%08d", rng.Intn(100000000))
}

func randomContent(rng *rand.Rand, catL3, name string) string {
	templates := map[string][]string{
		"不出警、出警慢": {
			"%s反映，%s在%s拨打110报警，民警在较长时间后才到达现场，认为出警速度太慢。",
			"投诉%s报警后民警未及时出警，导致事态扩大。",
		},
		"有案不立": {
			"%s反映其于%s到%s报案，但至今未收到立案通知书，认为存在有案不立情况。",
			"投诉%s前往%s派出所报案，派出所不予受理。",
		},
		"压案不查，久拖不决": {
			"%s反映%s办理其案件拖延多日，至今未有实质性进展，要求上级部门督办。",
			"投诉%s的案件在%s久拖不决，请求加快办理。",
		},
		"案件办理效率": {
			"%s建议提高%s的案件办理效率，缩短办案周期。",
			"建议%s加快案件办理速度，提高工作效率。",
		},
		"户政业务咨询": {
			"%s咨询户口迁移到%s需要哪些材料和流程。",
			"%s咨询身份证办理的具体流程和所需时间。",
		},
		"交管业务咨询": {
			"%s咨询驾驶证到期换领的具体流程。",
			"%s咨询车辆年检的相关规定和地点。",
		},
		"请求帮助寻人/寻物": {
			"%s请求帮助寻找走失亲属，最后出现在%s附近。",
			"%s求助寻找遗失物品，请求调取%s附近监控。",
		},
		"整体工作表扬": {
			"群众%s对%s民警认真负责的工作态度表示衷心感谢。",
			"%s对%s高效处理其反映问题表示感谢。",
		},
	}
	tmpls, ok := templates[catL3]
	if !ok {
		tmpls = []string{"%s反映在%s发生的相关事项，请求相关部门尽快处理回复。"}
	}
	tmpl := tmpls[rng.Intn(len(tmpls))]

	places := []string{"桃城区人民路", "高新区胜利路", "滨湖新区红旗大街",
		"桃城区中华大街", "景县县城", "故城县郑口镇",
		"安平县县城", "阜城县城区", "枣强县城区",
		"武邑县城区", "深州市区", "冀州市区"}
	dates := []string{"近日", "本月", "上周", "前几天"}

	result := fmt.Sprintf(tmpl,
		name,
		dates[rng.Intn(len(dates))],
		places[rng.Intn(len(places))])
	return result
}
