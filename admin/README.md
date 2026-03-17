# Oson Vocabulary Admin

Phase 3 admin web client.

Current MVP pages:

- Login page (`POST /v1/auth/login`)
- Add vocabulary form (`POST /v1/vocabulary`)
- Add admin form (`POST /v1/admins`)

## Run

```bash
cd /Users/sshovkatov/Projects/Vocabulary/admin
flutter pub get
flutter run -d chrome --dart-define=API_BASE_URL=http://localhost:8080
```

If `API_BASE_URL` is not provided, default is `http://localhost:8080`.
