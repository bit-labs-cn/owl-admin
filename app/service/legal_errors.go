package service

import errContract "bit-labs.cn/owl/contract/errors"

const (
	CodeLegalKeyInvalid           = "LEGAL_KEY_INVALID"
	CodeLegalDocumentNotFound     = "LEGAL_DOCUMENT_NOT_FOUND"
	CodeLegalDocumentExists       = "LEGAL_DOCUMENT_EXISTS"
	CodeLegalDocumentHasEvents    = "LEGAL_DOCUMENT_HAS_EVENTS"
	CodeLegalVersionInvalid       = "LEGAL_VERSION_INVALID"
	CodeLegalVersionNotFound      = "LEGAL_VERSION_NOT_FOUND"
	CodeLegalVersionExists        = "LEGAL_VERSION_EXISTS"
	CodeLegalVersionFrozen        = "LEGAL_VERSION_FROZEN"
	CodeLegalVersionNotDraft      = "LEGAL_VERSION_NOT_DRAFT"
	CodeLegalVersionNotCurrent    = "LEGAL_VERSION_NOT_CURRENT"
	CodeLegalSceneNotFound        = "LEGAL_SCENE_NOT_FOUND"
	CodeLegalSceneExists          = "LEGAL_SCENE_EXISTS"
	CodeLegalSceneItemInvalid     = "LEGAL_SCENE_ITEM_INVALID"
	CodeLegalWithdrawDenied       = "LEGAL_WITHDRAW_DENIED"
	CodeLegalAcceptanceIncomplete = "LEGAL_ACCEPTANCE_INCOMPLETE"
)

func LegalKeyInvalid() *errContract.BizError {
	return errContract.NewBizError(CodeLegalKeyInvalid, "标识只能用小写字母、数字和连字符，并以字母开头")
}

func LegalDocumentNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeLegalDocumentNotFound, "条款不存在或未发布")
}

func LegalDocumentExists() *errContract.BizError {
	return errContract.NewBizError(CodeLegalDocumentExists, "这个条款标识已经有了")
}

func LegalDocumentHasEvents() *errContract.BizError {
	return errContract.NewBizError(CodeLegalDocumentHasEvents, "已有同意记录，不能删除，只能停用")
}

func LegalVersionNumberInvalid() *errContract.BizError {
	return errContract.NewBizError(CodeLegalVersionInvalid, "版本号只能包含字母、数字、点、下划线和连字符")
}

func LegalContentEmpty() *errContract.BizError {
	return errContract.NewBizError(CodeLegalVersionInvalid, "正文不能为空")
}

func LegalVersionNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeLegalVersionNotFound, "版本不存在")
}

func LegalVersionExists() *errContract.BizError {
	return errContract.NewBizError(CodeLegalVersionExists, "这个版本号已经有了")
}

func LegalVersionFrozen() *errContract.BizError {
	return errContract.NewBizError(CodeLegalVersionFrozen, "已发布的版本不能修改")
}

func LegalVersionNotDraft() *errContract.BizError {
	return errContract.NewBizError(CodeLegalVersionNotDraft, "只有草稿可以这样操作")
}

func LegalVersionNotCurrent() *errContract.BizError {
	return errContract.NewBizError(CodeLegalVersionNotCurrent, "提交的不是当前已发布版本")
}

func LegalSceneNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeLegalSceneNotFound, "场景不存在")
}

func LegalSceneExists() *errContract.BizError {
	return errContract.NewBizError(CodeLegalSceneExists, "这个场景标识已经有了")
}

func LegalSceneItemInvalid(msg string) *errContract.BizError {
	return errContract.NewBizError(CodeLegalSceneItemInvalid, msg)
}

func LegalWithdrawDenied() *errContract.BizError {
	return errContract.NewBizError(CodeLegalWithdrawDenied, "合同条款不能撤回同意")
}

func LegalAcceptanceIncomplete(name string) *errContract.BizError {
	return errContract.NewBizError(CodeLegalAcceptanceIncomplete, "还没有同意"+name)
}
