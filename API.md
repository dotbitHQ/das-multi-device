* [API LIST](#api-list)
    * [Ecdsa Ecrecover](#ecdsa-ecrecover)
    * [Get Master Addr](#get-masters-addr)
    * [Authorize](#authorize)
    * [Transaction Send](#transaction-send)
    * [Transaction Status](#transaction-status)
## API LIST
### Ecdsa Ecrecover

#### Request

* path: /v1/webauthn/ecdsa-ecrecover
* params:
  * cid: credential id  (string, necessary)
  * sign_data:
    * authenticatorData: The hexadecimal representation of authenticatorData in the webauthn.get() response   (string, necessary)
    * clientDataJSON: The hexadecimal representation of clientDataJSON in the webauthn.get() response  (string, necessary)
    * signature:The hexadecimal representation of signature in the webauthn.get() response  (string, necessary)

```json
{
  "cid": "VCmB7OMwtqOtoRKlZRiQ1RvfQryMJu12wSAt3Zf_5EU",
  "sign_data": [
    {
      "authenticatorData": "49960de5880e8c687434170f6476605b8fe4aeb9a28632c7995cf3ba831d97630500000000",
      "clientDataJSON": "7b2274797065223a22776562617574686e2e676574222c226368616c6c656e6765223a2259574a6a222c226f726967696e223a22687474703a2f2f6c6f63616c686f7374222c2263726f73734f726967696e223a66616c73657d",
      "signature": "304402205f068d44525440ad9f3896d57e0a7cdb253240cd54726aa5e7bb2c7044228871022064706d26ec7bfb19f9d35d19117d1e879c5b8be3dec8f0e83aa0e47b9034c3f7"
    },
    {
      "authenticatorData": "49960de5880e8c687434170f6476605b8fe4aeb9a28632c7995cf3ba831d97630500000000",
      "clientDataJSON": "7b2274797065223a22776562617574686e2e676574222c226368616c6c656e6765223a2259574a6a222c226f726967696e223a22687474703a2f2f6c6f63616c686f7374222c2263726f73734f726967696e223a66616c73657d",
      "signature": "3045022100e61fa1ccc54615849a6a10f1f1567648ea499bdab0136e6162c3f59a94bb8c5a022016f9009bad0435acb2ccdf1a3beb40a01ea359a1283ba62db5c04e4402bcfb4d"
    }
  ]
}
```

#### Response
* ckb_address: Calculate the CKB address based on the recovered public key
```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "ckb_address": "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqu4qyfuzauwmj9k6qeenhmyt039rhu5xaqyqw2szy7pw78dezmdqvuemaj9hcj3m72rwsv94j9m"
  }
}
```

### Get Master Addr

#### Request

* path: /v1/webauthn/get-masters-addr
* params:
    * cid: credential id (string, necessary)
```json
{
  "cid": "VCmB7OMwtqOtoRKlZRiQ1RvfQryMJu12wSAt3Zf_5EU"
}
```

#### Response
* ckb_address: The CKB address authorized to this cid
```json
{
  "errno": 0,
  "errmsg": "",
  "data": {
    "ckb_address": [
      "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqu4qyfuzauwmj9k6qeenhmyt039rhu5xaqyqw2szy7pw78dezmdqvuemaj9hcj3m72rwsv94j9m",
      "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qg96smm36w2zm7wyjlnykrkps6kwg2zz0z6qh2r0w8fegt0ecjt7vjcwcxr2eepggfutgvl8jl7"
    ]
  }
}
```

### Authorize

#### Request

* path: /v1/webauthn/authorize
* params:
    * master_ckb_address: Authorized CKB address (string, necessary)
    * slave_ckb_address: CKB address to be authorized (string, necessary)
```json
{
  "master_ckb_address":"ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqu4qyfuzauwmj9k6qeenhmyt039rhu5xaqyqw2szy7pw78dezmdqvuemaj9hcj3m72rwsv94j9m",
  "slave_ckb_address":"ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqu4qyfuzauwmj9k6qeenhmyt039rhu5xaqyqw2szy7pw78dezmdqvuemaj9hcj3m72rwsv94j9m"
}
```

#### Response

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "a7b9e04717dfcb729a0becde959e49d3",
    "sign_list": [{
      "sign_type": 8,
      "sign_msg": "From .bit: 0cd1832efdd772927f8b3fb6274fbc33558d942bbfe19061daf19c456cea60af"
    }]
  }
}
```

### Transaction Send

#### Request
**Request**

* path: /transaction/send
* param:
  * sign_type:签名类型，webauthn是8
  * sign_address:签名的ckb地址
  * sign_msg:签名 
    * 签名是用lv的格式将webauthn.get()（签名方法）同步响应的signature, authnticatorData, clientDataJSON 三个字段按如下规则进行拼接
    * len(signature) + signature + len(authnticatorData) + authnticatorData + len(sha256(clientDataJSON)) + sha256(clientDataJSON)
    * 其中len(*) 为1个字节，signature 为64字节，sha256(clientDataJSON)为32字节，authnticatorData为37字节
```json
{
  "sign_key": "",
  "sign_list": [
    {
      "sign_type": 8,
      "sign_address": "ckt1qyqxgwjt9gn7vgk2rnny5lf33dtak4nexkasjsje75",
      "sign_msg": ""
    }
  ]
}
```

**Response**

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "hash": "0x7f620f1c709879034df1a447c303efa0dede62725273e11046a587e174c46ff3"
  }
}
```

#### Transaction Status

**Request**

* path: /transaction/status
* param:
  * actions: business transaction type
    * ActionUpdate_device_key_list TxAction = 30  // withdraw

```json
{
  "chain_type": 1,
  "address": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
  "actions": [
    30
  ]
}
```

**Response**
* status: tx pending

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "block_number": 0,
    "hash": "0x0343c250842fc57daef9fc30e5b9e1270c753a43215db46b19563c588417fcae",
    "action": 30,
    "status": 0
  }
}
```

```json
{
  "err_no": 11001,
  "err_msg": "not exits tx",
  "data": null
}
```