# DB Schema (BRD-aligned draft)

Hozirgi holatda monolith DB ishlatilmoqda, lekin sxema servislar kesimida ajratishga tayyor.

## Tables

### `admins`

- `id` (uuid, pk)
- `email` (text, unique, not null)
- `password_hash` (text, not null)
- `role` (text, not null, default `admin`)
- `created_at` (timestamp, not null)

### `vocabularies`

- `id` (uuid, pk)
- `word` (text, not null)
- `translation` (text, not null)
- `example` (text, null)
- `category` (text, null)
- `status` (text, default `pending`)
- `created_by` (uuid, fk -> admins.id)
- `updated_at` (timestamp, not null)
- `created_at` (timestamp, not null)

### `users`

- `id` (uuid, pk)
- `name` (text, not null, default `''`)
- `email` (text, unique, not null)
- `role` (text, not null, default `user`)
- `status` (text, not null, default `active`)
- `settings` (jsonb, not null, default `{}`)
- `updated_at` (timestamp, not null)
- `created_at` (timestamp, not null)

### `schedules`

- `id` (uuid, pk)
- `user_id` (uuid, not null)
- `word_id` (uuid, not null)
- `interval_days` (int, check in `1,3,7,30`)
- `remind_at` (timestamp, not null)
- `status` (text, default `pending`)
- `sent_at` (timestamp, nullable)
- `created_at` (timestamp, not null)

## Indexlar

- `admins(email)` unique
- `vocabularies(word)`
- `vocabularies(translation)`
- `vocabularies(status)`
- `vocabularies(category)`
- `users(status)`
- `schedules(remind_at)` partial (`status='pending'`)
- `schedules(user_id)`

## Eslatma

Qidiruv sifati uchun keyingi bosqichda `GIN` yoki full-text search qo'shiladi.
DB-per-servicega o'tishda jadval egaligi servislar bo'yicha ajratiladi.

