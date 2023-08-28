* [API LIST](#api-list)
    * [Ecdsa Ecrecover](#ecdsa-ecrecover)
    * [Get Master Addr](#get-masters-addr)
    * [Get Original Pk](#get-original-pk)
    * [Authorize](#authorize)
    * [Authorize Info](#authorize-info)
    * [Caculate CkbAddr](#caculate-ckbaddr)
    * [Transaction Send](#transaction-send)
    * [Transaction Status](#transaction-status)
    * [Webauthn Verify](#webautn-verify)
## API LIST
test api url https://test-webauthn-api.did.id

prod api url https://webauthn-api.did.id
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
  "cid": "ae8836575d7d139c19525ad11d9d5a77216525e0e50d483caa7b21613973f87a",
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
  "cid": "ae8836575d7d139c19525ad11d9d5a77216525e0e50d483caa7b21613973f87a"
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
### Get Original Pk

#### Request

* path: /v1/webauthn/get-original-pk
* params:
  * cid: credential id (string, necessary)
```json
{
  "cid": "0e259e88f8e40a1e6ba097acecb342c7b209b058c355442a6b3c5f73bf57fd59"
}
```

#### Response
* origin_pk: the original publicKey (ecdsa P256)
```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "origin_pk": "0xa6ae565f4a6137a8ed08e33988cbbe24698ea906ec84215ce042e4812c19502f33b03f6bcc027b41f503f2d25de9e346591cbd03aef5ce5826b3151fdc2aec21"
  }
}
```


### Authorize

#### Request

* path: /v1/webauthn/authorize
* params:
    * master_ckb_address: Authorized CKB address (string, necessary)
    * slave_ckb_address: CKB address to be authorized (string, necessary)
    * operation: add: add autorize，delete：delete autorize (string necessary)
```json
{
  "master_ckb_address":"ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqu4qyfuzauwmj9k6qeenhmyt039rhu5xaqyqw2szy7pw78dezmdqvuemaj9hcj3m72rwsv94j9m",
  "slave_ckb_address":"ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqu4qyfuzauwmj9k6qeenhmyt039rhu5xaqyqw2szy7pw78dezmdqvuemaj9hcj3m72rwsv94j9m"
  "operation" : "add"
}
```

#### Response

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "sign_key": "131fa067a0f34135898f1a85104bccf4",
    "sign_list": [
      {
        "sign_type": 8,
        "sign_msg": "0xea460b7fecf50e3cce7f25631e10da3e8c9e330b861ad5cbbfbfbcb5c14f1206"
      }
    ]
  }
}
```

### Authorize Info

#### Request

* path: /v1/webauthn/authorize-info
* params:
  * ckb_address: CKB address (string, necessary)
```json
{
  "ckb_address":"ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqu4qyfuzauwmj9k6qeenhmyt039rhu5xaqyqw2szy7pw78dezmdqvuemaj9hcj3m72rwsv94j9m"
}
```

#### Response
  * can_authorize: permission to enable backup
  * ckb_address
```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "can_authorize" :1,
    "ckb_address": ["ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqu4qyfuzauwmj9k6qeenhmyt039rhu5xaqyqw2szy7pw78dezmdqvuemaj9hcj3m72rwsv94j9m"]
  }
}
```


### Caculate CkbAddr

#### Request

* path: /v1/webauthn/caculate-ckbaddr
* params:
  * cid: cid (string, necessary)
  * pubkey
    * x: hex x of ecdsa pubkey (string, necessary)
    * y: hex y of ecdsa pubkey (string, necessary)
```json
{
  "cid":"ae8836575d7d139c19525ad11d9d5a77216525e0e50d483caa7b21613973f87a",
  "pubkey":{
    "x":"e03f17de734abd6e39fd2e950d74cd2692d26f1906537d68063e9fce4929bd78",
    "y":"77d16a61c64bba3277040c8bdbc4aa96bee28c39b3af7d012ff99c690b950694"
  }
}
```

#### Response

```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "ckb_address": "ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqacs8j3mxpxj9v8qcwmaaxndrxpx2mamagyqwugreganqnfzkrsv8d77nfk3nqn9d7a75ntlrnc"
  }
}
```



### Transaction Send

#### Request
**Request**

* path: /transaction/send
* param:
  * sign_type:sign type，webauthn是8
  * sign_address: signed ckb address
  * sign_msg:signature 
    * LV : webauthn.get() signature, authnticatorData, clientDataJSON Splice the three fields according to the following rules
    * len(signature) + signature + len(pubkey) + pubkey + len(authnticatorData) + authnticatorData + len(clientDataJSON) + clientDataJSON
    * len(signature) 1byte，len(pubkey) 1byte，len(authnticatorData)为1byte，len(clientDataJSON)为2byte（Little Endian）
```json
{
  "sign_key": "131fa067a0f34135898f1a85104bccf4",
  "sign_list": [
    {
      "sign_type": 8,
      "sign_msg": "40ebc710aca131c7289cb74470ee12d6e9b75a160debcf17c486bfe9c37ad24ddd4efb14f5224b90c46cc0fc66c9a4fdc8255dd27575dc53589408c489fe9a8332403edc4f6d1ba31cae2f8e7af0b6d82d79c695576c837df0105bf4d1d785831c556dd11a9ccb13ddafa8c9c978c9a98b4e74579956ffd36b2a00f09f858e8a22242549960de5880e8c687434170f6476605b8fe4aeb9a28632c7995cf3ba831d97630500000000b1007b2274797065223a22776562617574686e2e676574222c226368616c6c656e6765223a22596a55304d546b34596a45784e7a49354d5463795a4463774f4449314d4467784f4445314e445931595464694f54466d597a63775a5752685a574a6a595755305a446b334d6a67334d4463784d3255785a5441334e67222c226f726967696e223a22687474703a2f2f6c6f63616c686f73743a38303031222c2263726f73734f726967696e223a66616c73657d"
    }
  ],
  "sign_address":"ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggquyxv8jked54atrex9zwks38g48fy73vdsyqwzrxretvk62743unz38tggn52n5j0gkxcmk8jru"
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
  * tx_hash: if this param is not empty, it will return the transaction of this tx_hash, otherwise the transaction queried based on action and address will be returned
  * actions: business transaction type
    * ActionUpdate_device_key_list TxAction = 30  // withdraw
  * address: address
```json
{
  "tx_hash":"0x0dc2e55e524a2558cb822d53ae76c8c058c7522602331553a39c7fddf28326ad",
  "actions":[30],
  "chain_type": 8,
  "address":"ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggquyxv8jked54atrex9zwks38g48fy73vdsyqwzrxretvk62743unz38tggn52n5j0gkxcmk8jru"
}
```

**Response**
* status: 0:default(pending) -1:rejected 1:confirm

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

### Webauthn Verify

#### Request

* path: /v1/webauthn/verify
* params:
  * master_addr: login ckb address
  * backup_addr: sign ckb address
  * msg: sign msg
  * signature： 
    * LV format: webauthn.get() signature, authnticatorData, clientDataJSON Splice the three fields according to the following rules
    * len(signature) + signature + len(pubkey) + pubkey + len(authnticatorData) + authnticatorData + len(clientDataJSON) + clientDataJSON
    * len(signature) 1byte，len(pubkey) 1byte，len(authnticatorData)为1byte，len(clientDataJSON)为2byte（Little Endian）
```json
{
  "master_addr":"ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggq7w79h22yxg9h5r3vdw79yhka5vqn48t9yyq080zm49zryzm6pckxh0zjtmw6xqf6n4jj9r9323",
  "backup_addr":"ckt1qqexmutxu0c2jq9q4msy8cc6fh4q7q02xvr7dc347zw3ks3qka0m6qggqajfpetf3wxdye2vm0jnrgm0f5ksrn580qyqweysu45chrxjv4xdhef35dh56tgpe6rhsdqu8hw",
  "msg":"aaa",
  "signature":"40eb33e8e5d852e5cf340c492d115149ab5441034bb40a9db3af82e29f490f6551eecfa3933d95de15edddac2a05aee94936305fef34bdb81a112a7fae52bbc75940bdc4fbe5f27f521445baf9922e068a771280e36e6cc74440f481503f216568f5799587f4f23994b193e78a43a290f0467ec3b53593f6e019674e1d324aaf21d72549960de5880e8c687434170f6476605b8fe4aeb9a28632c7995cf3ba831d976305000000005f007b2274797065223a22776562617574686e2e676574222c226368616c6c656e6765223a2259574668222c226f726967696e223a22687474703a2f2f6c6f63616c686f73743a38303031222c2263726f73734f726967696e223a66616c73657d"
}
```

#### Response
* is_valid: bool, verify result 
```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {
    "is_valid": true
  }
}
```