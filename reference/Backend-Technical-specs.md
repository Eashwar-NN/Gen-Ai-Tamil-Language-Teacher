# Backend Server Technical Specs

## Business Goal:

A language learning school wants to build a prototype of learning portal which will act as three things:
- Inventory of possible vocabulary that can be learned
- Act as a  Learning record store (LRS), providing correct and wrong score on practice vocabulary
- A unified launchpad to launch different learning apps

## Technical Requirements

- The backend will be built using Go
- The database will be SQLite3
- The API will be built using gin
- The API will always return JSON
- There will be no authentication or authorization.
- Everything will be treated as single user.

## Database Schema

Our database will be a single sqlite database called `words.db` that will be in the root of the project folder of `backend_go`

We have the following tables:
- words - stores vocabulary words
    - id (int)
    - Tamil (str)
    - romaji (str)
    - english (str)
    - parts (json)

- words_groups - join table for words and groups many to many.
    - id (int)
    - word_id (int)
    - group_id (int)
    
- groups - thematic groups of words
    - id (int)
    - name (str)

- study_sessions - records of study sessions grouping word_review_items
    - id (int)
    - group_id (int)
    - created_at (datetime)
    - study_activities_id (int)

- study_activities - a specific study activity, linking a study session to a group
    - id (int)
    - study_session_id (int)
    - group_id (int)
    - created_at (datetime)
    
- word_review_items - a record of word practice, determining if the word was correct or not
    - word_id (int)
    - study_session_id (int)
    - correct (bool)
    - created_at (datetime)

## API Endpoints


### GET /api/dashboard/last_study_session
Returns information about the most recent study session.

#### JSON Response 
```json
{
    "id": 123,
    "group_id": 456,
    "created_at": "2024-03-20T15:30:00Z",
    "study_activities_id": 789,
    "group_id": 456,
    "group_name": "Basic Greetings"
}
```

### GET /api/dashboard/study_progress
Returns study progress statistics over time.
Please note that the frontend will determine progress bar based on total words studied and total words available.

#### JSON Response 
```json
{
    "daily_stats": [
        {
            "date": "2024-03-20",
            "total_words_studied": 50,
            "total_available_words": 124,
            "correct_percentage": 85,

        }
    ],
    "total_words_studied": 500,
    "total_available_words": 1024,
    "average_accuracy": 82.5
}
```

### GET /api/dashboard/quick-stats
Returns quick overview learning statistics.

#### JSON Response
```json
{
    "total_words": 1000,
    "words_studied": 300,
    "total_study_sessions": 25,
    "average_accuracy": 75.5,
    "study_streak_days": 5
}
```

### GET /api/study_activities/:id
Returns details of a specific study activity.

#### JSON Response
```json
{
    "id": 123,
    "name": "Vocabulary Quiz",
    "thumbnail_url": "https://example.com/thumbnail.jpg",
    "description": "Practise your vocabulary with flashcards",
}
```

### GET /api/study_activities/:id/study_sessions
Returns all study sessions for a specific activity.
Pagination with 100 items per page.
#### JSON Response
```json
{
    "items": [
        {
            "id": 456,
            "activity_name": "Vocabulary Quiz",
            "group_name": "Basic Greetings",
            "start_time": "2025-02-08T17:20:23-05:00",
            "end_time": "2025-02-08T17:30:23-05:00",
            "review_items_count": 20
        }
    ],
    "pagination": {
        "current_page": 1,
        "total_pages": 5,
        "total_items": 100,
        "items_per_page": 20
    }
}
```

### POST /api/study_activities
Creates a new study activity.
Required params: group_id (int), study_activity_id (int)

Request body:
```json
{
    "group_id": 1,
    "study_activity_id": 123
}
```
#### JSON Response:
```json
{
    "id": 789,
    "group_id": 456,
}
```

### GET /api/words
Returns paginated list of words.
Pagination of 100 items per page.

#### JSON Response
```json
{
    "items": [
        {
            "tamil": "வணக்கம்",
            "romaji": "vanakkam",
            "english": "hello",
            "correct_count": 5,
            "wrong_count": 2
            }
        }
    ],
    "pagination":{
        "current page":1,
        "total_pages": 10,
        "total_items": 1000,
        "items_per_page": 100
    }
}
```

### GET /api/words/:id
Returns details of a specific word.

#### JSON Response:
```json
{
    "tamil": "வணக்கம்",
    "romaji": "vanakkam",
    "english": "hello",
    "stats": {
        "correct_count": 5,
        "wrong_count": 1
    },
    "groups": [
        {
            "id": 1,
            "name": "Basic Greetings"
        }
    ]
}
```

