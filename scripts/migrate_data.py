#!/usr/bin/env python3
"""
数据迁移脚本: 老库 (81.70.230.137:53306/pot_data) -> 新库 (10.25.65.177:8306/letter_manage_db)

迁移映射:
  老库              ->  新库
  单位              ->  units
  police_users      ->  police_users
  user_sessions     ->  user_sessions
  信件表            ->  letters
  流转表            ->  letter_flows
  文件表            ->  letter_attachments
  反馈表            ->  feedbacks
  信件分类表        ->  categories
  下发权限表        ->  dispatch_permissions
  专项关注表        ->  special_focuses
  提示词            ->  prompts
"""

import pymysql
import json
import sys
from datetime import datetime

# ============ 数据库连接配置 ============
OLD_DB = dict(host='81.70.230.137', port=53306, user='pot', password='000001',
              db='pot_data', charset='utf8mb4', cursorclass=pymysql.cursors.DictCursor)
NEW_DB = dict(host='10.25.65.177', port=8306, user='root', password='000000',
              db='letter_manage_db', charset='utf8mb4', cursorclass=pymysql.cursors.DictCursor)


def safe_json(val):
    """将任意 JSON 值安全转为 JSON 字符串"""
    if val is None:
        return None
    if isinstance(val, (dict, list)):
        return json.dumps(val, ensure_ascii=False)
    if isinstance(val, str):
        # 尝试验证是否为有效 JSON
        try:
            json.loads(val)
            return val
        except Exception:
            return json.dumps(val, ensure_ascii=False)
    return json.dumps(val, ensure_ascii=False)


def connect(cfg):
    return pymysql.connect(**cfg)


def migrate_units(old_cur, new_cur):
    """单位 -> units"""
    old_cur.execute("SELECT * FROM `单位`")
    rows = old_cur.fetchall()
    inserted = 0
    skipped = 0
    for r in rows:
        try:
            new_cur.execute("""
                INSERT IGNORE INTO units (level1, level2, level3, system_code)
                VALUES (%s, %s, %s, %s)
            """, (r.get('一级'), r.get('二级'), r.get('三级'), r.get('系统编码')))
            if new_cur.rowcount > 0:
                inserted += 1
            else:
                skipped += 1
        except Exception as e:
            print(f"  [WARN] units insert failed: {e} | row={r}")
    print(f"  units: inserted={inserted}, skipped(duplicate)={skipped}, total={len(rows)}")
    return inserted


def migrate_police_users(old_cur, new_cur):
    """police_users -> police_users (字段名一一对应，password -> password_hash)"""
    old_cur.execute("SELECT * FROM police_users")
    rows = old_cur.fetchall()
    inserted = 0
    skipped = 0
    for r in rows:
        try:
            new_cur.execute("""
                INSERT IGNORE INTO police_users
                  (password_hash, name, nickname, police_number, phone, unit_name,
                   permission_level, is_active, created_at, last_login)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """, (
                r.get('password'), r.get('name'), r.get('nickname'),
                r.get('police_number'), r.get('phone'), r.get('unit_name'),
                r.get('permission_level'), int(r.get('is_active', 1)),
                r.get('created_at'), r.get('last_login')
            ))
            if new_cur.rowcount > 0:
                inserted += 1
            else:
                skipped += 1
        except Exception as e:
            print(f"  [WARN] police_users insert failed: {e} | police_number={r.get('police_number')}")
    print(f"  police_users: inserted={inserted}, skipped={skipped}, total={len(rows)}")
    return inserted


def migrate_user_sessions(old_cur, new_cur):
    """user_sessions -> user_sessions (user_id 对应新库 police_users.id)"""
    # 先建立老库 user_id -> 新库 id 的映射（通过 police_number 关联）
    old_cur.execute("SELECT id, police_number FROM police_users")
    old_users = {r['id']: r['police_number'] for r in old_cur.fetchall()}
    new_cur.execute("SELECT id, police_number FROM police_users")
    new_users = {r['police_number']: r['id'] for r in new_cur.fetchall()}

    old_cur.execute("SELECT * FROM user_sessions")
    rows = old_cur.fetchall()
    inserted = 0
    skipped = 0
    not_found = 0
    for r in rows:
        old_uid = r.get('user_id')
        pn = old_users.get(old_uid)
        new_uid = new_users.get(pn) if pn else None
        if new_uid is None:
            not_found += 1
            print(f"  [WARN] user_sessions: cannot map old user_id={old_uid} (police_number={pn})")
            continue
        try:
            new_cur.execute("""
                INSERT IGNORE INTO user_sessions
                  (user_id, session_key, ip_address, user_agent, created_at, expires_at)
                VALUES (%s, %s, %s, %s, %s, %s)
            """, (
                new_uid, r.get('session_key'), r.get('ip_address'),
                r.get('user_agent'), r.get('created_at'), r.get('expires_at')
            ))
            if new_cur.rowcount > 0:
                inserted += 1
            else:
                skipped += 1
        except Exception as e:
            print(f"  [WARN] user_sessions insert failed: {e} | session_key={r.get('session_key')}")
    print(f"  user_sessions: inserted={inserted}, skipped={skipped}, not_found={not_found}, total={len(rows)}")
    return inserted


