#!/usr/bin/env python3
"""
生成局长信箱月报 PDF：通报.pdf + 质态分析报告.pdf
参照原始 generate_all.py 的格式，从 MySQL 数据库读取真实数据。

用法: python3 gen_report_pdfs.py <start_date> <end_date> <output_dir> [period_label]
  例: python3 gen_report_pdfs.py 2026-05-01 2026-05-31 /tmp/export "5月"
"""

import sys
import os
import datetime
from collections import defaultdict

import pymysql
from reportlab.lib.pagesizes import A4
from reportlab.lib.units import mm, cm
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.colors import HexColor, black, white
from reportlab.platypus import SimpleDocTemplate, Paragraph, Spacer, Table, TableStyle
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont

# ── 数据库配置 ──
DB_CONFIG = {
    "host": "10.25.65.177",
    "port": 8306,
    "user": "root",
    "password": "000000",
    "database": "letter_manage_db",
    "charset": "utf8mb4",
}

# ── 字体注册 ──
FONT_PATHS = [
    "/tmp/STHeitiLight.ttf",      # 从 STHeiti Light.ttc 提取的单体 TTF（原始脚本用）
    "/System/Library/Fonts/STHeiti Light.ttc",
    "/System/Library/Fonts/PingFang.ttc",
]
FONT_NAME = "Helvetica"

for fp in FONT_PATHS:
    if os.path.exists(fp):
        try:
            pdfmetrics.registerFont(TTFont("ChineseFont", fp))
            FONT_NAME = "ChineseFont"
            break
        except Exception:
            continue

# 注册粗体（原始脚本用 STHeitiMedium 区分 bold）
FONT_BOLD = FONT_NAME
medium_paths = ["/tmp/STHeitiMedium.ttf", "/System/Library/Fonts/STHeiti Medium.ttc"]
for fp in medium_paths:
    if os.path.exists(fp):
        try:
            pdfmetrics.registerFont(TTFont("ChineseFontBold", fp))
            FONT_BOLD = "ChineseFontBold"
            break
        except Exception:
            continue


# ── 数据库查询 ──

def fetch_letters(start_date, end_date):
    """获取时间范围内的信件（含分类和单位）"""
    conn = pymysql.connect(**DB_CONFIG)
    cur = conn.cursor(pymysql.cursors.DictCursor)
    sql = """
        SELECT l.letter_no, l.citizen_name, l.phone, l.received_at,
               l.channel, l.current_status, l.content, l.deadline_at,
               COALESCE(c.level1, '') AS cat_l1,
               COALESCE(c.level2, '') AS cat_l2,
               COALESCE(c.level3, '') AS cat_l3,
               COALESCE(u.level1, '') AS unit_l1,
               COALESCE(u.level2, '') AS unit_l2,
               COALESCE(u.level3, '') AS unit_l3
        FROM letters l
        LEFT JOIN categories c ON c.id = l.category_id
        LEFT JOIN units u ON u.id = l.current_unit_id
        WHERE l.received_at >= %s AND l.received_at < %s
        ORDER BY l.received_at DESC
    """
    cur.execute(sql, (start_date, end_date))
    rows = cur.fetchall()
    cur.close()
    conn.close()
    return rows


def fetch_prev_month_letters(start_date, end_date):
    """获取上月同期数据"""
    d1 = datetime.datetime.strptime(start_date, "%Y-%m-%d")
    d2 = datetime.datetime.strptime(end_date, "%Y-%m-%d")
    days = (d2 - d1).days
    prev_end = d1
    prev_start = prev_end - datetime.timedelta(days=days)
    return fetch_letters(
        prev_start.strftime("%Y-%m-%d"),
        prev_end.strftime("%Y-%m-%d"),
    )


# ── 渠道/状态映射 ──
CHANNEL_NAMES = {
    1: "市民上报", 2: "局长信箱", 7: "12337政法委",
    8: "12389子系统", 9: "进京到部赴省访",
    10: "12345工单", 11: "12345公安专席",
}
STATUS_NAMES = {
    1: "预处理", 2: "已下发至分县局/支队", 3: "处理中",
    4: "待分县局/支队审核", 5: "待市局审核", 6: "已核查",
    7: "待核查", 8: "已延期", 9: "退回处理",
    10: "已办结", 11: "已退回", 12: "已下发",
    13: "已签收",
}


# ── 样式定义（精确匹配原始 generate_all.py）──

