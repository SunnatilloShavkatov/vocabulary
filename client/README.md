# Oson Vocabulary Mobile

Phase 4 mobile client.

Current MVP pages:

- Welcome page
- Vocabulary list
- Search input (`GET /v1/vocabulary`)

## Run

```bash
cd /Users/sshovkatov/Projects/Vocabulary/client
flutter pub get
flutter run --dart-define=API_BASE_URL=http://localhost:8080
```

If `API_BASE_URL` is not provided, default is `http://localhost:8080`.
