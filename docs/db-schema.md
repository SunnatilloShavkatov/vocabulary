# DB Schema (Draft)

MVP uchun bitta PostgreSQL baza ishlatiladi.

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
- `created_by` (uuid, fk -> admins.id)
- `created_at` (timestamp, not null)

## Indexlar

- `admins(email)` unique
- `vocabularies(word)`
- `vocabularies(translation)`

## Eslatma

Qidiruv sifati uchun keyingi bosqichda `GIN` yoki full-text search qo'shiladi.

