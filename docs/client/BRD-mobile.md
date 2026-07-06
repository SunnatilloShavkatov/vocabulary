# Business Requirements Document (BRD): Dictionary Client App (Mobile)

## 1. Umumiy maqsad

Dictionary Client App foydalanuvchilarga shaxsiy lug'atni boshqarish, so'z qidirish, SRS asosida takrorlash imkonini beradi.

## 2. Asosiy funksionallik

### A. Auth Flow

- Login email/parol bilan.
- Backenddan `access_token` va `refresh_token` olinadi.
- `name` va tokenlar `flutter_secure_storage` da saqlanadi.
- `401` holatda `refresh_token` bilan avtomatik token yangilash.
- Logout: token/sessionni tozalash, login sahifasiga qaytish.

Eslatma:

- Self-register yo'q. User account admin tomonidan yaratiladi.

### B. Home Page

- User lug'at ro'yxati.
- Real-time search (`TextField`).
- Empty state: so'z qo'shishga undovchi UI.

### C. Add Word (FAB)

- `FAB` orqali add-word modal/page.
- Input: word, translation, definition.
- So'z qo'shilgach backend `WordAdded` event chiqaradi (notification oqimi uchun).

## 3. Texnik talablar

- Platforma: Flutter Android/iOS.
- State management: BLoC yoki Riverpod.
- Networking: Dio + refresh interceptor.
- Secure storage: flutter_secure_storage.
- Notification: Firebase FCM (`1-3-7-30` reminderlar).

## 4. User Journey

1. App open -> Login.
2. Login success -> token/name secure storage.
3. Home list load.
4. FAB orqali so'z qo'shish -> list refresh.
5. Vaqti kelganda FCM reminder olish.

## 5. Xavfsizlik va integratsiya

- Har request `Authorization: Bearer <access_token>` bilan.
- `401` holatda refresh API chaqirish, original request retry.
- Gateway access/refresh token policy client bilan mos bo'lishi kerak.

## 6. Backend Dependency (Client uchun)

- Auth: `login`, `refresh`, `logout` endpointlari.
- Vocabulary: list/search/create/update/delete.
- User profile: name/settings olish.
- Notification token register endpoint.
