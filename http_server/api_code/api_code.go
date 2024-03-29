package api_code

//
//import "github.com/dotbitHQ/das-lib/http_api"
//
//type ApiCode = int
//
//const (
//	ApiCodeSuccess        ApiCode = 0
//	ApiCodeError500       ApiCode = 500
//	ApiCodeParamsInvalid  ApiCode = 10000
//	ApiCodeMethodNotExist ApiCode = 10001
//	ApiCodeDbError        ApiCode = 10002
//	ApiCodeCacheError     ApiCode = 10003
//
//	ApiCodeTransactionNotExist          ApiCode = 11001
//	ApiCodeInsufficientBalance          ApiCode = 11007
//	ApiCodeTxExpired                    ApiCode = 11008
//	ApiCodeRejectedOutPoint             ApiCode = 11011
//	ApiCodeSyncBlockNumber              ApiCode = 11012
//	ApiCodeOperationFrequent            ApiCode = 11013
//	ApiCodeNotEnoughChange              ApiCode = 11014
//	ApiCodeAccountNotExist              ApiCode = 30003
//	ApiCodeAccountIsExpired             ApiCode = 30010
//	ApiCodePermissionDenied             ApiCode = 30011
//	ApiCodeSystemUpgrade                ApiCode = 30019
//	ApiCodeRecordInvalid                ApiCode = 30020
//	ApiCodeRecordsTotalLengthExceeded   ApiCode = 30021
//	ApiCodeSameLock                     ApiCode = 30023
//	ApiCodeAccountStatusOnSaleOrAuction ApiCode = 30031
//	ApiCodeOnCross                      ApiCode = 30035
//
//	ApiCodeEnableSubAccountIsOn            ApiCode = 40000
//	ApiCodeNotExistEditKey                 ApiCode = 40001
//	ApiCodeNotExistConfirmAction           ApiCode = 40002
//	ApiCodeSignError                       ApiCode = 40003
//	ApiCodeNotExistSignType                ApiCode = 40004
//	ApiCodeNotSubAccount                   ApiCode = 40005
//	ApiCodeEnableSubAccountIsOff           ApiCode = 40006
//	ApiCodeCreateListCheckFail             ApiCode = 40007
//	ApiCodeTaskInProgress                  ApiCode = 40008
//	ApiCodeDistributedLockPreemption       ApiCode = 40009
//	ApiCodeRecordDoing                     ApiCode = 40010
//	ApiCodeUnableInit                      ApiCode = 40011
//	ApiCodeNotHaveManagementPermission     ApiCode = 40012
//	ApiCodeSmtDiff                         ApiCode = 40013
//	ApiCodeSuspendOperation                ApiCode = 40014
//	ApiCodeTaskNotExist                    ApiCode = 40015
//	ApiCodeSameCustomScript                ApiCode = 40016
//	ApiCodeNotExistCustomScriptConfigPrice ApiCode = 40017
//	ApiCodeCustomScriptSet                 ApiCode = 40018
//	ApiCodeProfitNotEnough                 ApiCode = 40019
//
//	ApiCodeHasNoAccessToCreate  ApiCode = 60000
//	ApiCodeCreateConfigCellFail ApiCode = 60001
//	ApiCodeHasNoAccessToRemove  ApiCode = 60002
//)
//
//const (
//	MethodEcdsaRecover      = "das_ecdsa_ecrecover"
//	MethodGetMasterAddr     = "das_get_masters_addr"
//	MethodGetOriginPk       = "das_get_origin_pk"
//	MethodAuthorize         = "das_authorize"
//	MethodTransactionSend   = "das_transactionSend"
//	MethodTransactionStatus = "das_transactionStatus"
//)
//
//const (
//	TextSystemUpgrade = "The service is under maintenance, please try again later."
//)
//
//type ApiResp struct {
//	ErrNo  ApiCode     `json:"err_no"`
//	ErrMsg string      `json:"err_msg"`
//	Data   interface{} `json:"data"`
//}
//
//func (a *ApiResp) ApiRespErr(errNo ApiCode, errMsg string) {
//	a.ErrNo = errNo
//	a.ErrMsg = errMsg
//}
//
//func (a *ApiResp) ApiRespOK(data interface{}) {
//	a.ErrNo = http_api.ApiCodeSuccess
//	a.Data = data
//}
//
//func ApiRespErr(errNo ApiCode, errMsg string) ApiResp {
//	return ApiResp{
//		ErrNo:  errNo,
//		ErrMsg: errMsg,
//		Data:   nil,
//	}
//}
