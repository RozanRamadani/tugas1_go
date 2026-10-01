$ErrorActionPreference = "Stop"

# 1. Login Admin
$adminLogin = Invoke-RestMethod -Uri "http://localhost:3000/api/v1/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body '{"email":"admin@siakad.com","password":"admin123"}'
$adminToken = $adminLogin.data.access_token
Write-Host "Admin Token: $adminToken"

# 2. Login Mahasiswa
$mhsLogin = Invoke-RestMethod -Uri "http://localhost:3000/api/v1/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body '{"email":"mhs01@student.com","password":"mhs12345"}'
$mhsToken = $mhsLogin.data.access_token
Write-Host "Mahasiswa Token: $mhsToken"

function Send-Req {
    param(
        [string]$Name,
        [string]$Uri,
        [string]$Token,
        [string]$Body
    )
    Write-Host "`n=== Test: $Name ==="
    $headers = @{}
    $headers["Content-Type"] = "application/json"
    if ($Token) {
        $headers["Authorization"] = "Bearer $Token"
    }

    try {
        if ($Body) {
            $resp = Invoke-WebRequest -Uri $Uri -Method POST -Headers $headers -Body $Body -UseBasicParsing
        } else {
            $resp = Invoke-WebRequest -Uri $Uri -Method POST -Headers $headers -UseBasicParsing
        }
        Write-Host "HTTP/1.1 $($resp.StatusCode) $($resp.StatusDescription)"
        Write-Host $resp.Content
    } catch {
        $ex = $_.Exception.Response
        if ($ex) {
            $statusCode = [int]$ex.StatusCode
            $statusDesc = $ex.StatusDescription
            $stream = $ex.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $content = $reader.ReadToEnd()
            Write-Host "HTTP/1.1 $statusCode $statusDesc"
            Write-Host $content
        } else {
            Write-Host "Error: $_"
        }
    }
}

Send-Req -Name "5. Valid Enrollment" -Uri "http://localhost:3000/api/v1/enrollments" -Token $mhsToken -Body '{"course_id": 1, "tahun_akademik": "2026/2027-Ganjil"}'
Send-Req -Name "6. Tanpa Auth (No Token)" -Uri "http://localhost:3000/api/v1/enrollments" -Token "" -Body '{"course_id": 1, "tahun_akademik": "2026/2027-Ganjil"}'
Send-Req -Name "7. Menggunakan Akun Admin (Role Check)" -Uri "http://localhost:3000/api/v1/enrollments" -Token $adminToken -Body '{"course_id": 2, "tahun_akademik": "2026/2027-Ganjil"}'
Send-Req -Name "8. Duplicate Enrollment" -Uri "http://localhost:3000/api/v1/enrollments" -Token $mhsToken -Body '{"course_id": 1, "tahun_akademik": "2026/2027-Ganjil"}'
Send-Req -Name "9. Invalid Format Tahun Akademik" -Uri "http://localhost:3000/api/v1/enrollments" -Token $mhsToken -Body '{"course_id": 2, "tahun_akademik": "2026-Ganjil"}'
Send-Req -Name "10. Course Not Found" -Uri "http://localhost:3000/api/v1/enrollments" -Token $mhsToken -Body '{"course_id": 999, "tahun_akademik": "2026/2027-Ganjil"}'
