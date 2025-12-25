# RSA 密钥对生成脚本 (PowerShell)
# 用于生成文件服务器的公钥和私钥

$ErrorActionPreference = "Stop"

$KeyDir = ".\keys"
$PrivateKeyFile = "$KeyDir\private.pem"
$PublicKeyFile = "$KeyDir\public.pem"

# 创建密钥目录
if (-not (Test-Path $KeyDir)) {
    New-Item -ItemType Directory -Path $KeyDir | Out-Null
}

Write-Host "正在生成 RSA 密钥对..." -ForegroundColor Green

# 检查是否安装了 OpenSSL
$opensslExists = Get-Command openssl -ErrorAction SilentlyContinue

if (-not $opensslExists) {
    Write-Host "错误: 未找到 OpenSSL。" -ForegroundColor Red
    Write-Host "请安装 OpenSSL 或使用 Git Bash 运行 generate_keys.sh" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "安装方法:" -ForegroundColor Cyan
    Write-Host "  1. Windows: 下载并安装 from https://slproweb.com/products/Win32OpenSSL.html" -ForegroundColor White
    Write-Host "  2. 或使用 Git for Windows 附带的 Git Bash 运行 .sh 脚本" -ForegroundColor White
    exit 1
}

# 生成私钥 (2048 位)
openssl genrsa -out $PrivateKeyFile 2048

if ($LASTEXITCODE -ne 0) {
    Write-Host "生成私钥失败!" -ForegroundColor Red
    exit 1
}

# 从私钥提取公钥
openssl rsa -in $PrivateKeyFile -pubout -out $PublicKeyFile

if ($LASTEXITCODE -ne 0) {
    Write-Host "提取公钥失败!" -ForegroundColor Red
    exit 1
}

Write-Host "密钥生成完成!" -ForegroundColor Green
Write-Host "私钥: $PrivateKeyFile"
Write-Host "公钥: $PublicKeyFile"
Write-Host ""
Write-Host "重要提示:" -ForegroundColor Yellow
Write-Host "  - 请妥善保管私钥，客户端需要使用私钥对请求进行签名"
Write-Host "  - 公钥将被部署到服务器用于验证签名"
Write-Host "  - 私钥文件请不要提交到版本控制系统"
