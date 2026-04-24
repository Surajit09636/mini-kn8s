# Mini-K8s Progress Tracker

Use this checklist to track exactly what has been completed in the project and what the next goals are.

## ✅ Phase 1: Completed Architecture

- [x] **Database Foundation**: Set up PostgreSQL and GORM schema (`auth-service`).
- [x] **Authentication Gateway**: Built `/signup` and `/login` with bcrypt hashing.
- [x] **Security Middleware**: Built shared JWT issuing and validation (`pkg/middleware`).
- [x] **Docker Integration**: Connected the Worker Node to the Docker Daemon via Moby SDK.
- [x] **Dynamic Node Discovery**: Workers automatically `POST /register` to the Master on boot.
- [x] **Node Self-Healing**: Workers send `POST /heartbeat` every 10s; Master evicts dead workers after 30s.
- [x] **Workload Scheduling**: Master uses a Thread-Safe Round-Robin Scheduler to distribute containers.
- [x] **Deployment Tracking**: Master now tracks running `Deployments` and their linked `Pods` in its own Database.
- [x] **Status API**: Built `GET /deployments` so users can see their running Docker Container IDs.

---

## 🚀 Phase 2: What to do Next

These are the next major upgrades we plan to build, in order of priority:

- [ ] **Teardown / Deletion API**
  - *Goal:* Give users the ability to stop their running apps.
  - *Task:* Create `DELETE /deployments/{id}` on Master, which tells the Worker to cleanly stop and remove the Docker containers via the SDK.
  
- [ ] **ReplicaSet Controller (Self-Healing)**
  - *Goal:* If a container crashes, the Master notices and restarts it.
  - *Task:* Make Workers include active container IDs in their 10s heartbeat. The Master compares this to the DB. If a container is missing, it triggers a new deployment to replace it.

- [ ] **Kube-Proxy (Load Balancing)**
  - *Goal:* Users get a single IP address to access their web apps, instead of hitting random worker ports.
  - *Task:* Build a Reverse Proxy service that listens on port `80` and routes incoming traffic round-robin to the actual worker containers.

- [ ] **Persistent Storage Volumes (Stretch Goal)**
  - *Goal:* Allow containers (like databases) to save data to the host machine so it survives restarts.
  - *Task:* Map host directories to container directories during the Worker's `ContainerCreate` step.
