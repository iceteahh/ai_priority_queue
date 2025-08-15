readme.so logo

light
Download
SectionsReset

Delete
Click on a section below to edit the contents

Click on a section below to add it to your readme
Search for a section

Custom Section

Acknowledgements

API Reference

Appendix

Authors

Badges

Color Reference

Contributing

Demo

Deployment

Documentation

Environment Variables

FAQ

Features

Feedback

Github Profile - About Me

Github Profile - Introduction

Github Profile - Links

Github Profile - Other

Github Profile - Skills

Installation

Lessons

License

Logo

Optimizations

Related

Roadmap

Run Locally

Screenshots

Support

Tech

Running Tests

Usage/Examples

Used By
Editor
significantly faster targeted operations.

### 2. Anti-Starvation Mechanism
I designed an **anti-starvation policy** that dynamically adjusts selection based on wait time vs. priority.  
- **Why:** Without this, lower-priority ads could wait indefinitely if higher-priority ads keep arriving.  
- **Approach:** We use a formula combining base priority and waited time to determine the next dequeue candidate.


Preview
Raw
Project Title
A brief description of what this project does and who it's for

Video Processing Priority Queue
A high-performance, priority-based queue implementation in Go for scheduling and processing video ads.
Supports multiple priorities, FIFO ordering within priorities, anti-starvation, and concurrent dequeue operations.

Features
Priority-based processing — Higher priority ads are processed first.
FIFO within priority levels — Items with the same priority are processed in arrival order.
Anti-starvation mechanism — Older low-priority ads can be promoted or processed to prevent indefinite waiting.
Maximum wait time enforcement — Ads can define a MaxWaitTime in seconds; they’ll be processed once they exceed it.
Game family index — Fast reprioritization and filtering by GameFamily.
Time index — Quick lookups of ads based on enqueue time.
Concurrent processing — Designed to work with multiple workers.
Metrics — Get distribution of ads by priority.
AI Agent Interface Supports natural language interface that can interpret and execute queue management commands
Project Structure
AI_PRIORITY_QUEUE/
├── ai_agents/ # Python-based agents
│ ├── .venv/ # Python virtual environment
│ └── queue_agent/ # ai agent
│
├── priority_queue/ # Go-based priority queue service
│ ├── .vscode/ # VS Code workspace settings
│ ├── client_sample/ # Sample clients for testing the API
│ │ ├── dequeue_client/ # Example client for dequeuing ads
│ │ └── enqueue_client/ # Example client for enqueuing ads
│ ├── cmd/server/ # Application entrypoint
│ │ └── server.go # Main server startup code
│ ├── config/ # Configuration files
│ │ ├── config.go # Config loader/structs
│ │ └── config.yaml # Example config file
│ ├── internal/ # Internal packages
│ │ ├── ads/ # Ad model definitions
│ │ ├── httpapi/ # HTTP API handlers and routing
│ │ └── queue/ # Core priority queue logic
│ ├── api_key/ # API key handling (if applicable)
│ ├── go.mod # Go module definition
│ ├── go.sum # Go module checksums
│ ├── priority-queue.postman_collection.json # Postman collection for API testing
│ ├── sample_api_calls # Sample API request scripts
│ └── README.md # Project documentation
Architecture Overview
Project Structure

Quick Start
Requirements
Go 1.24+
(Optional) Google BTree for indexing
Python 3
Google ADK
Installation
git clone https://github.com/iceteahh/ai_priority_queue.git

Setup and run priority_queue httpserver
Setup

cd ai_priority_queue/priority_queue
go mod tidy
Config file: config/config.yaml

totalPriority: 3
enableAntiStarvation: true
maximumWaitSeconds: 600   # seconds
btreeDegree: 16
Run the queue server

go run cmd/server/server.go
Run the sample enqueue_client

go run enqueue_client/enqueue_client.go -rate=5 -total=500000
rate: ads/second
total: total number of ads
Run the sample dequeue_client

go run dequeue_client/dequeue_client.go -workers=3 -rate=4
workers: total number of workers
rate: ads/second
Setup and run queue_agent
cd ai_priority_queue/ai_agents
source .venv/bin/activate
Run queue_agent

adk run queue_agent
Interfaces
Queue Http Server APIs
If you use postman use this file to import the collection

priority_queue/priority-queue.postman_collection.json
Enpoints
Method	Path	Description
GET	/healthz	Health check
POST	/enqueue	Add an ad to the queue
POST	/dequeue	Remove and return the next ad
GET	/peek?n={n}	View the next n ads without removing
GET	/distribution	Get priority distribution & anti-starvation flag
GET	/waiting?age={duration}	List ads waiting longer than a given age
POST	/reprioritize/family	Change priority for all ads in a game family
POST	/reprioritize/age	Change priority for all ads older than a given age
POST	/settings/antiStarvation	Enable/disable anti-starvation
POST	/settings/maximumWait	Set global maximum wait time (seconds)
Examples
/enqueue

Request

curl --location 'http://localhost:8080/enqueue' \
--header 'Content-Type: application/json' \
--data '{
  "ad": {
    "adId": "ad_103",
    "title": "Dragon",
    "gameFamily": "RPG",
    "targetAudience": [
      "18-34"
    ],
    "priority": 1,
    "createdAt": "2025-01-01T00:00:00Z",
    "maxWaitTime": 120
  }
}'
Response

