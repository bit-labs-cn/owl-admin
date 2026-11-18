package service

import errContract "bit-labs.cn/owl/contract/errors"

const (
	CodeNotFound      = "NOTIFY_NOT_FOUND"
	CodeInvalid       = "NOTIFY_INVALID"
	CodeConflict      = "NOTIFY_CONFLICT"
	CodeNotAllowed    = "NOTIFY_NOT_ALLOWED"
	CodeChannelMiss   = "NOTIFY_CHANNEL_MISSING"
	CodeTicketInvalid = "NOTIFY_TICKET_INVALID"
)

func NotFound(msg string) *errContract.BizError {
	return errContract.NewBizError(CodeNotFound, msg)
}

func Invalid(msg string) *errContract.BizError {
	return errContract.NewBizError(CodeInvalid, msg)
}

func Conflict(msg string) *errContract.BizError {
	return errContract.NewBizError(CodeConflict, msg)
}

func NotAllowed(msg string) *errContract.BizError {
	return errContract.NewBizError(CodeNotAllowed, msg)
}

func ChannelMissing(msg string) *errContract.BizError {
	return errContract.NewBizError(CodeChannelMiss, msg)
}

func TicketInvalid() *errContract.BizError {
	return errContract.NewBizError(CodeTicketInvalid, "实时连接票据无效或已过期")
}
