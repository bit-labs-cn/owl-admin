package service

import (
	errContract "bit-labs.cn/owl/contract/errors"
)

// 定时任务相关的业务错误码。
const (
	CodeSchedulerJobNotFound       = "SCHEDULER_JOB_NOT_FOUND"
	CodeSchedulerJobNotAllowed     = "SCHEDULER_JOB_NOT_ALLOWED"
	CodeSchedulerBuiltinImmutable  = "SCHEDULER_BUILTIN_IMMUTABLE"
	CodeSchedulerJobRunning        = "SCHEDULER_JOB_RUNNING"
	CodeSchedulerRunNotFound       = "SCHEDULER_RUN_NOT_FOUND"
	CodeSchedulerRunNotRetryable   = "SCHEDULER_RUN_NOT_RETRYABLE"
	CodeSchedulerRunNotCancelable  = "SCHEDULER_RUN_NOT_CANCELABLE"
	CodeSchedulerRunRunning        = "SCHEDULER_RUN_RUNNING"
	CodeSchedulerCronInvalid       = "SCHEDULER_CRON_INVALID"
	CodeSchedulerScriptInvalid     = "SCHEDULER_SCRIPT_INVALID"
	CodeSchedulerPayloadInvalid    = "SCHEDULER_PAYLOAD_INVALID"
	CodeSchedulerEngineUnavailable = "SCHEDULER_ENGINE_UNAVAILABLE"
)

func SchedulerJobNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerJobNotFound, "定时任务不存在")
}

// SchedulerJobNotAllowed 在任务状态不允许当前操作时返回。
func SchedulerJobNotAllowed(msg string) *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerJobNotAllowed, msg)
}

func SchedulerBuiltinImmutable() *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerBuiltinImmutable, "子系统内置任务不可修改或删除，只能启停与手动触发")
}

func SchedulerJobRunning() *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerJobRunning, "任务仍在执行，请稍后再操作")
}

func SchedulerRunNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerRunNotFound, "执行记录不存在")
}

func SchedulerRunNotRetryable() *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerRunNotRetryable, "只有失败、超时或已取消的执行记录可以重试")
}

func SchedulerRunNotCancelable() *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerRunNotCancelable, "该执行记录不能取消")
}

func SchedulerRunRunning() *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerRunRunning, "执行记录正在其它实例上运行，无法取消")
}

func SchedulerCronInvalid(msg string) *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerCronInvalid, msg)
}

func SchedulerScriptInvalid(msg string) *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerScriptInvalid, msg)
}

func SchedulerPayloadInvalid(msg string) *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerPayloadInvalid, msg)
}

func SchedulerEngineUnavailable() *errContract.BizError {
	return errContract.NewBizError(CodeSchedulerEngineUnavailable, "定时任务引擎未启用，无法执行该操作")
}