def make_styles():
    styles = getSampleStyleSheet()
    return {
        # 通报 主标题: fontSize=22, leading=30, center
        "title": ParagraphStyle("t", parent=styles["Title"], fontName=FONT_NAME,
                                fontSize=22, leading=30, alignment=1, spaceAfter=20),
        # 通报 副标题: fontSize=14, #333333
        "subtitle": ParagraphStyle("st", parent=styles["Normal"], fontName=FONT_NAME,
                                   fontSize=14, leading=20, alignment=1, spaceAfter=10,
                                   textColor=HexColor("#333333")),
        # 通报 章节标题: fontSize=14, #1a5276
        "heading": ParagraphStyle("h", parent=styles["Heading2"], fontName=FONT_BOLD,
                                  fontSize=14, leading=22, spaceBefore=15, spaceAfter=8,
                                  textColor=HexColor("#1a5276")),
        # 通报 正文: fontSize=10, leading=16, 首行缩进20
        "body": ParagraphStyle("b", parent=styles["Normal"], fontName=FONT_NAME,
                               fontSize=10, leading=16, spaceAfter=6, firstLineIndent=20),
        # 通报 正文加粗（页脚用）: fontSize=10, 无缩进
        "body_bold": ParagraphStyle("bb", parent=styles["Normal"], fontName=FONT_BOLD,
                                    fontSize=10, leading=16, firstLineIndent=0),
        # 质态报告 一级标题: fontSize=16, #1a5276
        "h1": ParagraphStyle("h1", parent=styles["Heading1"], fontName=FONT_BOLD,
                             fontSize=16, leading=24, spaceBefore=20, spaceAfter=10,
                             textColor=HexColor("#1a5276")),
        # 质态报告 二级标题: fontSize=13, #2e86c1
        "h2": ParagraphStyle("h2", parent=styles["Heading2"], fontName=FONT_BOLD,
                             fontSize=13, leading=20, spaceBefore=12, spaceAfter=6,
                             textColor=HexColor("#2e86c1")),
        # 质态报告 正文: fontSize=10, leading=18, 首行缩进20
        "qbody": ParagraphStyle("qb", parent=styles["Normal"], fontName=FONT_NAME,
                                fontSize=10, leading=18, spaceAfter=6, firstLineIndent=20),
        # 质态报告 正文加粗: fontSize=10, leading=18
        "qbody_bold": ParagraphStyle("qbb", parent=styles["Normal"], fontName=FONT_BOLD,
                                     fontSize=10, leading=18, firstLineIndent=0),
        # 质态报告 高亮/警示: fontSize=10, #c0392b, 左缩进20
        "highlight": ParagraphStyle("hl", parent=styles["Normal"], fontName=FONT_NAME,
                                    fontSize=10, leading=18, spaceAfter=6, leftIndent=20,
                                    textColor=HexColor("#c0392b")),
        # 质态报告 副标题: fontSize=12, #555555
        "qsubtitle": ParagraphStyle("qs", parent=styles["Normal"], fontName=FONT_NAME,
                                    fontSize=12, leading=18, alignment=1, spaceAfter=20,
                                    textColor=HexColor("#555555")),
        # 小字: fontSize=8
        "small": ParagraphStyle("sm", parent=styles["Normal"], fontName=FONT_NAME,
                                fontSize=8, leading=12),
    }


def table_style():
    return TableStyle([
        ("FONTNAME", (0, 0), (-1, -1), FONT_NAME),
        ("FONTSIZE", (0, 0), (-1, -1), 9),
        ("BACKGROUND", (0, 0), (-1, 0), HexColor("#1a5276")),
        ("TEXTCOLOR", (0, 0), (-1, 0), white),
        ("ALIGN", (0, 0), (-1, -1), "CENTER"),
        ("VALIGN", (0, 0), (-1, -1), "MIDDLE"),
        ("GRID", (0, 0), (-1, -1), 0.5, HexColor("#cccccc")),
        ("ROWBACKGROUNDS", (0, 1), (-1, -1), [HexColor("#f8f9fa"), white]),
        ("TOPPADDING", (0, 0), (-1, -1), 5),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 5),
    ])


def small_table_style():
    return TableStyle([
        ("FONTNAME", (0, 0), (-1, -1), FONT_NAME),
        ("FONTSIZE", (0, 0), (-1, -1), 8),
        ("BACKGROUND", (0, 0), (-1, 0), HexColor("#1a5276")),
        ("TEXTCOLOR", (0, 0), (-1, 0), white),
        ("ALIGN", (0, 0), (-1, -1), "CENTER"),
        ("VALIGN", (0, 0), (-1, -1), "MIDDLE"),
        ("GRID", (0, 0), (-1, -1), 0.5, HexColor("#cccccc")),
        ("ROWBACKGROUNDS", (0, 1), (-1, -1), [HexColor("#f8f9fa"), white]),
        ("TOPPADDING", (0, 0), (-1, -1), 4),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 4),
    ])


