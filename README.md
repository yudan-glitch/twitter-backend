# Twitter Backend Clone (Go + TDD)

Hi! This is a project I am building completely from scratch to teach myself backend web development. My goal is to move step by step from a basic localhost setup to a real, deployable system that can handle real world traffic. 

Instead of rushing to build tons of features, I want to learn how to write clean code. I am using Test-Driven Development (TDD) for every single endpoint, which means I write my tests first before writing any actual application logic.

## My Roadmap

I am taking things slow and building this system incrementally.

### Step 1: The Basics (Where I am now)
Right now I am still defining and implementing basic endpoints. The server runs on my local machine and uses a temporary in-memory mock store filled with fake data. I have written table-driven tests to check for successful lookups, unknown users, and bad text inputs. 

### Step 2: Adding a Database
My next goal will be to connect a real database like PostgreSQL. Because I used interfaces for Step 1, I should be able to plug in a real database without breaking my existing web router code.

### Step 3: Making it Twitter
Once the database is working, I will start adding actual Twitter features. This means writing code and tests for creating tweets, deleting tweets, following other users, building a basic timeline feed, and so on.

### Step 4: Security, Caching, and Cloud
The final steps will be learning about the things needed for a real launch. I want to add user authentication (like logins and secure tokens), protect the site from spam using rate limiters, speed things up with an in-memory cache, deploy the backend to the cloud and much more.

---

## How to Run It Locally

### Prerequisites
You just need Go 1.22+ installed on your computer.

### Running the Test Suite
To run my automated test suite and see the TDD cycles in action, run this in your terminal:
```bash
go test ./...
```

### Starting the Server
For now, to spin up the local server with my temporary seed data, run:
```bash
go run ./cmd/api/main.go
```
The server will start running on http://localhost:8080.

## About Me
I am a backend developer in training. This repository is my personal journal for tracking my progress. I am starting with a single endpoint, but I am excited to see how this project grows as I discover more about advanced backend.
