# Roadmap

## Current Status (BRD Alignment)

Maqsad: BRD talablariga o'tish.

- External integratsiya: Admin/Client -> REST (Gateway) [in progress]
- Internal integratsiya: Service-to-service gRPC [not started]
- Data ownership: DB-per-service [not started]
- Async event flow: RabbitMQ [partial]

## Phase A - Foundation + MVP (Done)

- [x] Monorepo va Go workspace (`go.work`) tayyor
- [x] Gateway + Auth + Vocabulary bazaviy endpointlar ishlaydi
- [x] Admin login va admin create ishlaydi
- [x] Vocabulary create/list/search ishlaydi
- [x] Basic logging, CORS, Docker build/composition tayyor
- [x] Admin va Client ilovalarda MVP UI oqimlari bor

Deliverable:

- Ishlaydigan MVP asos

## Phase B - BRD Skeleton (Done)

- [x] `users` service skeleton qo'shildi
- [x] `notification` service skeleton qo'shildi
- [x] Gateway hybrid identity forwarding (`X-User-ID`, `X-User-Role`) qo'shildi
- [x] `/metrics` endpoint qo'shildi
- [x] RabbitMQ + Prometheus + Grafana compose stack qo'shildi
- [x] `users` va `schedules` migratsiyasi qo'shildi
- [x] Docs 3 bo'limga ajratildi: `docs/admin`, `docs/backend`, `docs/client`

Deliverable:

- BRDga mos dastlabki karkas

## Phase C - BRD Core Migration (Next Priority)

### C1. Service Runtime Split

- [ ] Har service uchun alohida `cmd/server`
- [ ] Har service uchun alohida Docker image/service (`auth`, `users`, `dictionary`, `notification`, `gateway`)
- [ ] Compose/K8s manifestlarda service discovery va healthchecks

### C2. gRPC Internal Contracts

- [ ] `proto/` papka va versiyalangan proto contractlar
- [ ] Gateway -> services gRPC clients
- [ ] Service-to-service gRPC chaqiriqlar (REST ichki yo'lni yopish)
- [ ] gRPC auth metadata va RBAC interceptorlar

### C3. DB-per-Service

- [ ] `auth-db`, `users-db`, `dictionary-db`, `notification-db` ajratish
- [ ] Har service migratsiyasini alohida pipeline qilish
- [ ] Cross-service direct SQL ni to'liq taqiqlash

### C4. Event-driven Completion

- [ ] `dictionary-service` dan `WordAdded` publisher
- [ ] `notification-service` consumer + idempotency
- [ ] Cron + FCM sender to'liq ishlatish

Deliverable:

- Haqiqiy microservice runtime (separate run + gRPC + DB-per-service)

## Phase D - Validation + Release Readiness

- [ ] End-to-end smoke test (admin -> dictionary -> event -> notification)
- [ ] Contract tests (REST + gRPC)
- [ ] Observability dashboardlar (latency, error rate, queue lag)
- [ ] Security hardening (key rotation, rate-limit tuning, secrets policy)

Deliverable:

- BRDga to'liq mos release candidate

## Immediate Start Order (Confirmed)

1. Birinchi: `auth-service`ni alohida runtime/container qilish.
2. Keyin: `proto` contractlar + Gateway gRPC integration.
3. So'ng: `dictionary-service` split + `WordAdded` publisher.
4. Parallel: client/admin integratsiya ishlarini ulash.

## Execution Breakdown (Current Sprint)

### Track 1 - Auth First

- [x] `backend/modules/auth/cmd/server` entrypoint ochish
- [x] `auth-service` uchun alohida Docker service qo'shish
- [x] Health/readiness endpointlarni ajratish
- [x] Auth-specific metrics endpoint qo'shish (`/metrics`)

### Track 2 - gRPC Gateway Integration

- [x] `proto/auth/v1` contract yozish
- [x] gRPC codegen pipeline qo'shish
- [x] Gateway -> auth gRPC client ulash
- [x] Gateway REST `login/create-admin` oqimini auth gRPC ga o'tkazish

### Track 3 - Dictionary Split + Event

- [ ] `backend/modules/vocabulary/cmd/server` entrypoint ochish
- [ ] `dictionary-service` alohida containerga chiqarish
- [ ] `WordAdded` RabbitMQ publisherni production oqimga ulash

### Track 4 - Client/Admin Parallel Integration

- [ ] Client: Login + secure token flow (`access`/`refresh`) + vocabulary CRUD + FCM flow
- [ ] Admin: user CRUD + word moderation + Grafana WebView monitoring
- [ ] REST contractlar BRD/OpenAPI bilan sinxronlashtirish

### Track 5 - Client Mobile BRD Alignment (New)

- [ ] Gateway/Auth: `refresh` va `logout` endpointlarini production-ready qilish
- [ ] Gateway: access/refresh token validation policy va expiry contractini yakunlash
- [ ] Client: Dio interceptor bilan `401 -> refresh -> retry` oqimini implement qilish
- [ ] Client: token/name ni `flutter_secure_storage` ga yozish
- [ ] Client: Home list + realtime search + empty-state UXni BRDga moslash
- [ ] Client: FAB add-word (`word/translation/definition`) oqimini yakunlash
- [ ] Client: FCM token register + reminder deep-link handling
- [ ] Auth policy: self-register o'rniga admin-created user oqimini hujjatlashtirish

Deliverable:

- Mobile client BRD bo'yicha to'liq ishlaydigan auth + home + add-word + reminder oqimi

## Next Action (Client BRD)

1. Gatewayda `POST /v1/auth/refresh` va `POST /v1/auth/logout` endpointlarini qo'shish.
2. Client networking qatlamida Dio refresh interceptor yozish.
3. Secure storage service (`access`, `refresh`, `name`)ni ulash.
4. Home/FAB/Empty-state va FCM flowni sprint tasklarga bo'lish.

## Track 6 - Admin Dashboard BRD Alignment (New)

- [x] Users Service: admin user CRUD + block/unblock endpointlarini boshlash (skeleton)
- [x] Dictionary Service: global moderation CRUD + advanced search endpointlarini yakunlash
- [x] Gateway: admin-only route group va RBAC policy hardening (users/admin routes)
- [x] Admin backend: dashboard stats endpointi (`users_count`, `words_count`) DB-backed
- [x] Admin backend: backup/export endpointi (`.sql`/`.json`) DB-backed basic export
- [x] Admin backend: backup export async job + status + download URL lifecycle
- [x] Admin backend: audit log write/read endpointlari
- [ ] Admin panel: user management UI (`create/update/delete/block`)
- [x] Admin panel: user management integration start (`list/block/unblock`)
- [x] Admin panel: dictionary moderation integration start (`list/search/approve/reject`)
- [ ] Admin panel: backup/export action UI
- [ ] Admin panel: monitoring (Grafana WebView)

Deliverable:

- Admin dashboard BRD bo'yicha to'liq nazorat paneli (user/word/backup/monitoring)

## Next Action (Admin BRD)

1. Admin panelda backup export jobs UI (`start/poll/download`) ni ulash.
2. Audit log UI (`filter by actor/action`) ni ulash.
3. Audit logni barcha admin vocabulary endpointlariga ham to'liq avtomatik yozish.

