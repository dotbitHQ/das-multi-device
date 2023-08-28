package handle

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/http_api"
	"github.com/gin-gonic/gin"
	"github.com/scorpiotzh/toolib"
	"io/ioutil"
	"math/big"
	"net/http"
)

type ReqCidInfo struct {
	Cid    string `json:"cid" binding:"required"`
	Pubkey struct {
		X string `json:"x" binding:"required"`
		Y string `json:"y" binding:"required"`
	} `json:"pubkey" binding:"required"`
}

type RepCidInfo struct {
	IsOldCid   bool `json:"is_old_cid"`
	HasAccount bool `json:"has_account"`
}

func (h *HttpHandle) CidInfo(ctx *gin.Context) {
	var (
		funcName = "CidInfo"
		clientIp = GetClientIp(ctx)
		req      ReqCidInfo
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doCidInfo(&req, &apiResp); err != nil {
		log.Error("doCidInfo err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doCidInfo(req *ReqCidInfo, apiResp *http_api.ApiResp) (err error) {
	var resp RepCidInfo
	cid := req.Cid

	var oldCids []OldCid
	oldCids, err = h.getOldCid()
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return err
	}
	for _, v := range oldCids {
		if v.Cid == cid {
			resp.IsOldCid = true
		}
	}

	if resp.IsOldCid {
		curve := elliptic.P256()
		pubkey := new(ecdsa.PublicKey)
		pubkey.Curve = curve
		xBytes, err := hex.DecodeString(req.Pubkey.X)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, err.Error())
			return err
		}
		yBytes, err := hex.DecodeString(req.Pubkey.Y)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, err.Error())
			return err
		}
		pubkey.X = new(big.Int).SetBytes(xBytes)
		pubkey.Y = new(big.Int).SetBytes(yBytes)
		payload := common.GetWebauthnPayload(req.Cid, pubkey)

		res, err := h.checkCanBeCreated(payload)
		if err != nil {
			apiResp.ApiRespErr(http_api.ApiCodeError500, "check .bit  err")
			return fmt.Errorf("checkCanBeCreated err : %s", err.Error())
		}
		resp.HasAccount = res
	}
	apiResp.ApiRespOK(resp)
	return nil
}

type ReqQueryCid struct {
	Account string `json:"account" binding:"required"`
}
type RepQueryCid struct {
	Cid string `json:"cid"`
}

func (h *HttpHandle) QueryCid(ctx *gin.Context) {
	var (
		funcName = "CidInfo"
		clientIp = GetClientIp(ctx)
		req      ReqQueryCid
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doQueryCid(&req, &apiResp); err != nil {
		log.Error("doCidInfo err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}
func (h *HttpHandle) doQueryCid(req *ReqQueryCid, apiResp *http_api.ApiResp) (err error) {
	var resp RepQueryCid
	account := req.Account
	acc, err := h.dbDao.GetAccountInfoByAccount(account)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeDbError, "GetAccountInfoByAccount err")
		return fmt.Errorf("GetAccountInfoByAccount err: %s", err.Error())
	}
	if acc.Id == 0 || acc.OwnerChainType != common.ChainTypeWebauthn {
		apiResp.ApiRespErr(http_api.ApiCodeAccountNotExist, "account is not found or not registered by passkey")
		return
	}
	var oldCids []OldCid

	oldCids, err = h.getOldCid()
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return err
	}
	for _, v := range oldCids {
		if v.Cid1 == acc.Owner[:20] {
			resp.Cid = v.Cid
		}
	}
	apiResp.ApiRespOK(resp)
	return nil
}

//add test cid into config
type ReqAddTestCid struct {
	Cid string `json:"cid" binding:"required"`
}

func (h *HttpHandle) AddTestCid(ctx *gin.Context) {
	var (
		funcName = "AddTestCid"
		clientIp = GetClientIp(ctx)
		req      ReqAddTestCid
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doAddTestCid(&req, &apiResp); err != nil {
		log.Error("doCidInfo err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

type OldCid struct {
	Cid     string `json:"cid"`
	Cid1    string `json:"cid1"`
	IsCover bool   `json:"is_cover"`
	IsTest  bool   `json:"is_test""`
}

func (h *HttpHandle) doAddTestCid(req *ReqAddTestCid, apiResp *http_api.ApiResp) (err error) {
	var oldCids []OldCid

	oldCids, err = h.getOldCid()
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return
	}
	for _, v := range oldCids {
		if v.Cid == req.Cid {
			apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid, The cid already exists")
			return
		}

	}

	oldCids = append(oldCids, OldCid{
		Cid:     req.Cid,
		Cid1:    hex.EncodeToString(common.CalculateCid1(req.Cid)),
		IsCover: false,
		IsTest:  true,
	})
	//add
	content, err := json.Marshal(oldCids)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return
	}
	err = h.updateOldCid(string(content))
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "updateOldCid err")
		return
	}
	return
}

type ReqCoverCid struct {
	Cid string `json:"cid" binding:"required"`
}

func (h *HttpHandle) CoverCid(ctx *gin.Context) {
	var (
		funcName = "CoverCid"
		clientIp = GetClientIp(ctx)
		req      ReqCoverCid
		apiResp  http_api.ApiResp
		err      error
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error("ShouldBindJSON err: ", err.Error(), funcName, clientIp)
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "params invalid")
		ctx.JSON(http.StatusOK, apiResp)
		return
	}
	log.Info("ApiReq:", funcName, clientIp, toolib.JsonString(req))

	if err = h.doCoverCid(&req, &apiResp); err != nil {
		log.Error("doCoverCid err:", err.Error(), funcName, clientIp)
	}

	ctx.JSON(http.StatusOK, apiResp)
}

func (h *HttpHandle) doCoverCid(req *ReqCoverCid, apiResp *http_api.ApiResp) (err error) {
	var oldCids []OldCid
	oldCids, err = h.getOldCid()
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return
	}
	tag := false
	for i, v := range oldCids {
		if v.Cid == req.Cid && !v.IsCover {
			tag = true
			oldCids[i].IsCover = true
		}
	}
	if !tag {
		apiResp.ApiRespErr(http_api.ApiCodeParamsInvalid, "cid not found or has been covered")
		return
	}
	//update
	content, err := json.Marshal(oldCids)
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, err.Error())
		return
	}
	err = h.updateOldCid(string(content))
	if err != nil {
		apiResp.ApiRespErr(http_api.ApiCodeError500, "updateOldCid err")
		return
	}
	return
}

func (h *HttpHandle) getOldCid() (oldCids []OldCid, err error) {
	oldCidPath := "../config/old_cid.json"
	file, err := ioutil.ReadFile(oldCidPath)
	if err != nil {
		log.Fatalf("Some error occured while reading file. Error: %s", err)
		return
	}
	err = json.Unmarshal(file, &oldCids)
	if err != nil {
		log.Fatalf("Error occured during unmarshaling. Error: %s", err.Error())
		return
	}
	return
}
func (h *HttpHandle) updateOldCid(content string) (err error) {
	oldCidPath := "../config/old_cid.json"
	err = ioutil.WriteFile(oldCidPath, []byte(content), 0644)
	return
}
