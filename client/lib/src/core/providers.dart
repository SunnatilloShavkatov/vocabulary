import 'package:flutter_riverpod/legacy.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../auth/auth_repository.dart';
import '../auth/session_controller.dart';
import '../user/profile_repository.dart';
import '../vocabulary/vocabulary_api.dart';
import 'api_client.dart';

const _baseUrl = String.fromEnvironment('API_BASE_URL', defaultValue: 'http://localhost:8080');

final secureStorageProvider = Provider<FlutterSecureStorage>((ref) {
  return const FlutterSecureStorage();
});

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(storage: ref.watch(secureStorageProvider), baseUrl: _baseUrl);
});

final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient(baseUrl: _baseUrl, authRepository: ref.watch(authRepositoryProvider));
});

final sessionControllerProvider = StateNotifierProvider<SessionController, SessionState>((ref) {
  return SessionController(ref.watch(authRepositoryProvider));
});

final vocabularyApiProvider = Provider<VocabularyApi>((ref) {
  return VocabularyApi(dio: ref.watch(apiClientProvider).dio);
});

final profileRepositoryProvider = Provider<ProfileRepository>((ref) {
  return ProfileRepository(dio: ref.watch(apiClientProvider).dio);
});
