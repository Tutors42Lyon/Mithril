---
layout: default
title: Api Specification 
parent: Architecture
nav_order: 6
---

# Mithril API Specification

## Authentication

### 42 Login

Web

```
GET /api/auth/login/42

Query Params (optional):
  ?redirect_uri=/dashboard  (where to go after login)

Response (302 Found):
Location: https://api.intra.42.fr/oauth/authorize?
  client_id=YOUR_APP_ID&
  redirect_uri=http://localhost:3000/api/auth/callback&
  response_type=code&
  scope=public
```

TUI

```
GET /api/auth/login/42?cli=true

Response (200 OK):
{
  "auth_url": "https://api.intra.42.fr/oauth/authorize?...",
  "device_code": "abc123",
  "message": "Open this URL in browser: https://..."
}

```

## Exercise Pools
