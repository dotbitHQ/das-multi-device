package example

import (
	"das-multi-device/http_server/api_code"
	"fmt"
	"github.com/parnurzeal/gorequest"
)

const (
	TestUrl = "http://localhost:8125/v1"
)

func doReq(url string, req, data interface{}) error {
	var resp api_code.ApiResp
	resp.Data = &data
	_, _, errs := gorequest.New().Post(url).SendStruct(&req).EndStruct(&resp)
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	if resp.ErrNo != api_code.ApiCodeSuccess {
		return fmt.Errorf("%d - %s", resp.ErrNo, resp.ErrMsg)
	}
	return nil
}
