#!/bin/bash
# RSA 密钥对生成脚本
# 用于生成文件服务器的公钥和私钥

set -e

KEY_DIR="./keys"
PRIVATE_KEY_FILE="${KEY_DIR}/private.pem"
PUBLIC_KEY_FILE="${KEY_DIR}/public.pem"

# 创建密钥目录
mkdir -p "${KEY_DIR}"

echo "正在生成 RSA 密钥对..."

# 生成私钥 (2048 位)
openssl genrsa -out "${PRIVATE_KEY_FILE}" 2048

# 从私钥提取公钥
openssl rsa -in "${PRIVATE_KEY_FILE}" -pubout -out "${PUBLIC_KEY_FILE}"

echo "密钥生成完成!"
echo "私钥: ${PRIVATE_KEY_FILE}"
echo "公钥: ${PUBLIC_KEY_FILE}"
echo ""
echo "重要提示:"
echo "  - 请妥善保管私钥，客户端需要使用私钥对请求进行签名"
echo "  - 公钥将被部署到服务器用于验证签名"
echo "  - 私钥文件请不要提交到版本控制系统"
