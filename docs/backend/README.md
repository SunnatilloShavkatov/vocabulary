# Backend Docs

## Purpose

Backend layer microservice architecture asosida ishlaydi.
Har service alohida run/deploy bo'ladi.

## Service Boundaries

- `api-gateway`: external REST ingress, auth edge, rate-limit, routing.
- `auth-service`: login, JWT issue/refresh, public key distribution.
- `users-service`: profile, settings, role state.
- `dictionary-service`: word CRUD, search, ownership.
- `notification-service`: SRS schedule (`1, 3, 7, 30`), cron worker, FCM dispatch.

## Communication Rules

### External (Admin + Client -> Backend)

- Protocol: HTTP REST only.
- Entry point: `api-gateway`.
- Admin va client ilovalar to'g'ridan-to'g'ri ichki servicega ulanmaydi.

### Internal (Service <-> Service)

- Protocol: gRPC only.
- `api-gateway` ichki service bilan gRPC orqali gaplashadi.
- Service-to-service direct chaqiruvlar ham gRPC orqali bo'ladi.
- Async flowlar uchun RabbitMQ ishlatiladi (`WordAdded` kabi eventlar).

## Data Ownership (DB-per-Service)

Har microservice o'z DBsi (yoki o'z schema namespace) egasi:

- `auth-service-db`: admins, credentials, auth sessions.
- `users-service-db`: users, settings, preferences.
- `dictionary-service-db`: vocabularies, examples, user-word mapping.
- `notification-service-db`: schedules, notification_log, delivery_status.

Rule:

- Service boshqa service DB jadvaliga bevosita query qilmaydi.
- Cross-service data kerak bo'lsa gRPC yoki event consume ishlatiladi.

## Security Model

- Layer 1 (`api-gateway`): JWT signature/expiry validate.
- Layer 2 (microservices): RBAC authorize (`admin`/`user`) business level.
- Gateway validated identity metadata pass qiladi (`x-user-id`, `x-user-role`).

## Runtime Topology

- Har service alohida container/image.
- Har service o'z health/readiness endpointiga ega.
- Monitoring: Prometheus scrape + Grafana dashboards.

## Target API Split

- Public REST surface: `api-gateway`.
- Private RPC surface: gRPC methods (`*.proto` contracts) service ichida.
- Event surface: RabbitMQ exchanges/queues.
