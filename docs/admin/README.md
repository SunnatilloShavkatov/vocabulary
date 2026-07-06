# Admin Docs

## Role

`admin/` ilova admin operatsiyalarini bajaradi.
Faoliyat faqat REST orqali `api-gateway` ga chiqadi.

## Integration Contract

- Protocol: HTTP REST.
- Base URL: gateway (`/v1/...`).
- Auth: Bearer JWT (`Authorization: Bearer <token>`).

## Admin Required Flows

1. Login (`POST /v1/auth/login`).
2. User CRUD (`create/read/update/delete/block`).
3. Vocabulary global CRUD/moderation.
4. Dashboard stats (users/words totals).
5. DB Backup/Export (`.sql` yoki `.json`).
6. Monitoring sahifasi (Grafana dashboard WebView embed).

## Non-Goals

- Admin app gRPC ishlatmaydi.
- Admin app ichki microservice hostlarini bilmaydi.
- Admin app bevosita DBga ulanmaydi.

## Error Expectations

- `401`: token yo'q/yaroqsiz.
- `403`: role yetarli emas.
- `429`: rate-limit.
- `5xx`: gateway yoki service ichki xato.

## Security Notes

- Token storage secure bo'lishi kerak.
- Refresh/log out oqimi auth contractga mos bo'lishi kerak.
- Admin actionlar audit logga yozilishi kerak.

## Setup Priority

1. Auth flowni production-ready qilish (refresh, logout, session expire UX).
2. User CRUD + block/delete screenlarni gateway RBAC endpointlariga ulash.
3. Dictionary moderation va advanced searchni ulash.
4. Backup/export tugmalarini backend endpointlari bilan ulash.
5. Grafana dashboard sahifasini WebView bilan bog'lash.
