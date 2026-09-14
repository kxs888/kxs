package sm

import (
	"bytes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	intcrypto "github.com/kxs888/kxs/backend/internal/crypto"
	"github.com/kxs888/kxs/backend/internal/errcode"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/gmsm/sm4"
)

// Envelope 是可选第 5 层的请求/响应包装。密钥只来自环境变量。
type Envelope struct {
	IV         string `json:"iv"`
	Ciphertext string `json:"ciphertext"`
	MAC        string `json:"mac"`
}

func ParseKeyHex(hexKey string) ([]byte, error) {
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, errcode.New(errcode.SMCryptoInvalid, 500, "invalid SM4 key")
	}
	if len(b) != sm4.BlockSize {
		return nil, errcode.New(errcode.SMCryptoInvalid, 500, "SM4 key must be 16 bytes")
	}
	return b, nil
}

func Encrypt(key, plaintext []byte) (*Envelope, error) {
	iv, err := intcrypto.RandomBytes(sm4.BlockSize)
	if err != nil {
		return nil, err
	}
	ct, err := sm4CBCEncrypt(key, iv, plaintext)
	if err != nil {
		return nil, err
	}
	return &Envelope{
		IV:         hex.EncodeToString(iv),
		Ciphertext: hex.EncodeToString(ct),
		MAC:        macHex(key, iv, ct),
	}, nil
}

func Decrypt(key []byte, env *Envelope) ([]byte, error) {
	if env == nil || env.IV == "" || env.Ciphertext == "" {
		return nil, errcode.New(errcode.SMCryptoInvalid, 400, "missing sm envelope")
	}
	iv, err := hex.DecodeString(env.IV)
	if err != nil || len(iv) != sm4.BlockSize {
		return nil, errcode.New(errcode.SMCryptoInvalid, 400, "invalid iv")
	}
	ct, err := hex.DecodeString(env.Ciphertext)
	if err != nil {
		return nil, errcode.New(errcode.SMCryptoInvalid, 400, "invalid ciphertext")
	}
	if env.MAC == "" || env.MAC != macHex(key, iv, ct) {
		return nil, errcode.New(errcode.SMCryptoInvalid, 400, "mac mismatch")
	}
	pt, err := sm4CBCDecrypt(key, iv, ct)
	if err != nil {
		return nil, errcode.New(errcode.SMCryptoInvalid, 400, "decrypt failed")
	}
	return pt, nil
}

func macHex(key, iv, ct []byte) string {
	h := sm3.New()
	_, _ = h.Write(iv)
	_, _ = h.Write(ct)
	_, _ = h.Write(key)
	return hex.EncodeToString(h.Sum(nil))
}

func sm4CBCEncrypt(key, iv, plaintext []byte) ([]byte, error) {
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plain := pkcs7Pad(plaintext, block.BlockSize())
	out := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, plain)
	return out, nil
}

func sm4CBCDecrypt(key, iv, ciphertext []byte) ([]byte, error) {
	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext not full blocks")
	}
	out := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, ciphertext)
	return pkcs7Unpad(out, block.BlockSize())
}

func pkcs7Pad(b []byte, block int) []byte {
	n := block - (len(b) % block)
	return append(b, bytes.Repeat([]byte{byte(n)}, n)...)
}

func pkcs7Unpad(b []byte, block int) ([]byte, error) {
	if len(b) == 0 || len(b)%block != 0 {
		return nil, errors.New("invalid padding")
	}
	n := int(b[len(b)-1])
	if n == 0 || n > block || n > len(b) {
		return nil, errors.New("invalid padding")
	}
	for i := 0; i < n; i++ {
		if b[len(b)-1-i] != byte(n) {
			return nil, errors.New("invalid padding")
		}
	}
	return b[:len(b)-n], nil
}

func MarshalEnvelope(env *Envelope) ([]byte, error) {
	return json.Marshal(env)
}

func UnmarshalEnvelope(b []byte) (*Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, errcode.New(errcode.SMCryptoInvalid, 400, "sm envelope must be json")
	}
	return &env, nil
}

func MustKey(hexKey string) []byte {
	k, err := ParseKeyHex(hexKey)
	if err != nil {
		panic(fmt.Sprintf("sm4 key: %v", err))
	}
	return k
}
