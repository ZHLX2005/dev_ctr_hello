# RSA Key Generation Script (PowerShell)
# Generates RSA key pair for file server authentication

$ErrorActionPreference = "Stop"

$KeyDir = ".\keys"
$PrivateKeyFile = "$KeyDir\private.pem"
$PublicKeyFile = "$KeyDir\public.pem"

# Create key directory
if (-not (Test-Path $KeyDir)) {
    New-Item -ItemType Directory -Path $KeyDir | Out-Null
}

Write-Host "Generating RSA key pair..." -ForegroundColor Green

# Check if OpenSSL is installed
$opensslExists = Get-Command openssl -ErrorAction SilentlyContinue

if (-not $opensslExists) {
    Write-Host "Error: OpenSSL not found." -ForegroundColor Red
    Write-Host "Install OpenSSL or use Git Bash to run generate_keys.sh" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Installation:" -ForegroundColor Cyan
    Write-Host "  1. Windows: Download from https://slproweb.com/products/Win32OpenSSL.html" -ForegroundColor White
    Write-Host "  2. Or use Git Bash to run the .sh script" -ForegroundColor White
    exit 1
}

# Generate private key (2048 bit)
openssl genrsa -out $PrivateKeyFile 2048

if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to generate private key!" -ForegroundColor Red
    exit 1
}

# Extract public key from private key
openssl rsa -in $PrivateKeyFile -pubout -out $PublicKeyFile

if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to extract public key!" -ForegroundColor Red
    exit 1
}

Write-Host "Key generation complete!" -ForegroundColor Green
Write-Host "Private key: $PrivateKeyFile"
Write-Host "Public key: $PublicKeyFile"
Write-Host ""
Write-Host "Important:" -ForegroundColor Yellow
Write-Host "  - Keep the private key safe, client needs it to sign requests"
Write-Host "  - Public key will be deployed to server for verification"
Write-Host "  - DO NOT commit private key to version control"