def migrate_letters(old_cur, new_cur):
    """信件表 -> letters"""
    old_cur.execute("SELECT * FROM `信件表`")
    rows = old_cur.fetchall()
    inserted = 0
    skipped = 0
    for r in rows:
        try:
            new_cur.execute("""
                INSERT IGNORE INTO letters
                  (letter_no, citizen_name, phone, id_card, received_at, channel,
                   category_l1, category_l2, category_l3, content, special_tags,
                   current_unit, current_status)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """, (
                r.get('信件编号'), r.get('群众姓名'), r.get('手机号'), r.get('身份证号'),
                r.get('来信时间'), r.get('来信渠道'),
                r.get('信件一级分类'), r.get('信件二级分类'), r.get('信件三级分类'),
                r.get('诉求内容'), safe_json(r.get('专项关注标签')),
                r.get('当前信件处理单位'), r.get('当前信件状态')
            ))
            if new_cur.rowcount > 0:
                inserted += 1
            else:
                skipped += 1
        except Exception as e:
            print(f"  [WARN] letters insert failed: {e} | letter_no={r.get('信件编号')}")
    print(f"  letters: inserted={inserted}, skipped={skipped}, total={len(rows)}")
    return inserted


def migrate_letter_flows(old_cur, new_cur):
    """流转表 -> letter_flows"""
    old_cur.execute("SELECT * FROM `流转表`")
    rows = old_cur.fetchall()
    inserted = 0
    skipped = 0
    for r in rows:
        try:
            new_cur.execute("""
                INSERT IGNORE INTO letter_flows
                  (letter_no, flow_records, created_at, updated_at)
                VALUES (%s, %s, %s, %s)
            """, (
                r.get('信件编号'), safe_json(r.get('流转记录')),
                r.get('创建时间'), r.get('更新时间')
            ))
            if new_cur.rowcount > 0:
                inserted += 1
            else:
                skipped += 1
        except Exception as e:
            print(f"  [WARN] letter_flows insert failed: {e} | letter_no={r.get('信件编号')}")
    print(f"  letter_flows: inserted={inserted}, skipped={skipped}, total={len(rows)}")
    return inserted


def migrate_letter_attachments(old_cur, new_cur):
    """文件表 -> letter_attachments"""
    old_cur.execute("SELECT * FROM `文件表`")
    rows = old_cur.fetchall()
    inserted = 0
    skipped = 0
    for r in rows:
        try:
            new_cur.execute("""
                INSERT IGNORE INTO letter_attachments
                  (letter_no, city_dispatch_files, district_dispatch_files,
                   handler_feedback_files, district_feedback_files, call_recordings)
                VALUES (%s, %s, %s, %s, %s, %s)
            """, (
                r.get('信件编号'),
                safe_json(r.get('市局下发附件')),
                safe_json(r.get('区县局下发附件')),
                safe_json(r.get('办案单位反馈附件')),
                safe_json(r.get('区县局反馈附件')),
                safe_json(r.get('通话录音附件'))
            ))
            if new_cur.rowcount > 0:
                inserted += 1
            else:
                skipped += 1
        except Exception as e:
            print(f"  [WARN] letter_attachments insert failed: {e} | letter_no={r.get('信件编号')}")
    print(f"  letter_attachments: inserted={inserted}, skipped={skipped}, total={len(rows)}")
    return inserted


def migrate_feedbacks(old_cur, new_cur):
    """反馈表 -> feedbacks"""
    old_cur.execute("SELECT * FROM `反馈表`")
    rows = old_cur.fetchall()
    inserted = 0
    skipped = 0
    for r in rows:
        try:
            new_cur.execute("""
                INSERT IGNORE INTO feedbacks
                  (letter_no, feedback_info, created_at, updated_at)
                VALUES (%s, %s, %s, %s)
            """, (
                r.get('信件编号'), safe_json(r.get('反馈信息')),
                r.get('创建时间'), r.get('更新时间')
            ))
            if new_cur.rowcount > 0:
                inserted += 1
            else:
                skipped += 1
        except Exception as e:
            print(f"  [WARN] feedbacks insert failed: {e} | letter_no={r.get('信件编号')}")
    print(f"  feedbacks: inserted={inserted}, skipped={skipped}, total={len(rows)}")
    return inserted


def migrate_categories(old_cur, new_cur):
    """信件分类表 -> categories"""
    old_cur.execute("SELECT * FROM `信件分类表`")
    rows = old_cur.fetchall()
    inserted = 0
    skipped = 0
    for r in rows:
        try:
            new_cur.execute("""
                INSERT INTO categories
                  (level1, level2, level3, created_at, updated_at)
                VALUES (%s, %s, %s, %s, %s)
                ON DUPLICATE KEY UPDATE
                  level1=VALUES(level1), level2=VALUES(level2), level3=VALUES(level3),
                  updated_at=VALUES(updated_at)
            """, (
                r.get('一级分类'), r.get('二级分类'), r.get('三级分类'),
                r.get('创建时间'), r.get('更新时间')
            ))
            if new_cur.rowcount == 1:
                inserted += 1
            else:
                skipped += 1
        except Exception as e:
            print(f"  [WARN] categories insert failed: {e} | row={r}")
    print(f"  categories: inserted={inserted}, updated/skipped={skipped}, total={len(rows)}")
    return inserted


