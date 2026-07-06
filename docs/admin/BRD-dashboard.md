# Business Requirements Document (BRD): Dictionary Admin Dashboard

## 1. Umumiy maqsad

Admin Dashboard — Flutter Web asosidagi ma'muriy ilova.
Maqsad: user boshqaruvi, lug'at moderatsiyasi, tizim nazorati.

## 2. Asosiy funksiyalar

### A. User Management (CRUD)

- Create: yangi user yaratish (`role=client` default).
- Read: user list, search, filter.
- Update: name/email yangilash.
- Delete/Block: userni o'chirish yoki bloklash.

### B. Word Management (CRUD)

- Create/Update/Delete: global moderatsiya huquqi.
- Search: word/translation/category bo'yicha kengaytirilgan qidiruv.

### C. System & Database Administration

- DB Backup/Export: dictionary/users ma'lumotlarini export (`.sql` yoki `.json`).
- Monitoring: Grafana dashboard WebView orqali.

## 3. Texnik talablar

- Platforma: Flutter Web (responsive).
- Auth: admin JWT (`role=admin`).
- Security: admin action audit log.
- API access: faqat API Gateway orqali.

## 4. Admin User Journey

1. Login -> dashboard KPI (users count, words count).
2. User section -> create/update/block/delete.
3. Dictionary section -> moderatsiya (edit/delete).
4. Admin tools -> backup/export yuklab olish.
5. Monitoring section -> Grafana WebView.

## 5. RBAC va xavfsizlik

- Gateway admin endpointlarda role check qiladi.
- Auth service admin role validatsiyasini tasdiqlaydi.
- `Create/Delete/Block` actionlar audit logga yoziladi.

## 6. Servislar bilan interaction

- Yangi user yaratish: Users Service (+ Auth Service role policy).
- So'z moderatsiyasi: Dictionary Service.
- Backup/export: Admin Service yoki backup agent.
- Monitoring: Prometheus -> Grafana.

## 7. Backend dependency (Admin uchun)

- Users CRUD + block/unblock endpointlar.
- Dictionary global CRUD/moderation endpointlar.
- Backup/export endpointlar.
- Dashboard stats endpointlar.
- Audit log read endpointlar.
