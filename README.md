# Faultline

**Predictive observability and failure-analysis platform for microservice environments.**

Faultline monitors microservices, calculates explainable service-risk scores, analyzes service dependencies, detects anomalous behaviour, and identifies failure propagation paths before a complete outage occurs.

_See the failure before the failure._

---

## Problem Statement

In microservice architectures, failures cascade. A slow database query in one service can propagate through dependency chains, degrading response times across the entire system until a full outage occurs. Traditional monitoring tools alert _after_ the damage is done.

Faultline takes a different approach: it continuously evaluates service health using deterministic anomaly detection, maps dependency relationships, and calculates risk propagation paths — giving engineers visibility into emerging failures before they become incidents.

---

## Architecture

```
┌────────────────────────────────────────────────────────────┐
│                  React + JSX Dashboard                     │
│  (Overview · Services · Dependencies · Incidents · Sim)    │
└───────────────────────┬────────────────────────────────────┘
                        │ REST API
┌───────────────────────┴────────────────────────────────────┐
│                     Go Backend (Chi)                        │
│                                                            │
│  ┌──────────┐  ┌──────────┐  ┌─────────────┐              │
│  │ Metrics  │  │  Risk    │  │  Incident   │              │
│  │Collector │  │ Engine   │  │  Detector   │              │
│  └────┬─────┘  └────┬─────┘  └──────┬──────┘              │
│       │              │               │                     │
│  ┌────┴──────────────┴───────────────┴──────┐              │
│  │           Dependency Analyzer             │              │
│  │        (Propagation Analysis · BFS)       │              │
│  └───────────────────────────────────────────┘              │
│                                                            │
│  ┌────────────────┐  ┌────────────────────────┐            │
│  │  Simulation    │  │   Simulated Services   │            │
│  │    Engine      │  │  (5 virtual services)  │            │
│  └────────────────┘  └────────────────────────┘            │
└──────────┬───────────────────┬─────────────────────────────┘
           │                   │
    ┌──────┴──────┐    ┌───────┴──────┐
    │ PostgreSQL  │    │    Redis     │
    │ (persistence│    │  (caching,   │
    │  + history) │    │  sim state)  │
    └─────────────┘    └──────────────┘
```

---

## Technology Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.23, Chi router, pgx (PostgreSQL driver), go-redis |
| **Frontend** | React 18, JSX, Vite, Recharts, React Router |
| **Database** | PostgreSQL 16 (or cloud managed PostgreSQL) |
| **Cache** | Redis 7 (or cloud managed Redis, e.g. Upstash) |
| **Infrastructure** | Docker, Docker Compose, Render, Vercel |

---

## Key Features

### Deterministic Risk Scoring
Each service receives an explainable risk score (0-100) computed from:
- **Latency anomaly** — response time deviation from baseline
- **Error rate anomaly** — error rate spikes above historical norm
- **Timeout anomaly** — timeout rate increases
- **Traffic anomaly** — request rate spikes or drops
- **Dependency anomaly** — upstream service health degradation

The overall score is a weighted average: errors (30%), timeouts (20%), latency (20%), dependencies (20%), traffic (10%).

### Risk Levels
| Score | Level | Service Status |
|-------|-------|---------------|
| 0–29 | LOW | Healthy |
| 30–59 | MODERATE | Degraded |
| 60–79 | HIGH | Unhealthy |
| 80–100 | CRITICAL | Critical |

### Dependency Analysis
A directed graph models service dependencies:
```
API Gateway → Auth Service
API Gateway → Order Service → Inventory Service
                            → Payment Service
```

When a service becomes unhealthy, the system determines:
1. **Root cause candidates** — highest-risk upstream ancestor
2. **Affected services** — downstream dependents via BFS traversal
3. **Propagation path** — ordered chain from root cause to affected leaf

### Automatic Incident Management
- Incidents auto-created when risk ≥ 60 (HIGH)
- Incidents escalated on severity changes
- Incidents auto-resolved after 5 consecutive LOW evaluations
- Each incident includes: anomaly breakdown, root cause analysis, propagation path, affected services, and timeline

### Failure Simulator
Inject controlled failures to observe risk propagation:

| Scenario | Target | Effect |
|----------|--------|--------|
| Payment Latency Spike | Payment Service | 5× latency, +40% timeouts |
| Database Slowdown | Payment Service | 8× dep latency, +30% dep errors |
| Auth Failure | Auth Service | +60% errors, +30% timeouts |
| Network Delay | API Gateway | 3× latency, 4× dep latency |
| Traffic Surge | API Gateway | 10× request rate, 2× latency |

---

## Local Development (Docker Compose)

### Prerequisites
- Docker and Docker Compose

### Quick Start
```bash
# Start all 4 containers
docker compose up --build -d

# Check status
docker compose ps

# The dashboard is available at http://localhost:3000
# The API is available at http://localhost:8080/api
# Root health check at http://localhost:8080/healthz
```

---

## Production Deployment

### 1. Backend & Database (Render)

1. **Deploy Managed PostgreSQL**:
   - In Render Dashboard, click **New +** → **PostgreSQL**.
   - Name: `faultline-db`.
   - Plan: **Free**.
   - Copy the **Internal Database URL** (or External if connecting across platforms).

2. **Deploy Go Backend Web Service**:
   - In Render Dashboard, click **New +** → **Web Service**.
   - Connect repository: `https://github.com/RivinmaWethmini/faultline`.
   - Runtime: **Docker**.
   - Root Directory: `backend` (or leave empty if using root `Dockerfile`/`render.yaml`).
   - Dockerfile Path: `./Dockerfile` (relative to `backend` root).
   - Plan: **Free**.
   - **Environment Variables**:
     - `DATABASE_URL`: Your PostgreSQL connection string.
     - `REDIS_URL`: Your Redis connection string (e.g., from Upstash or Render Key-Value).
     - `REDIS_OPTIONAL`: `true` (enables resilient operation if Redis is not configured).
     - `CORS_ALLOWED_ORIGINS`: `https://your-app.vercel.app` (or `*`).
     - `LOG_LEVEL`: `info`.
   - Render automatically provides `PORT`.
   - Your backend URL will be: `https://faultline-backend.onrender.com`.

### 2. Frontend (Vercel)

1. **Import Project into Vercel**:
   - Go to [vercel.com](https://vercel.com) → **Add New Project**.
   - Connect your GitHub repository: `RivinmaWethmini/faultline`.
   - Set **Root Directory**: `frontend`.
   - Framework Preset: **Vite**.
   - Build Command: `npm run build`.
   - Output Directory: `dist`.
2. **Environment Variables**:
   - Set `VITE_API_URL` to your deployed Render backend URL:
     `https://faultline-backend.onrender.com`
3. Click **Deploy**.
   - Your frontend URL will be: `https://faultline.vercel.app`.

---

## License

MIT
