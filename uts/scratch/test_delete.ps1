$ErrorActionPreference = "SilentlyContinue"

# Logins
$adminToken = (curl.exe -s -X POST http://localhost:3000/api/v1/auth/login -H "Content-Type: application/json" -d "@.\scratch\payload_admin.json" | ConvertFrom-Json).data.access_token
$mhsToken = (curl.exe -s -X POST http://localhost:3000/api/v1/auth/login -H "Content-Type: application/json" -d "@.\scratch\payload_mhs.json" | ConvertFrom-Json).data.access_token

$mhs02Payload = '{"email":"mhs02@student.com","password":"mhs12345"}'
$mhs02Payload | Out-File -Encoding ASCII -FilePath ".\scratch\payload_mhs02.json"
$mhs02Token = (curl.exe -s -X POST http://localhost:3000/api/v1/auth/login -H "Content-Type: application/json" -d "@.\scratch\payload_mhs02.json" | ConvertFrom-Json).data.access_token

function Send-DelReq {
    param([string]$Name, [string]$Uri, [string]$Token)
    Write-Host "`n=== Test: $Name ==="
    $authHeader = ""
    if ($Token -ne "") {
        $authHeader = "-H `"Authorization: Bearer $Token`""
    }
    
    $res = curl.exe -s -i -X DELETE $Uri $authHeader
    Write-Host $res
}

Send-DelReq -Name "Valid DELETE (mhs01 on ID 1)" -Uri "http://localhost:3000/api/v1/enrollments/1" -Token $mhsToken
Send-DelReq -Name "Not Found DELETE (ID 999)" -Uri "http://localhost:3000/api/v1/enrollments/999" -Token $mhsToken
Send-DelReq -Name "Forbidden DELETE (mhs02 on ID 2)" -Uri "http://localhost:3000/api/v1/enrollments/2" -Token $mhs02Token
Send-DelReq -Name "Admin Attempt" -Uri "http://localhost:3000/api/v1/enrollments/2" -Token $adminToken
Send-DelReq -Name "No Token Attempt" -Uri "http://localhost:3000/api/v1/enrollments/2" -Token ""
Send-DelReq -Name "Invalid ID Format" -Uri "http://localhost:3000/api/v1/enrollments/abc" -Token $mhsToken
