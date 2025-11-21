---
layout: default
title: Overview
parent: Architecture
nav_order: 1
---
# Architecture Overview

## System Architecture
```mermaid
graph TB
    %% Styling
    classDef client fill:#e1f5ff,stroke:#01579b,stroke-width:2px
    classDef gateway fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef broker fill:#fff9c4,stroke:#f57f17,stroke-width:3px
    classDef service fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px
    classDef worker fill:#ffe0b2,stroke:#e65100,stroke-width:2px
    classDef storage fill:#fce4ec,stroke:#880e4f,stroke-width:2px

    subgraph CLIENT["Client Layer"]
        TUI[Terminal UI<br/>Go + Bubbletea]:::client
        WEB[Web Dashboard<br/>Mithril.js]:::client
    end
    
    subgraph GATEWAY["API Gateway"]
        GW[REST API<br/>Node.js]:::gateway
        WS[WebSocket]:::gateway
    end
    
    subgraph BROKER["Message Broker"]
        NATS[NATS Server]:::broker
    end
    
    subgraph SERVICES["Core Services"]
        direction LR
        AUTH[Auth]:::service
        USER[User]:::service
        EXERCISE[Exercise]:::service
        GRADING[Grading]:::service
        STATS[Stats]:::service
    end
    
    subgraph POOL["Grading Pool"]
        direction LR
        W1[Worker]:::worker
        W2[Worker]:::worker
        W3[Worker]:::worker
        W4[Worker]:::worker
    end
    
    subgraph DATA["Data Layer"]
        DB[(PostgreSQL)]:::storage
        REDIS[(Redis)]:::storage
        FILES[(YAML Files)]:::storage
    end
    
    %% Connections
    TUI --> GATEWAY
    WEB --> GATEWAY
    GW --> NATS
    WS --> NATS
    
    NATS <--> SERVICES
    
    GRADING --> POOL
    
    SERVICES --> DB
    EXERCISE --> FILES
    SERVICES --> REDIS
```

## User Flow
```mermaid
flowchart LR
    Start([Student starts Mithril])
    Login[Login with 42 credentials]
    Menu[Main Menu]
    Select[Select Exercise Level]
    Exercise[Get Random Exercise]
    Code[Write Code in Editor]
    Test[Run Tests Locally]
    Submit[Submit 'grademe']
    Result{Pass?}
    Next[Next Exercise]
    Stats[Update Statistics]
    End([Session Complete])
    
    Start --> Login
    Login --> Menu
    Menu --> Select
    Select --> Exercise
    Exercise --> Code
    Code --> Test
    Test --> Submit
    Submit --> Result
    Result -->|Yes| Stats
    Result -->|No| Code
    Stats --> Next
    Next --> Exercise
    Menu -->|Quit| End
```

## Data Flow
```mermaid
sequenceDiagram
    participant U as User/TUI
    participant API as API Gateway
    participant EX as Exam Engine
    participant C as Compilation Service
    participant D as Docker
    participant DB as Database
    
    U->>API: Submit code (grademe)
    API->>EX: Process submission
    EX->>C: Queue compilation job
    C->>D: Create container
    D->>D: Compile code
    D->>D: Run test cases
    D-->>C: Return results
    C-->>EX: Test results
    EX->>DB: Save attempt
    EX-->>API: Format response
    API-->>U: Display results + hints
    
    Note over U: Real-time via WebSocket
```
