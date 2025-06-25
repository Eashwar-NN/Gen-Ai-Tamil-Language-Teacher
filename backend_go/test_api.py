import requests
import json
from datetime import datetime
import time

BASE_URL = 'http://localhost:8080/api'

def print_test(name):
    print(f"\n{'='*20} Testing {name} {'='*20}")

def print_response(response):
    print(f"Status: {response.status_code}")
    try:
        print(f"Response: {json.dumps(response.json(), indent=2, ensure_ascii=False)}")
    except:
        print(f"Response: {response.text}")

def test_endpoint(method, endpoint, data=None, params=None, expected_status=200):
    print_test(f"{method} {endpoint}")
    url = f"{BASE_URL}/{endpoint}"
    
    if method == 'GET':
        response = requests.get(url, params=params)
    elif method == 'POST':
        headers = {'Content-Type': 'application/json'}
        response = requests.post(url, json=data, headers=headers)
    elif method == 'PUT':
        response = requests.put(url, json=data)
    elif method == 'DELETE':
        response = requests.delete(url)
        
    print_response(response)
    assert response.status_code == expected_status, f"Expected status {expected_status}, got {response.status_code}"
    return response.json() if response.text else {}

def main():
    # 1. Reset everything
    print("\n=== Starting Fresh ===")
    test_endpoint('POST', 'full_reset')

    # 2. Test health endpoint
    test_endpoint('GET', 'health')

    # 3. Test dashboard endpoints (should be empty initially)
    print("\n=== Testing Dashboard (Empty State) ===")
    test_endpoint('GET', 'dashboard/last_study_session')
    test_endpoint('GET', 'dashboard/study_progress')
    test_endpoint('GET', 'dashboard/quick-stats')

    # 4. Create test data
    print("\n=== Creating Test Data ===")
    
    # Create groups
    groups = [
        {"name": "Basic Greetings"},
        {"name": "Numbers"},
        {"name": "Colors"}
    ]
    created_groups = []
    for group in groups:
        result = test_endpoint('POST', 'groups', data=group, expected_status=201)
        created_groups.append(result)

    # Create words
    words = [
        {
            "tamil": "வணக்கம்",
            "romaji": "vanakkam",
            "english": "hello",
            "parts": {"type": "greeting"}
        },
        {
            "tamil": "நன்றி",
            "romaji": "nandri",
            "english": "thank you",
            "parts": {"type": "greeting"}
        },
        {
            "tamil": "ஒன்று",
            "romaji": "ondru",
            "english": "one",
            "parts": {"type": "number"}
        }
    ]
    created_words = []
    for word in words:
        result = test_endpoint('POST', 'words', data=word, expected_status=201)
        created_words.append(result)

    # Create study activities
    activities = [
        {"group_id": 1, "name": "Vocabulary Quiz"},
        {"group_id": 2, "name": "Number Practice"},
        {"group_id": 3, "name": "Color Recognition"}
    ]
    created_activities = []
    for activity in activities:
        result = test_endpoint('POST', 'study_activities', data=activity, expected_status=201)
        created_activities.append(result)

    # 5. Test GET endpoints with pagination
    print("\n=== Testing GET Endpoints with Pagination ===")
    test_endpoint('GET', 'words', params={"page": 1, "items_per_page": 10})
    test_endpoint('GET', 'groups', params={"page": 1, "items_per_page": 10})
    
    # 6. Test specific item endpoints
    print("\n=== Testing Specific Item Endpoints ===")
    test_endpoint('GET', f'words/1')
    test_endpoint('GET', f'groups/1')
    test_endpoint('GET', f'groups/1/words', params={"page": 1, "items_per_page": 10})
    test_endpoint('GET', f'study_activities/1')

    # 7. Create and test study sessions
    print("\n=== Testing Study Sessions ===")
    
    # Create study session
    session_data = {
        "group_id": 1,
        "study_activity_id": 1
    }
    session = test_endpoint('POST', 'study_sessions', data=session_data, expected_status=201)

    # Test study session endpoints
    test_endpoint('GET', 'study_sessions', params={"page": 1, "items_per_page": 10})
    test_endpoint('GET', f'study_sessions/{session["id"]}')
    test_endpoint('GET', f'study_sessions/{session["id"]}/words', params={"page": 1, "items_per_page": 10})
    test_endpoint('GET', f'groups/1/study_sessions', params={"page": 1, "items_per_page": 10})
    test_endpoint('GET', f'study_activities/1/study_sessions', params={"page": 1, "items_per_page": 10})

    # 8. Test word reviews
    print("\n=== Testing Word Reviews ===")
    
    # Test invalid review (missing correct field)
    test_endpoint('POST', f'study_sessions/{session["id"]}/words/1/review', 
                 data={}, 
                 expected_status=400)

    # Test invalid review (wrong type for correct field)
    test_endpoint('POST', f'study_sessions/{session["id"]}/words/1/review', 
                 data={"correct": "true"}, 
                 expected_status=400)

    # Test valid reviews
    test_endpoint('POST', f'study_sessions/{session["id"]}/words/1/review', 
                 data={"correct": True}, 
                 expected_status=200)

    test_endpoint('POST', f'study_sessions/{session["id"]}/words/2/review', 
                 data={"correct": False}, 
                 expected_status=200)

    test_endpoint('POST', f'study_sessions/{session["id"]}/words/3/review', 
                 data={"correct": True}, 
                 expected_status=200)

    # Test review for non-existent word
    test_endpoint('POST', f'study_sessions/{session["id"]}/words/999/review', 
                 data={"correct": True}, 
                 expected_status=404)

    # Test review for non-existent session
    test_endpoint('POST', f'study_sessions/999/words/1/review', 
                 data={"correct": True}, 
                 expected_status=404)

    # 9. Test dashboard endpoints again (should now have data)
    print("\n=== Testing Dashboard (With Data) ===")
    test_endpoint('GET', 'dashboard/last_study_session')
    test_endpoint('GET', 'dashboard/study_progress')
    test_endpoint('GET', 'dashboard/quick-stats')

    # 10. Test reset endpoints
    print("\n=== Testing Reset Endpoints ===")
    test_endpoint('POST', 'reset_history')
    
    # Verify history is reset but words remain
    test_endpoint('GET', 'dashboard/quick-stats')
    test_endpoint('GET', 'words', params={"page": 1, "items_per_page": 10})

    # Final full reset
    test_endpoint('POST', 'full_reset')
    
    # Verify everything is reset
    test_endpoint('GET', 'words', params={"page": 1, "items_per_page": 10})
    test_endpoint('GET', 'groups', params={"page": 1, "items_per_page": 10})

    print("\n=== All tests completed successfully! ===")

if __name__ == "__main__":
    main() 