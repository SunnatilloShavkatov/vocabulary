import 'package:dio/dio.dart';

import '../auth/auth_repository.dart';

class ApiClient {
  ApiClient({required this.baseUrl, required this.authRepository})
    : _dio = Dio(BaseOptions(baseUrl: baseUrl)) {
    _dio.interceptors.add(
      LogInterceptor(responseBody: true, requestBody: true),
    );
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          final tokens = authRepository.cachedTokens;
          if (tokens != null && tokens.accessToken.trim().isNotEmpty) {
            options.headers['Authorization'] = 'Bearer ${tokens.accessToken}';
          }
          handler.next(options);
        },
        onError: (err, handler) async {
          final status = err.response?.statusCode ?? 0;
          final alreadyRetried = err.requestOptions.extra['retried'] == true;
          if (status == 401 && !alreadyRetried) {
            final refreshed = await authRepository.tryRefresh();
            if (refreshed != null && refreshed.accessToken.trim().isNotEmpty) {
              final request = err.requestOptions;
              request.headers['Authorization'] =
                  'Bearer ${refreshed.accessToken}';
              request.extra['retried'] = true;
              try {
                final response = await _dio.fetch(request);
                handler.resolve(response);
                return;
              } catch (_) {
                // pass through original error
              }
            }
          }
          handler.next(err);
        },
      ),
    );
  }

  final String baseUrl;
  final AuthRepository authRepository;
  final Dio _dio;

  Dio get dio => _dio;
}
