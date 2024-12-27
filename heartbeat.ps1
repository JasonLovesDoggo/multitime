$time = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$jsonData = @{
    branch = "main"
    category = "coding"
    cursorpos = 1
    entity = "welcome.txt"
    type = "file"
    lineno = 1
    lines = 1
    project = "welcome"
    time = $time  # Removed quotes to keep it as a number
    user_agent = "wakatime/v1.102.1 (windows) go1.23.4 vscode/1.94.2 vscode-wakatime/24.6.2"
} | ConvertTo-Json -Compress
try {
    $response = Invoke-RestMethod -Uri "http://localhost:3005/api/v1/users/current/heartbeats" `
                                -Method Post `
                                -Headers @{
                                    Authorization = "Bearer Anything"
                                    "Content-Type" = "application/json"
                                } `
                                -Body $jsonData
    Write-Host "$([char]9830) Heartbeat sent successfully"
} catch {
    Write-Host "Error: Heartbeat failed with HTTP status $($_.Exception.Response.StatusCode)"
    return
}