{
    "adId": "ad_103",
    "title": "Dragon",
    "gameFamily": "RPG",
    "targetAudience": [
        "18-34"
    ],
    "priority": 1,
    "createdAt": "2025-01-01T00:00:00Z",
    "maxWaitTime": 120
}
/dequeue

Request

curl --location --request POST 'http://localhost:8080/dequeue'
Response

{
    "adId": "ad_101",
    "title": "Dragon",
    "gameFamily": "RPG",
    "targetAudience": [
        "18-34"
    ],
    "priority": 2,
    "createdAt": "2025-01-01T00:00:00Z",
    "maxWaitTime": 120
}
/peek?n={n}

Request

curl --location 'http://localhost:8080/peek?n=5'
Response

[
  {
    "adId": "ad_101",
    "title": "Dragon",
    "gameFamily": "RPG",
    "targetAudience": [
      "18-34"
    ],
    "priority": 2,
    "createdAt": "2025-01-01T00:00:00Z",
    "maxWaitTime": 120
  },
  {
    "adId": "ad_102",
    "title": "Dragon",
    "gameFamily": "RPG",
    "targetAudience": [
      "18-34"
    ],
    "priority": 2,
    "createdAt": "2025-01-01T00:00:00Z",
    "maxWaitTime": 120
  },
  {
    "adId": "ad_103",
    "title": "Dragon",
    "gameFamily": "RPG",
    "targetAudience": [
      "18-34"
    ],
    "priority": 1,
    "createdAt": "2025-01-01T00:00:00Z",
    "maxWaitTime": 120
  }
]
/distribution

Request

curl --location 'http://localhost:8080/distribution'
Response

{
    "total": 6,
    "distribution": [
        {
            "Priority": 3,
            "Count": 0,
            "Percent": 0
        },
        {
            "Priority": 2,
            "Count": 6,
            "Percent": 100
        },
        {
            "Priority": 1,
            "Count": 0,
            "Percent": 0
        }
    ],
    "enable_anti_starvation": true
}
/waiting?age={duration}

Request

curl --location 'http://localhost:8080/waiting?age=5s'
Response

[
  {
    "adId": "ad_101",
    "title": "Dragon",
    "gameFamily": "RPG",
    "targetAudience": [
      "18-34"
    ],
    "priority": 2,
    "createdAt": "2025-01-01T00:00:00Z",
    "maxWaitTime": 120
  },
  {
    "adId": "ad_102",
    "title": "Dragon",
    "gameFamily": "RPG",
    "targetAudience": [
      "18-34"
    ],
    "priority": 2,
    "createdAt": "2025-01-01T00:00:00Z",
    "maxWaitTime": 120
  },
  {
    "adId": "ad_103",
    "title": "Dragon",
    "gameFamily": "RPG",
    "targetAudience": [
      "18-34"
    ],
    "priority": 1,
    "createdAt": "2025-01-01T00:00:00Z",
    "maxWaitTime": 120
  }
]
reprioritize/family

Request

curl --location 'http://localhost:8080/reprioritize/family' \
--header 'Content-Type: application/json' \
--data '{
  "family": "RPG",
  "newPriority": 3
}'
Response

{
    "ok": true
}
/reprioritize/age

Request

curl --location 'http://localhost:8080/reprioritize/age' \
--header 'Content-Type: application/json' \
--data '{
  "age": "30s",
  "newPriority": 2
}'
Response

{
    "ok": true
}
settings/antiStarvation

Request

curl --location 'http://localhost:8080/settings/antiStarvation' \
--header 'Content-Type: application/json' \
--data '{
  "enable": false
}'
Response

{
    "ok": true
}
settings/maximumWait

Request

curl --location 'http://localhost:8080/settings/maximumWait' \
--header 'Content-Type: application/json' \
--data '{
  "maximumWait": 120
}'
Response

{
    "ok": true
}
Queue Agent
Commands

Change priority to 5 for all ads in the RPG-Fantasy family
Set priority to 1 for ads older than 10 minutes
Enable starvation mode" (ignore anti-starvation rules)
Set maximum wait time to 600 seconds
Show the next 5 ads to be processed
List all ads waiting longer than 5 minutes
What's the current queue distribution by priority?
Design Decisions
1. Queue Structure
I implemented a multi-queue model (one queue per priority level) with additional indexing by game family and enqueue time. And I use doubly linked list to implement the queue

Why: This allows constant-time insertion/removal while still supporting fast lookups for operations like reprioritization or “list ads waiting longer than X minutes.”
Trade-off: Slightly higher memory usage due to maintaining multiple indices, but significantly faster targeted operations.
2. Anti-Starvation Mechanism
I designed an anti-starvation policy that dynamically adjusts selection based on wait time vs. priority.

Why: Without this, lower-priority ads could wait indefinitely if higher-priority ads keep arriving.
Approach: We use a formula combining base priority and waited time to determine the next dequeue candidate.
3. Time Index with B-Tree
A B-tree is used for time-based indexing of ads in the queue.

Why: B-trees allow efficient range queries (e.g., “all ads older than 5 minutes”) and ordered traversal without scanning all queues.
Challenge Solved: Fast lookups for SLA enforcement and bulk operations like “reprioritize all RPG ads older than X minutes.”
4. Stable Reprioritization
When reprioritizing ads (by family or age), we maintain their relative order based on enqueue time.

Why: Preserves fairness and prevents “queue jumping.”
Approach: Items are reinserted into the target priority queue in ascending enqueue time order.
readme.so