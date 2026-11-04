package service

import errContract "bit-labs.cn/owl/contract/errors"

const (
	CodeSettingUnknown = "SETTING_UNKNOWN"
	CodeSettingInvalid = "SETTING_INVALID"
)

// SettingUnknown 表示子系统或配置键没有登记。
func SettingUnknown(msg string) *errContract.BizError {
	return errContract.NewBizError(CodeSettingUnknown, msg)
}

// SettingInvalid 表示配置值不符合登记时写死的规则。
func SettingInvalid(msg string) *errContract.BizError {
	if msg == "" {
		msg = "配置值不正确"
	}
	return errContract.NewBizError(CodeSettingInvalid, msg)
}