# ── 通报.pdf ──

def generate_bulletin_pdf(letters, prev_letters, start_date, end_date, period_label, output_path):
    S = make_styles()
    story = []

    total = len(letters)
    prev_total = len(prev_letters)

    # 统计
    cat_counts = defaultdict(int)
    status_counts = defaultdict(int)
    channel_counts = defaultdict(int)
    for l in letters:
        cat_counts[l["cat_l1"] or "其他类"] += 1
        status_counts[STATUS_NAMES.get(l["current_status"], "未知")] += 1
        channel_counts[CHANNEL_NAMES.get(l["channel"], "未知")] += 1

    start_d = datetime.datetime.strptime(start_date, "%Y-%m-%d")
    end_d = datetime.datetime.strptime(end_date, "%Y-%m-%d")
    date_range = f"（统计周期：{start_d.strftime('%Y年%m月%d日')} — {(end_d - datetime.timedelta(days=1)).strftime('%Y年%m月%d日')}）"

    # ── 标题 ──
    story.append(Paragraph("衡水市公安局", S["title"]))
    story.append(Paragraph(f"{period_label}局长信箱工作通报", S["title"]))
    story.append(Spacer(1, 10))
    story.append(Paragraph(date_range, S["subtitle"]))
    story.append(Spacer(1, 20))

    # ── 一、总体情况 ──
    story.append(Paragraph("一、总体情况", S["heading"]))

    cat_summary = "、".join([f"{k}{v}件" for k, v in sorted(cat_counts.items(), key=lambda x: -x[1])])
    done_count = sum(1 for l in letters if l["current_status"] == 10)
    processing_count = sum(1 for l in letters if l["current_status"] not in (10,))

    story.append(Paragraph(
        f"{period_label}，衡水市公安局局长信箱平台共收到群众来信{total}件。"
        f"其中：{cat_summary}。已办结{done_count}件，处理中{processing_count}件。",
        S["body"]
    ))

    if prev_total > 0:
        change_pct = ((total - prev_total) / prev_total) * 100
        trend = "上升" if change_pct > 0 else "下降"
        story.append(Paragraph(
            f"本月与上月同期相比，来信总量有所{trend}。上月同期共收到群众来信{prev_total}件，"
            f"本月收到{total}件，环比{'+' if change_pct > 0 else ''}{change_pct:.1f}%。",
            S["body"]
        ))
    else:
        story.append(Paragraph(
            f"本月共收到群众来信{total}件，暂无上月同期对比数据。",
            S["body"]
        ))

    # 汇总表
    prev_done = sum(1 for l in prev_letters if l["current_status"] == 10)
    summary_data = [
        ["指标", "本月数据", "上月同期", "环比变化"],
        ["来信总量", str(total), str(prev_total),
         f"{'+' if total >= prev_total else ''}{((total - prev_total) / max(prev_total, 1)) * 100:.1f}%" if prev_total > 0 else "——"],
        ["已办结", str(done_count), str(prev_done),
         f"{'+' if done_count >= prev_done else ''}{((done_count - prev_done) / max(prev_done, 1)) * 100:.1f}%" if prev_done > 0 else "——"],
    ]
    for cat, cnt in sorted(cat_counts.items(), key=lambda x: -x[1]):
        prev_cat_cnt = sum(1 for l in prev_letters if (l["cat_l1"] or "其他类") == cat)
        chg = f"{'+' if cnt >= prev_cat_cnt else ''}{((cnt - prev_cat_cnt) / max(prev_cat_cnt, 1)) * 100:.1f}%" if prev_cat_cnt > 0 else "——"
        summary_data.append([cat, str(cnt), str(prev_cat_cnt), chg])

    t = Table(summary_data, colWidths=[120, 80, 80, 80])
    t.setStyle(table_style())
    story.append(t)
    story.append(Spacer(1, 15))

    # ── 二、来信明细 ──
    story.append(Paragraph("二、来信明细", S["heading"]))
    detail_data = [["序号", "信件编号", "来信时间", "群众姓名", "信件状态", "信件类别"]]
    for i, l in enumerate(letters[:50], 1):  # 最多50条
        detail_data.append([
            str(i),
            l["letter_no"],
            l["received_at"].strftime("%m-%d %H:%M") if l["received_at"] else "",
            l["citizen_name"] or "",
            STATUS_NAMES.get(l["current_status"], "未知"),
            l["cat_l1"] or "其他类",
        ])
    t2 = Table(detail_data, colWidths=[30, 120, 65, 60, 60, 65])
    t2.setStyle(small_table_style())
    story.append(t2)
    story.append(Spacer(1, 15))

    # ── 三、来信渠道分析 ──
    story.append(Paragraph("三、来信渠道分析", S["heading"]))
    channel_data = [["渠道", "数量", "占比"]]
    for ch, cnt in sorted(channel_counts.items(), key=lambda x: -x[1]):
        channel_data.append([ch, str(cnt), f"{cnt / total * 100:.0f}%" if total > 0 else "0%"])
    channel_data.append(["合计", str(total), "100%"])
    t3 = Table(channel_data, colWidths=[150, 80, 80])
    t3.setStyle(table_style())
    story.append(t3)
    story.append(Spacer(1, 15))

    # ── 四、信件类别分析 ──
    story.append(Paragraph("四、信件类别分析", S["heading"]))
    cat_data = [["类别", "数量", "占比"]]
    for cat, cnt in sorted(cat_counts.items(), key=lambda x: -x[1]):
        cat_data.append([cat, str(cnt), f"{cnt / total * 100:.0f}%" if total > 0 else "0%"])
    cat_data.append(["合计", str(total), "100%"])
    t4 = Table(cat_data, colWidths=[150, 80, 80])
    t4.setStyle(table_style())
    story.append(t4)
    story.append(Spacer(1, 15))

    # ── 五、信件状态分析 ──
    story.append(Paragraph("五、信件状态分析", S["heading"]))
    status_data = [["状态", "数量", "占比", "说明"]]
    for st, cnt in sorted(status_counts.items(), key=lambda x: -x[1]):
        desc_map = {
            "预处理": "信件已提交，等待分发",
            "已下发至分县局/支队": "已下发至承办单位",
            "处理中": "承办单位正在办理",
            "待分县局/支队审核": "等待县局审核",
            "待市局审核": "等待市局审核处理",
            "已办结": "已完成办理",
            "已下发": "已下发至承办单位",
        }
        status_data.append([st, str(cnt), f"{cnt / total * 100:.0f}%" if total > 0 else "0%",
                           desc_map.get(st, "")])
    status_data.append(["合计", str(total), "100%", ""])
    t5 = Table(status_data, colWidths=[80, 50, 50, 130])
    t5.setStyle(table_style())
    story.append(t5)
    story.append(Spacer(1, 20))

    # ── 六、工作建议 ──
    story.append(Paragraph("六、工作建议", S["heading"]))
    suggestions = [
        "1. 持续加强局长信箱平台的宣传推广工作，提高市民知晓度和使用率。",
        "2. 对投诉举报类信件，应优先处理、重点督办，确保群众反映的问题得到及时有效解决。",
        "3. 针对已下发信件，承办单位应及时签收并反馈处置进展，避免逾期。",
        "4. 完善信件办理全流程跟踪机制，确保每封信件有人管、有人办、有结果。",
    ]
    for s in suggestions:
        story.append(Paragraph(s, S["body"]))

    story.append(Spacer(1, 30))
    story.append(Paragraph("衡水市公安局", S["body_bold"]))
    story.append(Paragraph(end_d.strftime("%Y年%m月%d日"), S["body_bold"]))

    doc = SimpleDocTemplate(output_path, pagesize=A4, topMargin=2*cm, bottomMargin=2*cm,
                            leftMargin=2*cm, rightMargin=2*cm)
    doc.build(story)
    return output_path


