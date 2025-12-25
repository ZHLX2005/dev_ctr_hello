// Package auth 提供基于 RSA 签名的身份验证功能
package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sync"
)

var (
	// ErrInvalidSignature 签名无效错误
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrInvalidPublicKey 公钥无效错误
	ErrInvalidPublicKey = errors.New("invalid public key")
	// ErrSignatureNotFound 签名未找到错误
	ErrSignatureNotFound = errors.New("signature not found in request")
)

// SignatureValidator 签名验证器
type SignatureValidator struct {
	publicKey *rsa.PublicKey
	mu        sync.RWMutex
}

// NewSignatureValidator 创建签名验证器
func NewSignatureValidator(publicKeyPath string) (*SignatureValidator, error) {
	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &SignatureValidator{
		publicKey: publicKey,
	}, nil
}

// NewSignatureValidatorFromBytes 从字节创建签名验证器
func NewSignatureValidatorFromBytes(pubKeyBytes []byte) (*SignatureValidator, error) {
	publicKey, err := parsePublicKey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &SignatureValidator{
		publicKey: publicKey,
	}, nil
}

// Verify 验证签名
// data: 原始数据
// signature: Base64 编码的签名
func (v *SignatureValidator) Verify(data []byte, signature string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if v.publicKey == nil {
		return ErrInvalidPublicKey
	}

	// 解码 Base64 签名
	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// 计算数据的 SHA256 哈希
	hashed := sha256.Sum256(data)

	// 使用 RSA 公钥验证签名
	err = rsa.VerifyPKCS1v15(v.publicKey, crypto.SHA256, hashed[:], sigBytes)
	if err != nil {
		return ErrInvalidSignature
	}

	return nil
}

// VerifyString 验证字符串签名
func (v *SignatureValidator) VerifyString(data string, signature string) error {
	return v.Verify([]byte(data), signature)
}

// ReloadPublicKey 重新加载公钥（支持热更新）
func (v *SignatureValidator) ReloadPublicKey(publicKeyPath string) error {
	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return err
	}

	v.mu.Lock()
	v.publicKey = publicKey
	v.mu.Unlock()

	return nil
}

// loadPublicKey 从文件加载公钥
func loadPublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	keyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	return parsePublicKey(keyBytes)
}

// parsePublicKey 解析公钥
func parsePublicKey(keyBytes []byte) (*rsa.PublicKey, error) {
	// 解码 PEM 块
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, ErrInvalidPublicKey
	}

	// 解析公钥
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, ErrInvalidPublicKey
	}

	return rsaPub, nil
}

// Sign 使用私钥对数据进行签名（供客户端使用）
func Sign(data []byte, privateKeyPath string) (string, error) {
	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return "", err
	}

	return SignWithKey(data, privateKey)
}

// SignWithKey 使用私钥对象对数据进行签名
func SignWithKey(data []byte, privateKey *rsa.PrivateKey) (string, error) {
	// 计算数据的 SHA256 哈希
	hashed := sha256.Sum256(data)

	// 使用 RSA 私钥签名
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}

	// 返回 Base64 编码的签名
	return base64.StdEncoding.EncodeToString(signature), nil
}

// SignString 对字符串进行签名（供客户端使用）
func SignString(data string, privateKeyPath string) (string, error) {
	return Sign([]byte(data), privateKeyPath)
}

// loadPrivateKey 从文件加载私钥
func loadPrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// 解码 PEM 块
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("failed to decode private key PEM block")
	}

	// 解析私钥
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// 尝试 PKCS8 格式
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key is not RSA type")
		}
		return rsaKey, nil
	}

	return privateKey, nil
}
