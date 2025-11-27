---
layout: default
title: Use case 
parent: Architecture
nav_order: 7
---

## MVP Feature 1: Login

```
User opens Mithril
  ↓
Clicks "Login with 42"
  ↓
Redirected to 42 Intra
  ↓
User logs in
  ↓
Redirected back to Mithril
  ↓
Session created
  ↓
User sees menu 
```

## MVP Feature 2: Browse Exercises
```
User clicks "Exercises"
  ↓
See list of exercise pools:
  - C Basics (5 exercises)
  - Algorithms (8 exercises)
  ↓
Click pool
  ↓
See exercises in order:
  1. Hello World 
  2. Arrays
  3. Loops
```

## MVP Feature 3: Solve Exercise
```
User clicks exercise
  ↓
See problem statement
  ↓
Option A (TUI):
  - Press "Edit"
  - Opens vim/nano with temp file
  - User codes
  - Saves and closes
  ↓
Click "Submit"
  ↓
Code goes to grading service
  ↓
Grading service tests code
  ↓
Results: PASS or FAIL
  ↓
User sees: Results + Feedback
```
