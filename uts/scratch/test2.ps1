$ErrorActionPreference = "SilentlyContinue"

# 1. Login Admin
$adminLogin = curl.exe -s -X POST http://localhost:3000/api/v1/auth/login -H "Content-Type: application/json" -d "@.\scratch\payload_admin.json"
$adminToken = ($adminLogin | ConvertFrom-Json).data.access_token
Write-Host "Admin Token: $adminToken"

# 2. Login Mahasiswa
$mhsLogin = curl.exe -s -X POST http://localhost:3000/api/v1/auth/login -H "Content-Type: application/json" -d "@.\scratch\payload_mhs.json"
$mhsToken = ($mhsLogin | ConvertFrom-Json).data.access_token
Write-Host "Mahasiswa Token: $mhsToken"

function Send-Req {
    param(
        [string]$Name,
        [string]$Uri,
        [string]$Token,
        [string]$Body
    )
    Write-Host "`n=== Test: $Name ==="
    $authHeader = ""
    if ($Token -ne "") {
        $authHeader = "-H `"Authorization: Bearer $Token`""
    }
    
    $BodyFile = ".\scratch\body_temp.json"
    $Body | Out-File -Encoding utf8 -FilePath $BodyFile
    
    $cmd = "curl.exe -s -i -X POST $Uri -H `"Content-Type: application/json`" $authHeader -d `"@$BodyFile`""
    $output = Invoke-Expression $cmd
    $output -join "`n" | Write-Host
}

Send-Req -Name "5. Valid Enrollment" -Uri "http://localhost:3000/api/v1/enrollments" -Token $mhsToken -Body '{"course_id": 1, "tahun_akademik": "2026/2027-Ganjil"}'
Send-Req -Name "6. Tanpa Auth (No Token)" -Uri "http://localhost:3000/api/v1/enrollments" -Token "" -Body '{"course_id": 1, "tahun_akademik": "2026/2027-Ganjil"}'
Send-Req -Name "7. Menggunakan Akun Admin (Role Check)" -Uri "http://localhost:3000/api/v1/enrollments" -Token $adminToken -Body '{"course_id": 2, "tahun_akademik": "2026/2027-Ganjil"}'
Send-Req -Name "8. Duplicate Enrollment" -Uri "http://localhost:3000/api/v1/enrollments" -Token $mhsToken -Body '{"course_id": 1, "tahun_akademik": "2026/2027-Ganjil"}'
Send-Req -Name "9. Invalid Format Tahun Akademik" -Uri "http://localhost:3000/api/v1/enrollments" -Token $mhsToken -Body '{"course_id": 2, "tahun_akademik": "2026-Ganjil"}'
Send-Req -Name "10. Course Not Found" -Uri "http://localhost:3000/api/v1/enrollments" -Token $mhsToken -Body '{"course_id": 999, "tahun_akademik": "2026/2027-Ganjil"}'
