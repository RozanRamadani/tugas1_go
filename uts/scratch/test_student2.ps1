$ErrorActionPreference = "SilentlyContinue"
$BaseUrl = "http://localhost:3000/api/v1"

$adminBody = @{ email="admin@siakad.com"; password="admin123" } | ConvertTo-Json
$adminToken = (Invoke-RestMethod -Uri "$BaseUrl/auth/login" -Method Post -Body $adminBody -ContentType "application/json").data.access_token

$mhsBody = @{ email="mhs01@student.com"; password="mhs12345" } | ConvertTo-Json
$mhsToken = (Invoke-RestMethod -Uri "$BaseUrl/auth/login" -Method Post -Body $mhsBody -ContentType "application/json").data.access_token

$mhs2Body = @{ email="mhs02@student.com"; password="mhs12345" } | ConvertTo-Json
$mhs2Token = (Invoke-RestMethod -Uri "$BaseUrl/auth/login" -Method Post -Body $mhs2Body -ContentType "application/json").data.access_token

function Test-Get {
    param([string]$Name, [string]$Uri, [string]$Token)
    Write-Host "`n=== Test: $Name ==="
    $headers = @{}
    if ($Token) { $headers["Authorization"] = "Bearer $Token" }
    
    try {
        $res = Invoke-RestMethod -Uri $Uri -Method Get -Headers $headers -ContentType "application/json"
        $res | ConvertTo-Json -Depth 5
    } catch {
        Write-Host "HTTP Code: $($_.Exception.Response.StatusCode.value__)"
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        $errResp = $reader.ReadToEnd()
        Write-Host $errResp
    }
}

Test-Get -Name "Admin akses mhs02 (ID 2)" -Uri "$BaseUrl/students/2" -Token $adminToken
Test-Get -Name "Mhs01 akses diri sendiri (ID 1)" -Uri "$BaseUrl/students/1" -Token $mhsToken
Test-Get -Name "Mhs01 akses mhs02 (ID 2)" -Uri "$BaseUrl/students/2" -Token $mhsToken
Test-Get -Name "ID tidak ada (999)" -Uri "$BaseUrl/students/999" -Token $mhsToken
Test-Get -Name "Mahasiswa soft-delete (ID 3)" -Uri "$BaseUrl/students/3" -Token $mhsToken
Test-Get -Name "Filter valid" -Uri "$BaseUrl/students/1?tahun_akademik=2026/2027-Ganjil" -Token $mhsToken
Test-Get -Name "Filter tidak valid" -Uri "$BaseUrl/students/1?tahun_akademik=invalid" -Token $mhsToken
Test-Get -Name "Tanpa filter" -Uri "$BaseUrl/students/1" -Token $mhsToken
