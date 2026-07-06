# Client Docs

## Role

`client/` ilova oxirgi foydalanuvchi uchun dictionary experience beradi.
Backend bilan aloqasi faqat REST orqali `api-gateway` dan o'tadi.

## Integration Contract

- Protocol: HTTP REST.
- Base URL: gateway (`/v1/...`).
- Public endpoints: initial vocabulary browse/search.
- Protected endpoints: auth, user profile, vocabulary CRUD, progress, reminders.

Auth constraint:

- Self-register yo'q. Account admin tomonidan yaratiladi.

## Core Flows

1. Login.
2. Vocabulary list olish va search.
3. User scope vocabulary CRUD.
4. FCM reminder qabul qilish.
5. Progress dashboard (yodlash tarixi).
6. Notification open -> so'z detailga o'tish.

## Non-Goals

- Client app gRPC ishlatmaydi.
- Client app microservice hostlarini bilmaydi.
- Client app bevosita DBga ulanmaydi.

## Reliability Expectations

- Offline fallback keyingi bosqich.
- Retry/backoff network xatolarida ishlatiladi.
- `401` holatda re-auth flow ishlaydi.

## Setup Priority

1. Login + refresh/logout UXni auth endpointlar bilan ulash.
2. Vocabulary CRUDni role/user scope bilan ulash.
3. FCM token register + push handlingni yoqish.
4. Progress dashboard endpointlari bilan chart/summaries ulash.

## Token Lifecycle

- `access_token` va `refresh_token` secure storagega yoziladi.
- Dio interceptor `401` holatda refresh qiladi.
- Refresh success bo'lsa original request retry qilinadi.
- Refresh fail bo'lsa logout oqimi ishga tushadi.
