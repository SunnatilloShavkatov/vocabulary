# Backend Foundation (Phase 1)

Bu papka `Vocabulary` loyihasining Go backend foundation qismini saqlaydi.

## Tuzilma

- `go.work` - barcha backend modullarni birlashtiruvchi workspace
- `gateway/` - alohida Go module, API kirish nuqtasi
- `modules/auth/` - alohida Go module (`controller` + `service`)
- `modules/vocabulary/` - alohida Go module (`controller` + `service` + `repository`)
- `libs/shared/` - alohida Go module, umumiy config/logger contractlar
- `migrations/` - SQL migration fayllari

## Hozirgi holat (Phase 1)

- Har service alohida Go module sifatida ajratilgan
- Transport va business logic ajratildi (`controller` qatlam qo'shildi)
- Har service o'z endpointlarini o'zi register qiladi
- `gateway` faqat composition point sifatida barcha routerlarni birlashtiradi
- `GET /healthz` endpoint ishlaydi
- `POST /v1/auth/login` bootstrap admin bilan ishlaydi
- `POST /v1/vocabulary` JWT bilan himoyalangan va DB ga yozadi
- `GET /v1/vocabulary` public list/search endpoint ishlaydi
- `POST /v1/admins` hozircha skeleton (`501 not implemented`)
- `admins` va `vocabularies` uchun migration draft bor
- Docker Compose bilan local run/test tayyor

## Docker bilan ishlatish (tavsiya)

1) `backend/.env.docker.example` nusxa olib `.env` qiling (xohlasangiz o'zgartiring):

```bash
cd /Users/sshovkatov/Projects/Vocabulary/backend
cp .env.docker.example .env
```

2) Testlarni container ichida ishga tushiring:

```bash
cd /Users/sshovkatov/Projects/Vocabulary/backend
docker compose --profile test run --rm gateway-test
```

3) Postgres + migrate + gateway stackni ko'taring:

```bash
cd /Users/sshovkatov/Projects/Vocabulary/backend
docker compose up -d postgres migrate gateway
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
cd /Users/sshovkatov/Projects/Vocabulary/backend
docker compose down
```

## Native (ixtiyoriy) tekshirish

```bash
cd /Users/sshovkatov/Projects/Vocabulary/backend
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

