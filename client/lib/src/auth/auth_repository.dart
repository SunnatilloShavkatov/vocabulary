import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../core/session_tokens.dart';

class AuthRepository {
  AuthRepository({required this.storage, required this.baseUrl})
    : _dio = Dio(BaseOptions(baseUrl: baseUrl));

  final FlutterSecureStorage storage;
  final String baseUrl;
  final Dio _dio;

  static const _accessTokenKey = 'access_token';
  static const _refreshTokenKey = 'refresh_token';
  static const _nameKey = 'name';

  SessionTokens? _cachedTokens;

  SessionTokens? get cachedTokens => _cachedTokens;

  Future<SessionTokens?> restoreSession() async {
    final accessToken = await storage.read(key: _accessTokenKey) ?? '';
    final refreshToken = await storage.read(key: _refreshTokenKey) ?? '';
    final name = await storage.read(key: _nameKey) ?? '';

    if (accessToken.trim().isEmpty) {
      _cachedTokens = null;
      return null;
    }

    _cachedTokens = SessionTokens(
      accessToken: accessToken,
      refreshToken: refreshToken,
      name: name,
    );
    return _cachedTokens;
  }

  Future<SessionTokens> login({
    required String email,
    required String password,
  }) async {
    final response = await _dio.post<Map<String, dynamic>>(
      '/v1/auth/login',
      data: {'email': email.trim(), 'password': password},
      options: Options(headers: {'Content-Type': 'application/json'}),
    );

    final body = response.data ?? const <String, dynamic>{};
    final accessToken = (body['access_token'] as String? ?? '').trim();
    final refreshToken = (body['refresh_token'] as String? ?? '').trim();
    final name = (body['name'] as String? ?? '').trim();

    if (accessToken.isEmpty) {
      throw Exception('Login response does not contain access_token');
    }

    final tokens = SessionTokens(
      accessToken: accessToken,
      refreshToken: refreshToken,
      name: name,
    );
    await saveSession(tokens);
    return tokens;
  }

  Future<void> logout() async {
    final tokens = _cachedTokens;
    if (tokens != null) {
      try {
        await _dio.post<void>(
          '/v1/auth/logout',
          data: {'refresh_token': tokens.refreshToken},
          options: Options(
            headers: {
              'Content-Type': 'application/json',
              'Authorization': 'Bearer ${tokens.accessToken}',
            },
          ),
        );
      } catch (_) {
        // Backend logout may not be implemented yet.
      }
    }
    await clearSession();
  }

  Future<SessionTokens?> tryRefresh() async {
    final tokens = _cachedTokens;
    if (tokens == null || !tokens.hasRefreshToken) {
      return null;
    }

    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/v1/auth/refresh',
        data: {'refresh_token': tokens.refreshToken},
        options: Options(headers: {'Content-Type': 'application/json'}),
      );
      final body = response.data ?? const <String, dynamic>{};
      final newAccessToken = (body['access_token'] as String? ?? '').trim();
      final newRefreshToken =
          (body['refresh_token'] as String? ?? tokens.refreshToken).trim();
      final newName = (body['name'] as String? ?? tokens.name).trim();
      if (newAccessToken.isEmpty) {
        return null;
      }
      final updated = SessionTokens(
        accessToken: newAccessToken,
        refreshToken: newRefreshToken,
        name: newName,
      );
      await saveSession(updated);
      return updated;
    } on DioException catch (e) {
      final code = e.response?.statusCode ?? 0;
      if (code == 404 || code == 400 || code == 401) {
        await clearSession();
        return null;
      }
      rethrow;
    }
  }

  Future<void> saveSession(SessionTokens tokens) async {
    _cachedTokens = tokens;
    await storage.write(key: _accessTokenKey, value: tokens.accessToken);
    await storage.write(key: _refreshTokenKey, value: tokens.refreshToken);
    await storage.write(key: _nameKey, value: tokens.name);
  }

  Future<void> clearSession() async {
    _cachedTokens = null;
    await storage.delete(key: _accessTokenKey);
    await storage.delete(key: _refreshTokenKey);
    await storage.delete(key: _nameKey);
  }

  String extractError(dynamic data, {required String fallback}) {
    if (data is Map<String, dynamic>) {
      final msg = data['error'] as String?;
      if (msg != null && msg.isNotEmpty) {
        return msg;
      }
      final detail = data['message'] as String?;
      if (detail != null && detail.isNotEmpty) {
        return detail;
      }
    }
    if (data is String && data.trim().isNotEmpty) {
      try {
        final parsed = jsonDecode(data);
        return extractError(parsed, fallback: fallback);
      } catch (_) {
        return data;
      }
    }
    return fallback;
  }
}
