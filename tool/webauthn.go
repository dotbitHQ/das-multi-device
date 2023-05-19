package tool

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"fmt"
	"math/big"
)

func GetPubKey(hash []byte, R, S *big.Int) ([2]*ecdsa.PublicKey, error) {
	var possiblePubkey [2]*ecdsa.PublicKey
	curve := elliptic.P256()

	N := curve.Params().N
	z := new(big.Int).SetBytes(hash)
	//乘法逆元，点P除以1个数s等于点P乘以s的逆元 s的逆元为 s^-1，计算方式为 new(bit.Int).ModInverse(s,N)
	//k^-1计算方式
	sInv := new(big.Int).ModInverse(S, N)
	rInv := new(big.Int).ModInverse(R, N)
	x := R
	//根据x求y
	ySquared := new(big.Int).Exp(x, new(big.Int).SetInt64(3), curve.Params().P)
	ySquared.Sub(ySquared, new(big.Int).Mul(x, big.NewInt(int64(3))))
	ySquared.Add(ySquared, curve.Params().B)
	y := new(big.Int).ModSqrt(ySquared, curve.Params().P)
	if y == nil {
		fmt.Println("err 111")
		return possiblePubkey, nil
	}

	for j := 0; j < 2; j++ {
		if j == 1 {
			y = new(big.Int).Neg(y)
		}
		p := new(ecdsa.PublicKey)
		p.Curve = curve
		p.X = x
		p.Y = y
		//u1 := new(big.Int).Mul(z, rInv)
		//u1.Mod(u1, N)
		u1 := new(ecdsa.PublicKey)
		u1.X, u1.Y = curve.ScalarBaseMult(z.Bytes())
		u1.X, u1.Y = curve.ScalarMult(u1.X, u1.Y, sInv.Bytes())

		////点减去另一个点，等于加上另一个点的x轴对称点
		//p-u1
		u2 := new(ecdsa.PublicKey)
		u1.Y = new(big.Int).Neg(u1.Y)
		u2.X, u2.Y = curve.Add(p.X, p.Y, u1.X, u1.Y)

		Qa := new(ecdsa.PublicKey)
		Qa.Curve = curve
		//Qa = u2 * SR^-1
		tempX, tempY := curve.ScalarMult(u2.X, u2.Y, S.Bytes())
		Qa.X, Qa.Y = curve.ScalarMult(tempX, tempY, rInv.Bytes())
		recoverPubKey := new(ecdsa.PublicKey)
		recoverPubKey.Curve = curve
		recoverPubKey.X = Qa.X
		recoverPubKey.Y = Qa.Y
		fmt.Println("possible x: ", Qa.X)
		fmt.Println("possible y: ", Qa.Y)
		//isValid := ecdsa.Verify(recoverPubKey, hash[:], R, S)
		//fmt.Println(isValid)
		possiblePubkey[j] = recoverPubKey
	}
	return possiblePubkey, nil
}

func GetWebauthnPayload(cid string, pk *ecdsa.PublicKey) (dasLockKey []byte) {
	cid1 := CaculateCid1(cid)
	pk1 := CaculatePk1(pk)

	alg := make([]byte, 1)
	alg[0] = 8

	subAlg := make([]byte, 1)
	subAlg[0] = 7
	dasLockKey = append(dasLockKey, alg...)
	dasLockKey = append(dasLockKey, subAlg...)
	dasLockKey = append(dasLockKey, cid1...)
	dasLockKey = append(dasLockKey, pk1...)
	return
}

//cid' = hash(cid)*5 [:10]
func CaculateCid1(cid string) (cid1 []byte) {
	hash := sha256.Sum256([]byte(cid))
	for i := 0; i < 4; i++ {
		hash = sha256.Sum256(hash[:])
	}
	return hash[22:]
}

//pk' = hash(X+Y)*5 [:10]
func CaculatePk1(pk *ecdsa.PublicKey) (cid1 []byte) {
	xy := append(pk.X.Bytes(), pk.Y.Bytes()...)
	hash := sha256.Sum256([]byte(xy))
	for i := 0; i < 4; i++ {
		hash = sha256.Sum256(hash[:])
	}
	return hash[22:]
}
