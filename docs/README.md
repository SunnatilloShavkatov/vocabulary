# Dictionary Platform docs

Bu papka `Vocabulary` monorepo loyihasi uchun asosiy texnik hujjatlarni saqlaydi.

## Docs Structure (3 alohida bo'lim)

- `admin/` - admin ilova contractlari
- `admin/BRD-dashboard.md` - admin dashboard uchun maxsus BRD
- `backend/` - backend microservice arxitekturasi va integratsiya qoidalari
- `client/` - client ilova contractlari
- `client/BRD-mobile.md` - mobil ilova uchun maxsus BRD
- `BRD.md` - yakuniy business requirements (source-of-truth)

## Legacy Hujjatlar

- `architecture.md` - umumiy arxitektura va modul chegaralari
- `mvp-scope.md` - birinchi release (MVP) scope
- `auth-flow.md` - autentifikatsiya va avtorizatsiya oqimi
- `db-schema.md` - ma'lumotlar bazasi sxemasi (draft)
- `roadmap.md` - bosqichma-bosqich ish rejasi
- `api/openapi.yaml` - backend API kontrakti (MVP)

## Qisqa qarorlar

- Admin panel: Flutter Web (`admin/`)
- Oson Vocabulary client app: Flutter (`client/`)
- Backend: Go microservices (`api-gateway`, `auth`, `users`, `dictionary`, `notification`)
- External integration: Admin/Client -> REST via Gateway
- Internal integration: service-to-service gRPC
- Data model: DB-per-service
- Event bus: RabbitMQ (`WordAdded`)
- Observability: Prometheus + Grafana
- Arxitektura: monorepo + modul (service) yondashuv

