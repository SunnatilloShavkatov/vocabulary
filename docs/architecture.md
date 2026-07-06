# Architecture

## Maqsad

`Dictionary Platform` loyihasi 3 qismdan iborat:

1. Admin panel (Flutter Web)
2. Oson Vocabulary client app (Flutter)
3. Backend (Go microservices)

## Monorepo struktura

```text
Vocabulary/
  admin/                 # Flutter Web admin
  client/                # Oson Vocabulary Flutter client (rejalashtirilgan)
  backend/
    go.work              # backend workspace
    gateway/             # alohida Go module, API kirish nuqtasi
    modules/
      auth/              # login + admin create
      users/             # profile/settings contract
      vocabulary/        # dictionary CRUD (MVP create/list/search)
      notification/      # SRS schedule logic (1-3-7-30)
    libs/
      shared/            # alohida Go module, common contract/config
    migrations/          # DB migration fayllari
  docs/
    ...
```

## Komponentlar vazifasi

### 1) API Gateway

- Tashqi API endpointlarni beradi (`/v1/...`)
- Admin/client ilovalar uchun yagona REST entrypoint
- Auth middleware orqali tokenni tekshiradi
- Valid tokendan `X-User-ID` va `X-User-Role` metadata beradi
- Ichki service chaqiruvini gRPC orqali bajaradi

### 2) Auth Service

- Admin login (`email` + `password`)
- JWT token berish (`access`, keyinroq `refresh`)
- Yangi admin yaratish (role asosida)
- JWT issue/refresh va key management
- Boshqa service/ gateway bilan gRPC contract orqali ishlaydi
- O'z ma'lumotlar bazasiga ega (`auth-service-db`)

### 3) Users Service

- Profil va settings state boshqaradi
- O'z ma'lumotlar bazasiga ega (`users-service-db`)
- gRPC orqali boshqa servicelarga user context beradi

### 4) Dictionary Service

- So'z qo'shish
- So'zlar ro'yxatini qaytarish
- Search (`word`, `translation` bo'yicha)
- So'z CRUD va search funksiyasini bajaradi
- O'z ma'lumotlar bazasiga ega (`dictionary-service-db`)
- `WordAdded` eventini RabbitMQ orqali chiqaradi

### 5) Notification Service

- `WordAdded` eventini qabul qiladi
- `1-3-7-30` kunlik SRS jadvalini hisoblaydi
- Cron orqali due schedulelarni ishlab chiqadi
- FCM orqali clientga push yuboradi
- O'z ma'lumotlar bazasiga ega (`notification-service-db`)

## Communication Model

- Admin/client <-> backend: HTTP REST (faqat Gateway orqali).
- Gateway <-> microservices: gRPC.
- Microservice <-> microservice: gRPC.
- Async: RabbitMQ events.

## Event-driven oqim

1. `Dictionary Service` so'zni saqlaydi (`dictionary-service-db`).
2. `WordAdded` event RabbitMQ orqali `Notification Service` ga uzatiladi.
3. `Notification Service` `notification-service-db` ga `schedules` yozadi (`1, 3, 7, 30`).
4. Cron worker due schedulelarni olib FCM push yuboradi.

## Observability

- Gateway `GET /metrics` endpointi Prometheus formatida metrikalar beradi.
- `docker-compose` ichida RabbitMQ + Prometheus + Grafana stack qo'shilgan.

## DB-per-Service qoidasi

- Har service faqat o'z DBsini o'qiydi/yozadi.
- Boshqa service DBga direct query taqiqlanadi.
- Cross-service data gRPC yoki event orqali olinadi.

## NFR (MVP)

- API response bir xil formatda (`data`, `error`, `meta`)
- Oddiy audit: kim qachon vocabulary qo'shdi
- Basic monitoring: request log va xato log

