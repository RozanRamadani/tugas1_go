$ErrorActionPreference = "SilentlyContinue"
$BaseUrl = "http://localhost:3000/api/v1"

$adminBody = @{ email="admin@siakad.com"; password="admin123" } | ConvertTo-Json
$adminToken = (Invoke-RestMethod -Uri "$BaseUrl/auth/login" -Method Post -Body $adminBody -ContentType "application/json").data.access_token

$mhsBody = @{ email="mhs01@student.com"; password="mhs12345" } | ConvertTo-Json
$mhsToken = (Invoke-RestMethod -Uri "$BaseUrl/auth/login" -Method Post -Body $mhsBody -ContentType "application/json").data.access_token

function Test-DeleteCurl {
    param([string]$Name, [string]$Uri, [string]$Token)
    Write-Host "`n=== Test: $Name ==="
    $authHeader = ""
    if ($Token) { $authHeader = "-H `"Authorization: Bearer $Token`"" }
    
    $res = curl.exe -s -w "\nHTTP Code: %{http_code}\n" -X DELETE $Uri $authHeader
    Write-Host $res
}

function Test-GetCurl {
    param([string]$Name, [string]$Uri, [string]$Token)
    Write-Host "`n=== Test: $Name ==="
    $authHeader = ""
    if ($Token) { $authHeader = "-H `"Authorization: Bearer $Token`"" }
    
    $res = curl.exe -s -w "\nHTTP Code: %{http_code}\n" -X GET $Uri $authHeader
    Write-Host $res
}

function Test-PostCurl {
    param([string]$Name, [string]$Uri, [string]$Token, [string]$Body)
    Write-Host "`n=== Test: $Name ==="
    $authHeader = ""
    if ($Token) { $authHeader = "-H `"Authorization: Bearer $Token`"" }
    
    $res = curl.exe -s -w "\nHTTP Code: %{http_code}\n" -X POST $Uri $authHeader -H "Content-Type: application/json" -d $Body
    Write-Host $res
}

# Use ID 6 to test success, as 4 is already deleted
Test-DeleteCurl -Name "1. Admin menghapus mahasiswa yang valid (ID 6)" -Uri "$BaseUrl/students/6" -Token $adminToken
Test-DeleteCurl -Name "2. Admin menghapus ID yang tidak ditemukan (ID 999)" -Uri "$BaseUrl/students/999" -Token $adminToken
Test-DeleteCurl -Name "3. Admin menghapus mahasiswa yang sudah di-soft-delete (ID 6)" -Uri "$BaseUrl/students/6" -Token $adminToken
Test-DeleteCurl -Name "4. Mahasiswa mencoba menghapus data (ID 5)" -Uri "$BaseUrl/students/5" -Token $mhsToken
Test-DeleteCurl -Name "5. Request tanpa token (ID 5)" -Uri "$BaseUrl/students/5" -Token ""
Test-DeleteCurl -Name "6. ID tidak valid (abc)" -Uri "$BaseUrl/students/abc" -Token $adminToken

Test-GetCurl -Name "7. Mhs ID 6 tidak muncul di GET /students/6 (Soft delete check)" -Uri "$BaseUrl/students/6" -Token $adminToken

Write-Host "`n=== Test: 8. Mahasiswa ID 6 melakukan enrollment (Should fail) ==="
$mhs6Body = @{ email="mhs06@student.com"; password="mhs12345" } | ConvertTo-Json
$mhs6Token = (Invoke-RestMethod -Uri "$BaseUrl/auth/login" -Method Post -Body $mhs6Body -ContentType "application/json").data.access_token

$enrollBody = '{"course_id": 1, "tahun_akademik": "2026/2027-Ganjil"}'
Test-PostCurl -Name "Mhs ID 6 Enroll" -Uri "$BaseUrl/enrollments" -Token $mhs6Token -Body $enrollBody
