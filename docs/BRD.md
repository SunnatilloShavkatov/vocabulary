# Business Requirements Document (BRD): Dictionary Platform

## 1. Loyiha Umumiy Ko'rinishi

Dictionary Platform — so'zlarni yodlash uchun mo'ljallangan ekotizim.
Asosiy mexanizm: Spaced Repetition System (SRS) (`1, 3, 7, 30` kunlik interval).
Tizim 2 asosiy interfeysdan iborat:

- Client App (Flutter Mobile: Android/iOS)
- Admin Panel (Flutter Web)

## 2. Tizim Arxitekturasi (Stack)

- Backend: Go (Monorepo, Microservices)
- External communication: REST API
- Internal communication: gRPC
- Async communication: RabbitMQ (event-driven)
- Database: PostgreSQL (database-per-service)
- Client App: Flutter (Android/iOS)
- Admin Panel: Flutter Web
- Monitoring: Prometheus + Grafana
- API Documentation: Swagger (OpenAPI)

## 3. Foydalanuvchilar va Interfeyslar

### A) Client App (Flutter - Android/iOS)

Maqsad:

- So'zlarni qo'shish, yodlash, eslatmalar qabul qilish.

Asosiy funksiyalar:

- Login/Register
- Lug'at CRUD (user scope)
- Push-notification (Firebase FCM)
- Progress dashboard (yodlash tarixi)

### B) Admin Panel (Flutter - Web)

Maqsad:

- Tizimni boshqarish va moderatsiya qilish.

Asosiy funksiyalar:

- Foydalanuvchilarni boshqarish (bloklash/o'chirish)
- Lug'at moderatsiyasi (tekshirish/o'chirish)
- Monitoring sahifasi (Grafana dashboard WebView orqali)

## 4. Microservice Modullari va Mas'uliyatlar

- Auth Service:
  JWT token issue/refresh, public key orqali validatsiya contractlari.
- Users Service:
  Profil, role (`admin`/`user`), account holati.
- Dictionary Service:
  Lug'at CRUD, `WordAdded` event publish.
- Notification Service:
  SRS (`1,3,7,30`), schedule yaratish, cron, FCM yuborish.

## 5. Security (Hybrid)

- Layer 1: API Gateway
  JWT signature/expiration validate, `X-User-ID`, `X-User-Role` uzatish.
- Layer 2: Microservices
  RBAC tekshiruv (`admin`/`user`) business logic darajasida.

## 6. Event-Driven Oqim

1. Client/App so'z qo'shadi (`Dictionary Service`).
2. Dictionary Service DBga yozadi, keyin `WordAdded` event RabbitMQga publish qiladi.
3. Notification Service eventni consume qiladi, `next_review_date`/schedules saqlaydi.
4. Cron due schedulelarni olib, FCM orqali reminder yuboradi.

## 7. Data Ownership (DB-per-Service)

Har service faqat o'z DBsi bilan ishlaydi:

- `auth-db`
- `users-db`
- `dictionary-db`
- `notification-db`

Qoida:

- Service boshqa service DBsiga direct query qilmaydi.
- Cross-service data faqat gRPC yoki event orqali olinadi.

## 8. Monitoring va Dokumentatsiya

- Har service metrics endpoint beradi (Prometheus scrape uchun).
- Grafana dashboardlar:
  latency, error rate, queue lag, notification delivery.
- Swagger/OpenAPI hujjatlari har service uchun alohida manzilda bo'ladi.

## 9. Amaliy Qoidalar

- Admin va Client ilovalar faqat REST orqali gatewayga ulanadi.
- Ichki service chaqiriqlari REST emas, gRPC bo'ladi.
- Har microservice alohida run/deploy qilinadi.

## 10. Keyingi Boshlash Tartibi

1. Auth Service kodini alohida service runtime sifatida boshlash (`cmd/server`).
2. gRPC proto contractlar (`auth/users/dictionary/notification`) yaratish.
3. Gatewayni ichki gRPC calllarga o'tkazish.
4. DB-per-service migratsiyasini bosqichma-bosqich qilish.
5. Client/Admin contractlarini shu BRDga qat'iy moslab yakunlash.