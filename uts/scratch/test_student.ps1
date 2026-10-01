$ErrorActionPreference = "SilentlyContinue"

# Logins
$adminToken = (curl.exe -s -X POST http://localhost:3000/api/v1/auth/login -H "Content-Type: application/json" -d "@.\scratch\payload_admin.json" | ConvertFrom-Json).data.access_token
$mhsToken = (curl.exe -s -X POST http://localhost:3000/api/v1/auth/login -H "Content-Type: application/json" -d "@.\scratch\payload_mhs.json" | ConvertFrom-Json).data.access_token

function Send-GetReq {
    param([string]$Name, [string]$Uri, [string]$Token)
    Write-Host "`n=== Test: $Name ==="
    $authHeader = ""
    if ($Token -ne "") {
        $authHeader = "-H `"Authorization: Bearer $Token`""
    }
    
    $res = curl.exe -s -w "\nHTTP Code: %{http_code}\n" -X GET $Uri $authHeader
    Write-Host $res
}

# Delete Student ID 3 (soft delete) by admin
curl.exe -s -X DELETE http://localhost:3000/api/v1/students/3 -H "Authorization: Bearer $adminToken" | Out-Null

Send-GetReq -Name "1. Admin akses mahasiswa 2" -Uri "http://localhost:3000/api/v1/students/2" -Token $adminToken
Send-GetReq -Name "2. Mhs01 akses diri sendiri (ID 1)" -Uri "http://localhost:3000/api/v1/students/1" -Token $mhsToken
Send-GetReq -Name "3. Mhs01 akses mahasiswa lain (ID 2)" -Uri "http://localhost:3000/api/v1/students/2" -Token $mhsToken
Send-GetReq -Name "4. ID tidak ditemukan (999)" -Uri "http://localhost:3000/api/v1/students/999" -Token $mhsToken
Send-GetReq -Name "5. Mahasiswa soft-delete (ID 3)" -Uri "http://localhost:3000/api/v1/students/3" -Token $mhsToken
Send-GetReq -Name "6. Filter valid" -Uri "http://localhost:3000/api/v1/students/1?tahun_akademik=2026/2027-Ganjil" -Token $mhsToken
Send-GetReq -Name "7. Filter tidak valid" -Uri "http://localhost:3000/api/v1/students/1?tahun_akademik=2026-Ganjil" -Token $mhsToken
Send-GetReq -Name "8. Tanpa filter" -Uri "http://localhost:3000/api/v1/students/1" -Token $mhsToken
