$time = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()

# Sample heartbeats data (replace with your actual data)
$heartbeats = @(
  @{
    branch = "main"
    category = "coding"
    cursorpos = 1
    entity = "testing.py"
    type = "file"
    lineno = 1
    lines = 1
    project = "welcome"
    time = $time
  },
  @{
    branch = "dev"
    category = "debugging"
    entity = "app.js"
    type = "file"
    lineno = 10
    lines = 50
    project = "my-app"
    time = $time - 30
  }
)

# Convert each heartbeat object to JSON and combine them into a single JSON array
$jsonData = ($heartbeats | ConvertTo-Json -Compress) -join ','
$jsonData = "[$($jsonData)]"  # Wrap the array in square brackets

try {
  $response = Invoke-RestMethod -Uri "http://localhost:3005/users/current/heartbeats.bulk" `
                              -Method Post `
                              -Headers @{
                                Authorization = "Bearer Anything"
                                "Content-Type" = "application/json"
                              } `
                              -Body $jsonData
  Write-Host "$([char]9830) Heartbeats sent successfully"
} catch {
  Write-Host "Error: Heartbeats failed with HTTP status $($_.Exception.Response.StatusCode)"
  return
}
