# Faultline ⚡

> **See the failure before the failure.**

Faultline is a lightweight predictive observability dashboard for microservices. Instead of waiting for an outage and alerting after things break, Faultline continuously analyzes live service metrics, calculates explainable 0–100 risk scores, maps service dependencies, and tracks failure propagation across your entire stack.

---

## 🚀 Quick Start (Run Locally)

You only need **Docker** installed.

```bash
# 1. Clone the repo and navigate into the folder
git clone https://github.com/RivinmaWethmini/faultline.git
cd faultline

# 2. Start all services in the background
docker compose up -d
```

Once running:
- **Web Dashboard**: [http://localhost:3000](http://localhost:3000)
- **Backend API**: [http://localhost:8080/api](http://localhost:8080/api)
- **Health Check**: [http://localhost:8080/healthz](http://localhost:8080/healthz)

To shut down everything:
```bash
docker compose down
```

---

## 🕹️ How to Use the Dashboard

Faultline comes preloaded with 5 simulated microservices: `API Gateway`, `Auth Service`, `Order Service`, `Inventory Service`, and `Payment Service`.

### 1. Overview (`/`)
- Shows overall system health (`Healthy`, `Degraded`, or `Critical`).
- Displays active incident counts and highlights the service with the highest risk.
- Shows live risk scores and status badges for all monitored microservices.

### 2. Services (`/services`)
- Click any service to view its detailed breakdown.
- **Risk Gauge**: A 0–100 score indicating operational health:
  - `0–29`: **LOW** (Healthy)
  - `30–59`: **MODERATE** (Slight degradation)
  - `60–79`: **HIGH** (Unhealthy / Incident threshold)
  - `80–100`: **CRITICAL** (Imminent failure)
- **Explainable Factors**: Clear, plain-English reasons for the score (e.g., *"Response latency is 3.5x above baseline"*).
- **Live Charts**: Real-time graphs for response time, error rate, request volume, and historical risk.

### 3. Dependency Map (`/dependencies`)
- Visual graph showing how your microservices call each other:
  ```
  API Gateway ──► Auth Service
  API Gateway ──► Order Service ──┬──► Inventory Service
                                 └──► Payment Service
  ```
- Nodes change colors dynamically as services degrade, showing upstream impact.

### 4. Incidents (`/incidents`)
- Incidents are automatically opened whenever a service's risk score crosses **60 (HIGH)**.
- Shows the **Root Cause** candidate, **Propagation Path**, affected downstream services, and a timestamped event timeline.
- Incidents auto-resolve once service metrics return to healthy levels.

### 5. Failure Simulator (`/simulator`)
Simulate real-world outages and watch Faultline detect them in real-time:
1. Pick a scenario (e.g., **Payment Latency Spike** or **Database Slowdown**).
2. Click **Trigger Scenario**.
3. Watch the risk score climb, dependencies turn orange/red, and an incident automatically generate.
4. Click **Stop & Revert** to return the system to normal.

---

## 🛠️ Tech Stack

- **Backend**: Go (Chi router, PostgreSQL driver, Redis client, goroutines)
- **Frontend**: React 18 (JSX, Vite, Recharts, Vanilla CSS dark mode)
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Containerization**: Docker & Docker Compose

---

## 📄 License

MIT
