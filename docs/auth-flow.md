# Auth Flow

## Rollar

- `admin`: vocabulary qo'shadi, admin qo'shadi
- `client` (Oson Vocabulary): hozircha login talab qilinmaydi (public read)

## Login oqimi (admin)

1. Admin `POST /v1/auth/login` ga `email/password` yuboradi.
2. Auth service hozircha local bootstrap admin credentialsni tekshiradi.
3. To'g'ri bo'lsa `access_token` (JWT) qaytaradi.
4. Admin panel keyingi protected so'rovlarda `Authorization: Bearer <token>` yuboradi.

## Local bootstrap admin

Local development uchun vaqtinchalik bootstrap admin env orqali beriladi:

- `BOOTSTRAP_ADMIN_EMAIL`
- `BOOTSTRAP_ADMIN_PASSWORD`

Bu faqat MVP/local foundation uchun. Keyingi bosqichda login DB dagi `admins` jadvali orqali ishlaydi.

## Protected endpointlar

- `POST /v1/vocabulary`
- `POST /v1/admins`
- `GET /v1/users/me`
- `POST /internal/notifications/word-added`

Gateway middleware:

- JWT signature va expiry ni tekshiradi
- `X-User-ID` va `X-User-Role` headerlarini qo'shib requestni service'ga uzatadi

Microservice layer:

- Forward qilingan identity headerlarni qabul qiladi
- RBAC qoidalarini endpoint darajasida tekshiradi
- Admin-only operatsiyalar (`POST /v1/admins`) qat'iy nazoratda qoladi

## Xatolik holatlari

- Bootstrap admin sozlanmagan: `503 Service Unavailable`
- Noto'g'ri login: `401 Unauthorized`
- Token yo'q/yaroqsiz: `401 Unauthorized`
- Role yetarli emas: `403 Forbidden`