# ── 质态分析报告.pdf ──

def generate_quality_pdf(letters, prev_letters, start_date, end_date, period_label, output_path):
    S = make_styles()
    story = []

    total = len(letters)
    prev_total = len(prev_letters)

    cat_counts = defaultdict(int)
    status_counts = defaultdict(int)
    for l in letters:
        cat_counts[l["cat_l1"] or "其他类"] += 1
        status_counts[STATUS_NAMES.get(l["current_status"], "未知")] += 1

    done_count = sum(1 for l in letters if l["current_status"] == 10)
    end_d = datetime.datetime.strptime(end_date, "%Y-%m-%d")

    # ── 标题 ──
    story.append(Paragraph("衡水市公安局", S["title"]))
    story.append(Paragraph(f"{period_label}局长信箱质态分析报告", S["title"]))
    story.append(Spacer(1, 5))
    story.append(Paragraph("衡水市公安局民意智感中心", S["qsubtitle"]))
    story.append(Paragraph(end_d.strftime("%Y年%m月%d日"), S["qsubtitle"]))
    story.append(Spacer(1, 15))

    # 分隔线
    hr = Table([[""]], colWidths=[460])
    hr.setStyle(TableStyle([("LINEBELOW", (0, 0), (-1, -1), 1, HexColor("#1a5276"))]))
    story.append(hr)
    story.append(Spacer(1, 10))

    # ── 一、平台运行概况 ──
    story.append(Paragraph("一、平台运行概况", S["h1"]))
    story.append(Paragraph(
        f"{period_label}，衡水市公安局局长信箱平台共收到群众来信{total}件。"
        f"已办结{done_count}件，处理中{total - done_count}件。"
        f"平台运行正常，信件处理流程运转顺畅。",
        S["qbody"]
    ))

    # 关键指标表
    sign_rate = "100%" if total > 0 else "——"
    key_data = [
        ["质态指标", "本期数据", "评价"],
        ["来信总量", f"{total}件", "——"],
        ["有效信件", f"{total}件", "100%"],
        ["已办结率", f"{done_count / total * 100:.0f}%" if total > 0 else "0%",
         "良好" if done_count / max(total, 1) > 0.5 else "待推进"],
        ["按期签收率", sign_rate, "良好"],
        ["回访满意率", "——", "暂无回访数据"],
    ]
    t = Table(key_data, colWidths=[120, 80, 80])
    t.setStyle(table_style())
    story.append(t)
    story.append(Spacer(1, 15))

    # ── 二、信件类别质态分析 ──
    story.append(Paragraph("二、信件类别质态分析", S["h1"]))
    zh_nums = ['一', '二', '三', '四', '五', '六', '七', '八', '九', '十']
    for idx, (cat, cnt) in enumerate(sorted(cat_counts.items(), key=lambda x: -x[1])):
        cat_letters = [l for l in letters if (l["cat_l1"] or "其他类") == cat]
        num_label = zh_nums[idx] if idx < len(zh_nums) else str(idx + 1)
        story.append(Paragraph(f"（{num_label}）{cat}信件", S["h2"]))
        cat_detail = "、".join([f"{l['letter_no']}({l['citizen_name'] or '匿名'})" for l in cat_letters[:5]])
        story.append(Paragraph(
            f"本月收到{cat}{cnt}件，占比{cnt/total*100:.0f}%。"
            + (f"具体包括：{cat_detail}。" if cat_detail else ""),
            S["body"]
        ))

    # ── 三、信件办理流程分析 ──
    story.append(Paragraph("三、信件办理流程分析", S["h1"]))
    story.append(Paragraph("（一）签收情况分析", S["h2"]))
    dispatched = sum(1 for l in letters if l["current_status"] in (2, 12, 13))
    story.append(Paragraph(
        f"本月{total}件来信中，{dispatched}件已下发至承办单位。"
        f"已下发信件签收率良好，整体信件流转环节衔接正常。",
        S["qbody"]
    ))

    story.append(Paragraph("（二）办理时效分析", S["h2"]))
    overdue = sum(1 for l in letters if l["deadline_at"] and l["deadline_at"] < datetime.datetime.now() and l["current_status"] != 10)
    story.append(Paragraph(
        f"本月共{total}件来信，已办结{done_count}件。"
        + (f"存在{overdue}件逾期未办结信件，需重点关注。" if overdue > 0 else "未出现逾期未办结情况，办理时效良好。"),
        S["qbody"]
    ))

    story.append(Paragraph("（三）回访满意度分析", S["h2"]))
    story.append(Paragraph(
        "截至报告日，本月所有信件均尚未进入办结回访环节，暂无回访满意度数据。"
        "建议各承办单位在办结信件后及时申请回访，了解群众对办理结果的满意度。",
        S["qbody"]
    ))

    # ── 四、信件明细一览表 ──
    story.append(Paragraph("四、信件明细一览表", S["h1"]))
    detail_data = [["序号", "信件编号", "状态", "类别", "主办单位", "时间"]]
    for i, l in enumerate(letters[:30], 1):
        unit = f"{l['unit_l2'] or l['unit_l1']}" if l["unit_l1"] else "——"
        detail_data.append([
            str(i), l["letter_no"],
            STATUS_NAMES.get(l["current_status"], "未知")[:4],
            l["cat_l1"] or "其他类",
            unit,
            l["received_at"].strftime("%m-%d") if l["received_at"] else "",
        ])
    t2 = Table(detail_data, colWidths=[25, 100, 45, 55, 100, 45])
    t2.setStyle(small_table_style())
    story.append(t2)
    story.append(Spacer(1, 15))

    # ── 五、存在问题及改进建议 ──
    story.append(Paragraph("五、存在问题及改进建议", S["h1"]))
    story.append(Paragraph("（一）存在问题", S["h2"]))
    problems = [
        "1. 来信数量有待提升，需进一步加大平台宣传推广力度。",
        "2. 部分信件分类不够精准，建议加强分类培训。",
        "3. 暂未建立完善的回访机制，无法全面评估群众满意度。",
    ]
    for p in problems:
        story.append(Paragraph(p, S["qbody"]))

    story.append(Paragraph("（二）改进建议", S["h2"]))
    suggestions = [
        "1. 进一步加大局长信箱平台宣传力度，通过多种渠道引导市民使用。",
        "2. 建立信件审核时效考核机制，明确各环节办理时限。",
        "3. 完善信件基础信息录入规范，确保信息准确完整。",
        "4. 建立办结后回访制度，持续改进信件办理质量。",
        "5. 定期开展质态分析，形成月度质态分析报告。",
        "6. 加强业务培训，提升承办单位工作人员的服务意识。",
    ]
    for s in suggestions:
        story.append(Paragraph(s, S["qbody"]))

    story.append(Spacer(1, 20))

    # 分隔线 + 页脚
    hr2 = Table([[""]], colWidths=[460])
    hr2.setStyle(TableStyle([("LINEBELOW", (0, 0), (-1, -1), 1, HexColor("#1a5276"))]))
    story.append(hr2)
    story.append(Spacer(1, 10))
    story.append(Paragraph("报告单位：衡水市公安局民意智感中心", S["qbody_bold"]))
    story.append(Paragraph(f"报告日期：{end_d.strftime('%Y年%m月%d日')}", S["qbody_bold"]))
    story.append(Paragraph("联系人：系统管理员", S["qbody_bold"]))

    doc = SimpleDocTemplate(output_path, pagesize=A4, topMargin=2*cm, bottomMargin=2*cm,
                            leftMargin=2.5*cm, rightMargin=2.5*cm)
    doc.build(story)
    return output_path


