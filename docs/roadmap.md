# Roadmap

## Phase 1 - Foundation (1 hafta) [Done]

- Monorepo papkalarini tayyorlash
- `docs/` kontraktlarini yakunlash
- DB migration skeleton
- Gateway skeleton

Phase 1 checklist:

- [x] `docs/` hujjatlari yaratildi
- [x] `backend/` Go module skeleton yaratildi
- [x] `gateway` server skeleton yaratildi (`/healthz`)
- [x] `migrations/000001_init` yaratildi
- [x] `client/` (Oson Vocabulary) papka skeleton yaratildi
- [x] Docker Compose orqali local test/build verify qilindi
- [x] `gateway/auth/vocabulary/shared` alohida Go modullarga ajratildi (`go.work`)

Deliverable:

- Build bo'ladigan backend skeleton
- Tasdiqlangan API draft

## Phase 2 - Backend MVP (1 hafta) [Done]

- [x] `auth` login endpoint (bootstrap admin + JWT)
- [x] `vocabulary` create/list/search endpointlar
- [x] Admin create endpoint
- [x] Basic middleware va logging

Phase 2 next action order:

1. `POST /v1/admins` endpointni implement qilish (hozir `not implemented`)
2. Auth login'ni DB bilan ishlaydigan holatga o'tkazish (bootstrap fallback optional)
3. `vocabulary` write endpointda `created_by` ni JWT claimdan olish
4. Basic request/error logging qo'shish (`gateway` level)
5. Integration test: auth + vocabulary flow (Postgres bilan)

Deliverable:

- API locally testdan o'tadi

## Phase 3 - Admin Web (1 hafta) [Done]

- [x] Login page
- [x] Add vocabulary form
- [x] Add admin form
- [x] API integration (MVP)

Deliverable:

- Admin webdan real ma'lumot qo'shish ishlaydi

## Phase 4 - Client App + Stabilization (1 hafta) [Started]

- [x] Welcome page
- [x] Vocabulary list + search
- [x] Error/loading holatlari
- [x] CORS middleware (admin web brauzerdan API'ga murojaat uchun)
- [x] Dockerfile multi-stage binary build (go run o'rniga)
- [x] docker-compose.yml `gateway-test` profili tuzatildi
- [x] `.env.example` `CORS_ALLOWED_ORIGINS` bilan yangilandi
- [x] `openapi.yaml` v0.2.0 — barcha endpointlar to'liq schema bilan freeze qilindi
- [x] `oson_vocabulary_mobile` Android/iOS scaffold to'g'irlandi
- [ ] QA: Docker Compose bilan end-to-end smoke test

Deliverable:

- End-to-end MVP demo tayyor