def migrate_dispatch_permissions(old_cur, new_cur):
    """下发权限表 -> dispatch_permissions"""
    old_cur.execute("SELECT * FROM `下发权限表`")
    rows = old_cur.fetchall()
    inserted = 0
    skipped = 0
    for r in rows:
        try:
            new_cur.execute("""
                INSERT INTO dispatch_permissions
                  (unit_name, dispatch_scope, created_at, updated_at)
                VALUES (%s, %s, %s, %s)
                ON DUPLICATE KEY UPDATE
                  dispatch_scope=VALUES(dispatch_scope),
                  updated_at=VALUES(updated_at)
            """, (
                r.get('单位全称'), safe_json(r.get('下发范围')),
                r.get('创建时间'), r.get('最后更新时间')
            ))
            if new_cur.rowcount == 1:
                inserted += 1
            else:
                skipped += 1
        except Exception as e:
            print(f"  [WARN] dispatch_permissions insert failed: {e} | unit={r.get('单位全称')}")
    print(f"  dispatch_permissions: inserted={inserted}, updated/skipped={skipped}, total={len(rows)}")
    return inserted


def migrate_special_focuses(old_cur, new_cur):
    """专项关注表 -> special_focuses"""
    old_cur.execute("SELECT * FROM `专项关注表`")
    rows = old_cur.fetchall()
    inserted = 0
    skipped = 0
    for r in rows:
        try:
            new_cur.execute("""
                INSERT IGNORE INTO special_focuses
                  (tag_name, description, created_at, updated_at)
                VALUES (%s, %s, %s, %s)
            """, (
                r.get('专项关注标题'), r.get('描述'),
                r.get('创建时间'), r.get('最后更新时间')
            ))
            if new_cur.rowcount > 0:
                inserted += 1
            else:
                skipped += 1
        except Exception as e:
            print(f"  [WARN] special_focuses insert failed: {e} | tag={r.get('专项关注标题')}")
    print(f"  special_focuses: inserted={inserted}, skipped={skipped}, total={len(rows)}")
    return inserted


def migrate_prompts(old_cur, new_cur):
    """提示词 -> prompts"""
    old_cur.execute("SELECT * FROM `提示词`")
    rows = old_cur.fetchall()
    inserted = 0
    skipped = 0
    for r in rows:
        try:
            new_cur.execute("""
                INSERT IGNORE INTO prompts
                  (prompt_type, content, created_at, updated_at)
                VALUES (%s, %s, %s, %s)
            """, (
                r.get('类型'), r.get('内容'),
                r.get('created_at'), r.get('updated_at')
            ))
            if new_cur.rowcount > 0:
                inserted += 1
            else:
                skipped += 1
        except Exception as e:
            print(f"  [WARN] prompts insert failed: {e} | type={r.get('类型')}")
    print(f"  prompts: inserted={inserted}, skipped={skipped}, total={len(rows)}")
    return inserted


def main():
    print("=" * 60)
    print(f"迁移开始: {datetime.now()}")
    print("老库: 81.70.230.137:53306/pot_data")
    print("新库: 10.25.65.177:8306/letter_manage_db")
    print("=" * 60)

    old_conn = connect(OLD_DB)
    new_conn = connect(NEW_DB)

    try:
        old_cur = old_conn.cursor()
        new_cur = new_conn.cursor()

        # 按依赖顺序迁移
        migrations = [
            ("units",                migrate_units),
            ("police_users",         migrate_police_users),
            ("user_sessions",        migrate_user_sessions),
            ("letters",              migrate_letters),
            ("letter_flows",         migrate_letter_flows),
            ("letter_attachments",   migrate_letter_attachments),
            ("feedbacks",            migrate_feedbacks),
            ("categories",           migrate_categories),
            ("dispatch_permissions", migrate_dispatch_permissions),
            ("special_focuses",      migrate_special_focuses),
            ("prompts",              migrate_prompts),
        ]

        total_inserted = 0
        for name, fn in migrations:
            print(f"\n[{name}]")
            try:
                n = fn(old_cur, new_cur)
                new_conn.commit()
                total_inserted += n
            except Exception as e:
                new_conn.rollback()
                print(f"  [ERROR] {name} 迁移失败，已回滚: {e}")
                import traceback
                traceback.print_exc()

        print("\n" + "=" * 60)
        print(f"迁移完成: {datetime.now()}")
        print(f"共插入记录: {total_inserted}")
        print("=" * 60)

    finally:
        old_cur.close()
        old_conn.close()
        new_cur.close()
        new_conn.close()


if __name__ == '__main__':
    main()
