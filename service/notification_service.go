package service

import (
	"fmt"
	"letter-manage-backend/dao"
)

// SendNotification creates a notification for a specific user
func SendNotification(userID uint, notifType, title, message, letterNo string) {
	_ = dao.NotifyUser(userID, notifType, title, message, letterNo)
}

// NotifyDispatch sends dispatch notification to target unit users
func NotifyDispatch(letterNo string, targetUnitID uint, operatorName string) {
	title := "新信件下发"
	msg := fmt.Sprintf("%s 将信件 %s 下发至你单位", operatorName, letterNo)
	dao.NotifyUnitUsers(targetUnitID, "dispatch", title, msg, letterNo)
}

// NotifyProcessing sends processing notification to the handler
func NotifyHandleBySelf(letterNo string, userID uint, operatorName string) {
	title := "信件待处理"
	msg := fmt.Sprintf("信件 %s 已由 %s 转为自行处理", letterNo, operatorName)
	_ = dao.NotifyUser(userID, "handle_by_self", title, msg, letterNo)
}

// NotifyAudit sends audit notification to the parent unit
func NotifySubmitProcessing(letterNo string, parentUnitID uint, operatorName string) {
	title := "信件待核查"
	msg := fmt.Sprintf("%s 提交了信件 %s 的处理结果，请核查", operatorName, letterNo)
	dao.NotifyUnitUsers(parentUnitID, "audit", title, msg, letterNo)
}

// NotifyAuditResult sends audit result notification
func NotifyAuditResult(letterNo string, unitID uint, action, operatorName string) {
	title := "核查结果"
	actionText := "通过"
	if action == "reject" {
		actionText = "驳回"
	}
	msg := fmt.Sprintf("信件 %s 已被 %s %s", letterNo, operatorName, actionText)
	dao.NotifyUnitUsers(unitID, "audit_result", title, msg, letterNo)
}

// NotifyNewLetter sends notification when a citizen submits a new letter
func NotifyNewLetter(letterNo, citizenName string) {
	title := "新信件上报"
	msg := fmt.Sprintf("群众 %s 上报了信件 %s，请及时处理", citizenName, letterNo)
	// 通知市局民意智感中心（unit_id=1）
	dao.NotifyUnitUsers(1, "new_letter", title, msg, letterNo)
}

// GetUserIDByLetterNo returns the handler's user ID for a letter
func GetHandlerUserID(letterNo string) uint {
	letter, err := dao.GetLetterByNo(letterNo)
	if err != nil || letter == nil || letter.HandlerUserID == nil {
		return 0
	}
	return *letter.HandlerUserID
}
