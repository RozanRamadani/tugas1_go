$ErrorActionPreference = "Stop"
$BaseUrl = "http://localhost:3000/api/v1"

# Login
$adminBody = @{ email="admin@siakad.com"; password="admin123" } | ConvertTo-Json
$adminToken = (Invoke-RestMethod -Uri "$BaseUrl/auth/login" -Method Post -Body $adminBody -ContentType "application/json").data.access_token

$mhsBody = @{ email="mhs01@student.com"; password="mhs12345" } | ConvertTo-Json
$mhsToken = (Invoke-RestMethod -Uri "$BaseUrl/auth/login" -Method Post -Body $mhsBody -ContentType "application/json").data.access_token

function Test-Put {
    param([string]$Name, [string]$Uri, [string]$Token, [hashtable]$BodyHash)
    Write-Host "`n=== Test: $Name ==="
    $headers = @{}
    if ($Token) { $headers["Authorization"] = "Bearer $Token" }
    
    $bodyJson = ""
    if ($BodyHash) {
        $bodyJson = $BodyHash | ConvertTo-Json
    }

    try {
        $res = Invoke-RestMethod -Uri $Uri -Method Put -Headers $headers -Body $bodyJson -ContentType "application/json"
        $res | ConvertTo-Json -Depth 5
    } catch {
        Write-Host "HTTP Code: $($_.Exception.Response.StatusCode.value__)"
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        $errResp = $reader.ReadToEnd()
        Write-Host $errResp
    }
}

Test-Put -Name "1. Admin update mhs02 (valid)" -Uri "$BaseUrl/students/2" -Token $adminToken -BodyHash @{ nim="NIM_BARU_YANG_HARUSNYA_DIABAIKAN"; nama="Siti Aminah Updated"; prodi="Ilmu Komputer"; angkatan=2024; ipk_terakhir=3.99 }
Test-Put -Name "2. Admin update ID tidak ada (999)" -Uri "$BaseUrl/students/999" -Token $adminToken -BodyHash @{ nama="Ghost"; prodi="Sistem Informasi"; angkatan=2023; ipk_terakhir=3.5 }
Test-Put -Name "3. Admin update tidak valid (prodi kosong)" -Uri "$BaseUrl/students/2" -Token $adminToken -BodyHash @{ nama="Siti Aminah"; angkatan=2024; ipk_terakhir=3.99 }
Test-Put -Name "4. Mahasiswa update data" -Uri "$BaseUrl/students/1" -Token $mhsToken -BodyHash @{ nama="Hack"; prodi="Sistem Informasi"; angkatan=2023; ipk_terakhir=4.0 }
Test-Put -Name "5. Tanpa Token" -Uri "$BaseUrl/students/1" -Token "" -BodyHash @{ nama="Anonymous"; prodi="Sistem Informasi"; angkatan=2023; ipk_terakhir=3.5 }

# Verify NIM remains the same
Write-Host "`n=== Test: 6. Verify NIM (Should not change) ==="
$verifyRes = Invoke-RestMethod -Uri "$BaseUrl/students/2" -Method Get -Headers @{Authorization="Bearer $adminToken"} -ContentType "application/json"
Write-Host "NIM saat ini: $($verifyRes.data.nim) (Harusnya 11223302, bukan NIM_BARU_YANG_HARUSNYA_DIABAIKAN)"
