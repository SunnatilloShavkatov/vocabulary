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

## Phase 2 - Backend MVP (1 hafta) [Started]

- [x] `auth` login endpoint (bootstrap admin + JWT)
- [x] `vocabulary` create/list/search endpointlar
- [x] Admin create endpoint
- [ ] Basic middleware va logging

Phase 2 next action order:

1. `POST /v1/admins` endpointni implement qilish (hozir `not implemented`)
2. Auth login'ni DB bilan ishlaydigan holatga o'tkazish (bootstrap fallback optional)
3. `vocabulary` write endpointda `created_by` ni JWT claimdan olish
4. Basic request/error logging qo'shish (`gateway` level)
5. Integration test: auth + vocabulary flow (Postgres bilan)

Deliverable:

- API locally testdan o'tadi

## Phase 3 - Admin Web (1 hafta)

- Login page
- Add vocabulary form
- Add admin form
- API integration

Deliverable:

- Admin webdan real ma'lumot qo'shish ishlaydi

## Phase 4 - Client App + Stabilization (1 hafta)

- Welcome page
- Vocabulary list + search
- Error/loading holatlari
- QA, bugfix, release prep

Deliverable:

- End-to-end MVP demo tayyor

