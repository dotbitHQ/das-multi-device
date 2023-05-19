package example

import (
	"das-multi-device/http_server/handle"
	"fmt"
	"testing"
)

func TestEcdsaRecover(t *testing.T) {
	req := handle.ReqEcrecover{
		Cid: "VCmB7OMwtqOtoRKlZRiQ1RvfQryMJu12wSAt3Zf_5EU",
		SignData: []*handle.WebauthnSignData{
			{
				AuthenticatorData: "49960de5880e8c687434170f6476605b8fe4aeb9a28632c7995cf3ba831d97630500000000",
				ClientDataJson:    "7b2274797065223a22776562617574686e2e676574222c226368616c6c656e6765223a2259574a6a222c226f726967696e223a22687474703a2f2f6c6f63616c686f7374222c2263726f73734f726967696e223a66616c73657d",
				Signature:         "304402205f068d44525440ad9f3896d57e0a7cdb253240cd54726aa5e7bb2c7044228871022064706d26ec7bfb19f9d35d19117d1e879c5b8be3dec8f0e83aa0e47b9034c3f7",
			},
			{
				AuthenticatorData: "49960de5880e8c687434170f6476605b8fe4aeb9a28632c7995cf3ba831d97630500000000",
				ClientDataJson:    "7b2274797065223a22776562617574686e2e676574222c226368616c6c656e6765223a2259574a6a222c226f726967696e223a22687474703a2f2f6c6f63616c686f7374222c2263726f73734f726967696e223a66616c73657d",
				Signature:         "3045022100e61fa1ccc54615849a6a10f1f1567648ea499bdab0136e6162c3f59a94bb8c5a022016f9009bad0435acb2ccdf1a3beb40a01ea359a1283ba62db5c04e4402bcfb4d",
			},
		},
	}
	url := TestUrl + "/webauthn/ecdsa-ecrecover"
	var data handle.RespEcrecover
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}
func TestGetMasterAddr(t *testing.T) {
	req := handle.ReqGetMasters{
		Cid: "VCmB7OMwtqOtoRKlZRiQ1RvfQryMJu12wSAt3Zf_5EU",
	}
	url := TestUrl + "/webauthn/get-masters-addr"
	var data handle.RespGetMasters
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}
func TestAuthorize(t *testing.T) {
	req := handle.ReqAuthorize{
		MasterCkbAddress: "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqu4qyfuzauwmj9k6qeenhmyt039rhu5xaqyqw2szy7pw78dezmdqvuemaj9hcj3m72rwsv94j9m",
		SlaveCkbAddress:  "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqu4qyfuzauwmj9k6qeenhmyt039rhu5xaqyqw2szy7pw78dezmdqvuemaj9hcj3m72rwsv94j9m",
	}
	url := TestUrl + "/webauthn/authorize"
	var data handle.ReqAuthorize
	if err := doReq(url, req, &data); err != nil {
		t.Fatal(err)
	}
	fmt.Println(data)
}
