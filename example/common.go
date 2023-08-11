package example

import (
	"fmt"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/parnurzeal/gorequest"
)

const (
	TestUrl = "http://localhost:8125/v1"
)

func doReq(url string, req, data interface{}) error {
	var resp http_api.ApiResp
	resp.Data = &data
	_, _, errs := gorequest.New().Post(url).SendStruct(&req).EndStruct(&resp)
	if errs != nil {
		return fmt.Errorf("%v", errs)
	}
	if resp.ErrNo != http_api.ApiCodeSuccess {
		return fmt.Errorf("%d - %s", resp.ErrNo, resp.ErrMsg)
	}
	return nil
}