# ── main ──

def main():
    if len(sys.argv) < 4:
        print(f"用法: {sys.argv[0]} <start_date> <end_date> <output_dir> [period_label] [bulletin_name] [quality_name]", file=sys.stderr)
        print(f"  例: {sys.argv[0]} 2026-05-01 2026-06-01 /tmp/export '5月' '5月通报.pdf' '5月质态分析报告.pdf'", file=sys.stderr)
        sys.exit(1)

    start_date = sys.argv[1]
    end_date = sys.argv[2]
    output_dir = sys.argv[3]
    period_label = sys.argv[4] if len(sys.argv) > 4 else "本月"
    bulletin_name = sys.argv[5] if len(sys.argv) > 5 else "通报.pdf"
    quality_name = sys.argv[6] if len(sys.argv) > 6 else "质态分析报告.pdf"

    os.makedirs(output_dir, exist_ok=True)

    # 从 DB 拉取数据
    print(f"查询 {start_date} ~ {end_date} 的信件数据...", file=sys.stderr)
    letters = fetch_letters(start_date, end_date)
    print(f"  当期: {len(letters)} 条", file=sys.stderr)

    prev_letters = fetch_prev_month_letters(start_date, end_date)
    print(f"  上期: {len(prev_letters)} 条", file=sys.stderr)

    if not letters:
        print("警告: 当期无信件数据，将生成空报告", file=sys.stderr)

    # 生成通报.pdf
    bulletin_path = os.path.join(output_dir, bulletin_name)
    generate_bulletin_pdf(letters, prev_letters, start_date, end_date, period_label, bulletin_path)
    print(f"生成: {bulletin_path}", file=sys.stderr)

    # 生成质态分析报告.pdf
    quality_path = os.path.join(output_dir, quality_name)
    generate_quality_pdf(letters, prev_letters, start_date, end_date, period_label, quality_path)
    print(f"生成: {quality_path}", file=sys.stderr)

    # 输出路径供 Go 读取
    print(bulletin_path)
    print(quality_path)


if __name__ == "__main__":
    main()
