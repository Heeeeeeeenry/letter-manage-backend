package service

import (
	"math/rand"
	"strings"
	"time"
)

// FakeDataSet generates realistic fake data for export testing
type FakeDataSet struct {
	rng *rand.Rand
}

func NewFakeDataSet() *FakeDataSet {
	return &FakeDataSet{rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// --- 中文字段生成器 ---

func (f *FakeDataSet) RandomLetterNo() string {
	prefixes := []string{"XJ", "SM", "GZ"}
	p := prefixes[f.rng.Intn(len(prefixes))]
	return p + time.Now().Add(-time.Duration(f.rng.Intn(60*24*30))*time.Hour).Format("20060102150405") + f.padNum(f.rng.Intn(9999))
}

func (f *FakeDataSet) RandomCitizenName() string {
	surnames := []string{"张", "李", "王", "刘", "陈", "杨", "赵", "黄", "周", "吴",
		"徐", "孙", "马", "朱", "胡", "郭", "何", "高", "林", "罗"}
	given := []string{"伟", "芳", "娜", "秀英", "敏", "静", "丽", "强", "磊", "军",
		"洋", "勇", "艳", "杰", "娟", "涛", "明", "超", "秀兰", "霞"}
	return surnames[f.rng.Intn(len(surnames))] + given[f.rng.Intn(len(given))]
}

func (f *FakeDataSet) RandomPhone() string {
	prefixes := []string{"138", "139", "150", "151", "152", "188", "187", "135", "136", "133"}
	return prefixes[f.rng.Intn(len(prefixes))] + f.padNum(f.rng.Intn(100000000))
}

func (f *FakeDataSet) RandomChannel() string {
	channels := []string{"局长信箱", "12345工单", "12345专席", "12389子系统", "信访件"}
	weights := []int{40, 25, 15, 10, 10}
	return f.weightedChoice(channels, weights)
}

func (f *FakeDataSet) RandomStatus() string {
	statuses := []string{"已办结", "办理中", "待审核", "已回复", "待签收"}
	weights := []int{45, 25, 15, 10, 5}
	return f.weightedChoice(statuses, weights)
}

func (f *FakeDataSet) RandomCategory() (level1, level2, level3 string) {
	categories := map[string]map[string][]string{
		"投诉举报类": {
			"执法办案投诉": {"不出警出警慢", "有案不立", "压案不查久拖不决", "不按法定程序和要求办理案件", "推诿扯皮办事拖拉"},
			"交通管理投诉": {"事故处理慢不公正", "交通秩序差", "交通设施损坏", "交通违法行为"},
			"其他投诉":   {"违规查扣冻", "插手经济纠纷", "乱罚款乱收费", "办人情案关系案金钱案"},
		},
		"咨询类": {
			"户籍业务": {"户籍迁移", "身份证办理", "户口本补办", "新生儿上户"},
			"交通业务": {"驾驶证换领", "车辆年检", "违章查询", "车辆过户"},
			"案件查询": {"案件进展", "处理结果", "办案流程"},
		},
		"意见建议类": {
			"案件办理":   {"案件办理效率", "案件反馈公开"},
			"交通管理":   {"车驾管业务方面", "道路设施方面", "公共秩序方面"},
			"文明执法":   {"公正文明执法", "提升队伍业务素质"},
			"窗口服务":   {"优化窗口服务机制模式", "户籍业务方面"},
		},
		"申诉类": {
			"案件申诉":   {"案件办理程序不理解", "行政处罚决定有异议"},
			"事故认定":   {"事故认定结论不认同"},
			"业务办理":   {"不符合业务办理条件"},
		},
		"求助类": {
			"寻人寻物": {"请求帮助寻人寻物"},
			"纠纷调解": {"请求调解民事纠纷"},
			"业务协调": {"请求上级协调业务办理指定管辖", "请求受理初查"},
		},
	}

	l1Keys := make([]string, 0, len(categories))
	for k := range categories {
		l1Keys = append(l1Keys, k)
	}
	level1 = l1Keys[f.rng.Intn(len(l1Keys))]

	l2Keys := make([]string, 0, len(categories[level1]))
	for k := range categories[level1] {
		l2Keys = append(l2Keys, k)
	}
	level2 = l2Keys[f.rng.Intn(len(l2Keys))]
	level3 = categories[level1][level2][f.rng.Intn(len(categories[level1][level2]))]
	return
}

func (f *FakeDataSet) RandomContent(catL3 string) string {
	templates := map[string][]string{
		"不出警出警慢": {
			"{市民}反映，{日期}在{地点}拨打110报警，但民警在{时间}后才到达现场，报警人称出警速度太慢，严重影响案件处理。",
			"投诉人称于{日期}{时间}左右在{地点}报警，值班民警出警延迟，请求处理。",
		},
		"有案不立": {
			"投诉人{市民}反映其于{日期}到{单位}报案，但至今未收到立案通知书，认为存在有案不立情况。",
		},
		"压案不查久拖不决": {
			"{市民}反映{单位}办理其案件拖延{时间}，至今未有实质性进展，请求上级部门督办。",
		},
		"户籍迁移": {
			"{市民}咨询户口从{地点}迁移到{地点}需要哪些材料，流程如何办理。",
		},
		"驾驶证换领": {
			"市民咨询驾驶证到期换领的具体流程和所需材料。",
		},
	}

	tmpls, ok := templates[catL3]
	if !ok {
		tmpls = []string{"{市民}反映在{地点}发生的{事项}，请求相关部门尽快处理回复。"}
	}
	tmpl := tmpls[f.rng.Intn(len(tmpls))]

	places := []string{"桃城区人民路", "高新区胜利路", "滨湖新区红旗大街", "桃城区中华大街", "景县县城", "故城县郑口镇", "安平县县城", "阜城县城区"}
	units := []string{"人民路派出所", "胜利路派出所", "桃城分局治安大队", "交管支队", "桃城分局刑警大队", "高新分局", "景县公安局"}
	dates := []string{"2026年3月初", "2026年4月中旬", "2026年5月1日", "2026年4月28日"}

	result := strings.ReplaceAll(tmpl, "{市民}", f.RandomCitizenName())
	result = strings.ReplaceAll(result, "{日期}", dates[f.rng.Intn(len(dates))])
	result = strings.ReplaceAll(result, "{时间}", "30分钟")
	result = strings.ReplaceAll(result, "{地点}", places[f.rng.Intn(len(places))])
	result = strings.ReplaceAll(result, "{单位}", units[f.rng.Intn(len(units))])
	result = strings.ReplaceAll(result, "{事项}", "相关事项")
	return result
}

func (f *FakeDataSet) RandomUnit() (county, station string) {
	counties := []string{"桃城分局", "高新分局", "滨湖分局", "冀州分局", "枣强县局", "武邑县局", "深州局", "武强县局", "饶阳县局", "安平县局", "景县局", "阜城县局", "故城县局"}
	stations := []string{"人民路派出所", "胜利路派出所", "新华路派出所", "河东派出所", "路北派出所",
		"治安大队", "刑警大队", "经侦大队", "交管大队", "禁毒大队", "食药大队", "网安大队"}
	county = counties[f.rng.Intn(len(counties))]
	station = stations[f.rng.Intn(len(stations))]
	return
}

func (f *FakeDataSet) RandomReceivedTime() time.Time {
	now := time.Now()
	daysAgo := f.rng.Intn(30)
	hoursAgo := f.rng.Intn(24)
	minsAgo := f.rng.Intn(60)
	return now.AddDate(0, 0, -daysAgo).Add(-time.Duration(hoursAgo)*time.Hour - time.Duration(minsAgo)*time.Minute)
}

func (f *FakeDataSet) RandomBool() string {
	if f.rng.Intn(2) == 0 {
		return "否"
	}
	return "是"
}

func (f *FakeDataSet) RandomSatisfaction() string {
	opts := []string{"满意", "基本满意", "不满意", "未回访"}
	weights := []int{40, 30, 15, 15}
	return f.weightedChoice(opts, weights)
}

func (f *FakeDataSet) RandomAttitude() string {
	opts := []string{"配合", "一般", "不配合", "激动"}
	weights := []int{45, 30, 15, 10}
	return f.weightedChoice(opts, weights)
}

// --- helpers ---

func (f *FakeDataSet) padNum(n int) string {
	s := ""
	switch {
	case n < 10:
		s = "000" 
	case n < 100:
		s = "00"
	case n < 1000:
		s = "0"
	}
	return s + string(rune('0'+n%10)) + string(rune('0'+(n/10)%10)) + string(rune('0'+(n/100)%10)) + string(rune('0'+(n/1000)%10))
}

func (f *FakeDataSet) weightedChoice(options []string, weights []int) string {
	total := 0
	for _, w := range weights {
		total += w
	}
	r := f.rng.Intn(total)
	cum := 0
	for i, w := range weights {
		cum += w
		if r < cum {
			return options[i]
		}
	}
	return options[len(options)-1]
}

func (f *FakeDataSet) RandSignStatus() string {
	opts := []string{"已签收", "未签收", "已签收", "已签收"}
	return opts[f.rng.Intn(len(opts))]
}

func (f *FakeDataSet) RandCheckResult() string {
	opts := []string{"查实", "查否", "部分属实", "属实", ""}
	weights := []int{15, 40, 20, 10, 15}
	return f.weightedChoice(opts, weights)
}

func (f *FakeDataSet) RandFirstSatisfaction() string {
	opts := []string{"满意", "基本满意", "不满意", ""}
	weights := []int{30, 25, 15, 30}
	return f.weightedChoice(opts, weights)
}

// GenerateFakeLetters generates n fake letter records
func (f *FakeDataSet) GenerateFakeLetters(n int) []LetterExport {
	letters := make([]LetterExport, n)
	for i := 0; i < n; i++ {
		county, station := f.RandomUnit()
		cat1, cat2, cat3 := f.RandomCategory()
		rt := f.RandomReceivedTime()

		letters[i] = LetterExport{
			LetterNo:    f.RandomLetterNo(),
			CitizenName: f.RandomCitizenName(),
			Phone:       f.RandomPhone(),
			CreatedAt:   rt.Format("2006-01-02 15:04:05"),
			ChannelName: f.RandomChannel(),
			StatusName:  f.RandomStatus(),
			StatusCode:  f.rng.Intn(9) + 1,
			Content:     f.RandomContent(cat3),
			CatL1:       cat1,
			CatL2:       cat2,
			CatL3:       cat3,
			UnitName:    county + " / " + station,
			CountyName:  county,
			StationName: station,
		}
	}
	return letters
}
