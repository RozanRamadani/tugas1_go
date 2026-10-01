$ErrorActionPreference = "SilentlyContinue"

$mhsLogin = curl.exe -s -X POST http://localhost:3000/api/v1/auth/login -H "Content-Type: application/json" -d "@.\scratch\payload_mhs.json"
$mhsToken = ($mhsLogin | ConvertFrom-Json).data.access_token

Write-Host "Enrolling..."
for ($i=1; $i -le 10; $i++) {
    $body = "{`"course_id`": $i, `"tahun_akademik`": `"2026/2027-Ganjil`"}"
    $body | Out-File -Encoding ASCII -FilePath ".\scratch\body_temp.json"
    $res = curl.exe -s -w "\nHTTP Code: %{http_code}\n" -X POST http://localhost:3000/api/v1/enrollments -H "Content-Type: application/json" -H "Authorization: Bearer $mhsToken" -d "@.\scratch\body_temp.json"
    
    Write-Host "Course $i -> $res"
}
