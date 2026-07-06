# Backend Foundation (BRD Alignment)

Bu papka `Dictionary Platform` loyihasining Go backend qismini saqlaydi.

## Tuzilma

- `go.work` - barcha backend modullarni birlashtiruvchi workspace
- `gateway/` - alohida Go module, API kirish nuqtasi
- `modules/auth/` - alohida Go module (`controller` + `service`)
- `modules/auth/cmd/server/` - auth-service alohida runtime entrypoint
- `proto/` - gRPC contractlar (`auth/v1`)
- `modules/users/` - alohida Go module (`controller` + `service`)
- `modules/vocabulary/` - alohida Go module (`controller` + `service` + `repository`)
- `modules/notification/` - alohida Go module (`controller` + `service`)
- `libs/shared/` - alohida Go module, umumiy config/logger contractlar
- `migrations/` - SQL migration fayllari
- `monitoring/` - Prometheus scrape config

## Hozirgi holat

- Har service alohida Go module sifatida ajratilgan
- Transport va business logic ajratildi (`controller` qatlam qo'shildi)
- Har service o'z endpointlarini o'zi register qiladi
- `gateway` faqat composition point sifatida barcha routerlarni birlashtiradi
- `GET /healthz` endpoint ishlaydi
- `auth-service` alohida `GET /healthz` bilan ishlaydi (`:8081`)
- `auth-service` `GET /readyz` va `GET /metrics` endpointlari mavjud
- `auth-service` gRPC endpointi `:9091` da ishlaydi (`auth.v1.AuthService`)
- `gateway` auth REST endpointlari (`/v1/auth/login`, `/v1/admins`) auth-service gRPC orqali ishlaydi
- `GET /metrics` endpoint ishlaydi (Prometheus format)
- `POST /v1/auth/login` bootstrap admin bilan ishlaydi
- `POST /v1/vocabulary` JWT bilan himoyalangan va DB ga yozadi
- `GET /v1/vocabulary` public list/search endpoint ishlaydi
- `POST /v1/admins` ishlaydi
- `GET /v1/users/me` protected profile contract mavjud
- `POST /internal/notifications/word-added` SRS (`1,3,7,30`) schedule generator mavjud
- `admins`, `vocabularies`, `users`, `schedules` migrationlari mavjud
- Docker Compose bilan local run/test tayyor

## Docker bilan ishlatish (tavsiya)

Tez yo'l (script):

```bash
cd /Users/sshovkatov/Projects/dictionary/backend
./scripts/local.sh up
./scripts/local.sh status
./scripts/local.sh logs
./scripts/local.sh down
```

1) `backend/.env.docker.example` nusxa olib `.env` qiling (xohlasangiz o'zgartiring):

```bash
cd /Users/sshovkatov/Projects/dictionary/backend
cp .env.docker.example .env
```

2) Testlarni container ichida ishga tushiring:

```bash
cd /Users/sshovkatov/Projects/dictionary/backend
docker compose --profile test run --rm gateway-test
```

3) Postgres + Redis + RabbitMQ + migrate + gateway + auth-service stackni ko'taring:

```bash
cd /Users/sshovkatov/Projects/dictionary/backend
docker compose up -d postgres redis rabbitmq migrate gateway auth-service prometheus grafana
```

4) Proto code generation (Track 2):

```bash
cd /Users/sshovkatov/Projects/dictionary/backend
./scripts/gen-proto.sh
```

4) Health check:

```bash
curl http://localhost:8080/healthz
```

5) Local bootstrap admin login sinovi:

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@osonvocabulary.local","password":"admin12345"}'
```

Default local credentials:

- `email`: `admin@osonvocabulary.local`
- `password`: `admin12345`

6) Ish tugagach to'xtatish:

```bash
cd /Users/sshovkatov/Projects/dictionary/backend
docker compose down
```

## Native (ixtiyoriy) tekshirish

```bash
cd /Users/sshovkatov/Projects/dictionary/backend
cd gateway && go test ./...
cd ../libs/shared && go test ./...
cd ../../modules/auth && go test ./...
cd ../vocabulary && go test ./...
cd ../../gateway && go run ./cmd/server
```

## Keyingi qadam (Phase 2)

- Vocabulary create/list/search
- Admin create endpoint
- JWT middleware va protected route
- DB ulash va repository qatlami

