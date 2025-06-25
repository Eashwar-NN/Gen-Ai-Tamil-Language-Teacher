# Function to make HTTP requests and format output
function Test-Endpoint {
    param (
        [string]$Method,
        [string]$Endpoint,
        [string]$Body = "",
        [string]$Description = "",
        [hashtable]$Query = @{}
    )
    Write-Host "`n=== Testing $Method $Endpoint ==="
    if ($Description) {
        Write-Host "Description: $Description"
    }
    
    # Build query string
    $uri = "http://localhost:8080$Endpoint"
    if ($Query.Count -gt 0) {
        $queryParams = @()
        foreach ($key in $Query.Keys) {
            $queryParams += "$key=$($Query[$key])"
        }
        $uri = "$uri`?$($queryParams -join '&')"
    }
    
    Write-Host "URI: $uri"
    if ($Body) {
        Write-Host "Body: $Body"
    }
    
    $params = @{
        Method = $Method
        Uri = $uri
        ContentType = 'application/json; charset=utf-8'
    }
    
    if ($Body -ne "") {
        $params.Body = $Body
    }
    
    try {
        $response = Invoke-WebRequest @params
        Write-Host "Status: $($response.StatusCode)"
        Write-Host "Response: $($response.Content)"
        return $response.Content | ConvertFrom-Json
    } catch {
        Write-Host "Error: $($_.Exception.Response.StatusCode.value__)"
        Write-Host "Message: $($_.Exception.Message)"
        return $null
    }
}

Write-Host "`n=== Step 1: Reset Database ==="
Test-Endpoint "POST" "/api/full_reset" -Description "Resetting database"

# Wait for reset to complete
Start-Sleep -Seconds 1

Write-Host "`n=== Step 2: Create Test Data ==="
# Create a test group
$groupBody = '{"name": "Basic Greetings"}'
$group = Test-Endpoint "POST" "/api/groups" $groupBody -Description "Creating test group"

# Create test word
$wordBody = @"
{
    "tamil": "வணக்கம்",
    "romaji": "vanakkam",
    "english": "hello",
    "parts": {"type": "greeting"}
}
"@
$word = Test-Endpoint "POST" "/api/words" $wordBody -Description "Creating test word"

# Create test study activity
$activityBody = '{"group_id": 1, "name": "Vocabulary Quiz"}'
$activity = Test-Endpoint "POST" "/api/study_activities" $activityBody -Description "Creating study activity"

Write-Host "`n=== Step 3: Testing GET Endpoints ==="
# Test health endpoint
Test-Endpoint "GET" "/api/health" -Description "Health check"

# Test dashboard endpoints
Test-Endpoint "GET" "/api/dashboard/last_study_session" -Description "Get last study session"
Test-Endpoint "GET" "/api/dashboard/study_progress" -Description "Get study progress"
Test-Endpoint "GET" "/api/dashboard/quick-stats" -Description "Get quick stats"

# Test words endpoints with pagination
Test-Endpoint "GET" "/api/words" -Description "Get all words" -Query @{
    page = 1
    items_per_page = 10
}
Test-Endpoint "GET" "/api/words/1" -Description "Get specific word"

# Test groups endpoints with pagination
Test-Endpoint "GET" "/api/groups" -Description "Get all groups" -Query @{
    page = 1
    items_per_page = 10
}
Test-Endpoint "GET" "/api/groups/1" -Description "Get specific group"
Test-Endpoint "GET" "/api/groups/1/words" -Description "Get group words" -Query @{
    page = 1
    items_per_page = 10
}
Test-Endpoint "GET" "/api/groups/1/study_sessions" -Description "Get group study sessions" -Query @{
    page = 1
    items_per_page = 10
}

# Test study activities endpoints
Test-Endpoint "GET" "/api/study_activities/1" -Description "Get specific study activity"
Test-Endpoint "GET" "/api/study_activities/1/study_sessions" -Description "Get study activity sessions" -Query @{
    page = 1
    items_per_page = 10
}

# Test study sessions endpoints with pagination
Test-Endpoint "GET" "/api/study_sessions" -Description "Get all study sessions" -Query @{
    page = 1
    items_per_page = 10
}

Write-Host "`n=== Step 4: Testing POST Endpoints ==="
# Create a study session
$sessionBody = '{"group_id": 1, "study_activity_id": 1}'
$session = Test-Endpoint "POST" "/api/study_sessions" $sessionBody -Description "Creating study session"

# Test word review
$reviewBody = '{"correct": true}'
Test-Endpoint "POST" "/api/study_sessions/1/words/1/review" $reviewBody -Description "Submit word review"

# Test reset endpoints
Test-Endpoint "POST" "/api/reset_history" -Description "Reset study history" 