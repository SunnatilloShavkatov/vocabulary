# Architecture

## Maqsad

`Vocabulary` loyihasi 3 qismdan iborat:

1. Admin panel (Flutter Web)
2. Oson Vocabulary client app (Flutter)
3. Backend (Go)

## Monorepo struktura

```text
Vocabulary/
  admin/                 # Flutter Web admin
  client/                # Oson Vocabulary Flutter client (rejalashtirilgan)
  backend/
    go.work              # backend workspace
    gateway/             # alohida Go module, API kirish nuqtasi
    modules/
      auth/              # alohida Go module (controller + service)
      vocabulary/        # alohida Go module (controller + service + repository)
    libs/
      shared/            # alohida Go module, common contract/config
    migrations/          # DB migration fayllari
  docs/
    ...
```

## Komponentlar vazifasi

### 1) Gateway

- Tashqi API endpointlarni beradi (`/v1/...`)
- Auth middleware orqali tokenni tekshiradi
- Kerakli service'ga requestni yo'naltiradi
- `auth` va `vocabulary` routerlarini bitta joyda compose qiladi
- Alohida module bo'lgani uchun keyin mustaqil servicega ajratish oson

### 2) Auth service

- Admin login (`email` + `password`)
- JWT token berish (`access`, keyinroq `refresh`)
- Yangi admin yaratish (role asosida)
- O'z routeri orqali auth endpointlariga egalik qiladi
- Dastlab monorepo ichida, keyin micro/macro servicega ko'chirishga tayyor

### 3) Vocabulary service

- So'z qo'shish
- So'zlar ro'yxatini qaytarish
- Search (`word`, `translation` bo'yicha)
- O'z routeri orqali vocabulary endpointlariga egalik qiladi
- Alohida module bo'lgani uchun mustaqil release qilish imkoniyati bor

## Frontend oqimi

- Admin panel faqat login bilan kiradi
- Oson Vocabulary client app public list va qidiruvni ko'rsatadi
- Keyingi bosqichda client auth qo'shilishi mumkin

## NFR (MVP)

- API response bir xil formatda (`data`, `error`, `meta`)
- Oddiy audit: kim qachon vocabulary qo'shdi
- Basic monitoring: request log va xato log