### GET /api/groups
Returns paginated list of groups.

#### JSON Response
```json
{
    "items": [
        {
            "id": 1,
            "name": "Basic Greetings",
            "word_count": 20
        }
    ],
    "pagination":{
        "current_page": 1,
        "total_pages": 5,
        "total_items": 500,
        "items_per_page": 100
    }
}
```

### GET /api/groups/:id
Returns details of a specific group.

#### JSON Response
```json
{
    "id": 1,
    "name": "Basic Greetings",
    "stats": {
        "total_word_count": 20,
    }
}
```

#### GET /api/groups/:id/words
Returns all words in a specific group.

#### JSON Response
```json
{
    "items": [
        {
            "tamil": "வணக்கம்",
            "romaji": "vanakkam",
            "english": "hello",
            "correct_count": 5,
            "wrong_count": 2
        
        }
    ],
    "pagination":{
        "current_page": 1,
        "total_pages": 5,
        "total_items": 500,
        "items_per_page": 100
    }
}
```

### GET /api/groups/:id/study_sessions
Returns all study sessions for a specific group.

#### JSON Response
```json
{
    "items":[
        {
            "id":123,
            "activity_name": "Vocabulary Quiz",
            "group_name": "Basic Greetings",
            "start_time": "2025-02-08T17:20:23-05:00",
            "end_time": "2025-02-08T17:30:23-05:00",
            "review_items_count":20            
        }
    ],
    "pagination":{
        "current_page": 1,
        "total_pages": 1,
        "total_items": 5,
        "items_per_page": 100
    }
}
```


### GET /api/study_sessions
Returns paginated list of study sessions.

#### JSON Response
```json
{
    "items": [
        {
            "id": 123,
            "activity_name": "Vocabulary Quiz",
            "group_name": "Basic Greetings",
            "start_time": "2025-02-08T17:20:23-05:00",
            "end_time": "2025-02-08T17:30:23-05:00",
            "review_items_count":20    
        }
    ],
    "pagination":{
        "current_page": 1,
        "total_pages": 1,
        "total_items": 5,
        "items_per_page": 100
    }
}
```

### GET /api/study_sessions/:id
Returns details of a specific study session.
#### JSON Response
```json
{
    "id": 123,
    "activity_name": "Vocabulary Quiz",
    "group_name": "Basic Greetings",
    "start_time": "2025-02-08T17:20:23-05:00",
    "end_time": "2025-02-08T17:30:23-05:00",
    "review_items_count":20   
    }
```

### GET /api/study_sessions/:id/words
Returns all words reviewed in a specific study session.
Pagination with 100 items per page.
#### JSON Response
```json
{
    "items":[
        "tamil": "வணக்கம்",
        "romaji": "vanakkam",
        "english": "hello",
        "correct_count": 5,
        "wrong_count":2
    ],
    "pagination":{
        "current_page": 1,
        "total_pages": 1,
        "total_items": 5,
        "items_per_page": 100
    }
}
```

### POST /api/reset_history
Resets all study history.

#### JSON Response
```json
{
    "success": true,
    "message": "Study history has been reset",
    "reset_timestamp": "2024-03-20T15:30:00Z"
}
```

### POST /api/full_reset
Resets all data including words and groups.

#### JSON Response
```json
{
    "success": true,
    "message": "System has been fully reset",
    "reset_timestamp": "2024-03-20T15:30:00Z"
}
```

### POST /api/study_sessions/:id/words/:word_id/review
Records a word review result.

#### Required params
- id (study_session_id) (int)
- word_id (int)
- correct (bool)

#### Request Payload
```json
{
    "correct": true
}
```

#### JSON Response:
```json
{
    "success": true,
    "word_id": 1,
    "session_id": 123,
    "correct": true,
    "created_at": "2024-03-20T15:30:00Z"
}
```
## Mage Tasks

Mage is a task runner for Go.
Lets list out the possible tasks we need for our lang portal.

### Initialize Database
This will initialize the SQLite database called `words.db`

### Migrate Database
This will run a series of migration sql files on the database

Migration live in the `migration` folder.
The migration files will run in the order of their file name.
The file names should look like this:

```sql
0001_init.sql
0002_create_words_table.sql
```

### Seed Data
This task will import json files and transform them into target data for our database.

All seed files live in the `seeds` folder.
All seed files should be loaded.

In our task, we should have a DSL to specify each seed file and its expected group word name.
```json
[
    {
        "tamil": "வணக்கம்",
        "romaji": "vanakkam",
        "english": "hello",
    },
]

```

