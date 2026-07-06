import 'dart:convert';

import 'package:dio/dio.dart';

class VocabularyItem {
  const VocabularyItem({
    required this.id,
    required this.word,
    required this.translation,
    required this.example,
  });

  final String id;
  final String word;
  final String translation;
  final String example;

  factory VocabularyItem.fromJson(Map<String, dynamic> json) {
    return VocabularyItem(
      id: json['id'] as String? ?? '',
      word: json['word'] as String? ?? '',
      translation: json['translation'] as String? ?? '',
      example: json['example'] as String? ?? '',
    );
  }
}

class VocabularyListResult {
  const VocabularyListResult({required this.items, required this.total});

  final List<VocabularyItem> items;
  final int total;
}

class VocabularyApi {
  VocabularyApi({required Dio dio}) : _dio = dio;

  final Dio _dio;

  Future<VocabularyListResult> fetchVocabulary({
    String search = '',
    int page = 1,
    int limit = 20,
  }) async {
    final query = <String, String>{
      'search': search,
      'page': page.toString(),
      'limit': limit.toString(),
    };
    final response = await _dio.get<Map<String, dynamic>>(
      '/v1/vocabulary',
      queryParameters: query,
    );

    if (response.statusCode != 200 || response.data == null) {
      throw Exception(
        _extractError(response.data, fallback: 'Failed to fetch vocabulary'),
      );
    }

    final map = response.data!;
    final rawItems = map['items'] as List<dynamic>? ?? const [];
    final items = rawItems
        .whereType<Map<String, dynamic>>()
        .map(VocabularyItem.fromJson)
        .toList(growable: false);

    final meta = map['meta'] as Map<String, dynamic>? ?? const {};
    final total = (meta['total'] as num?)?.toInt() ?? items.length;
    return VocabularyListResult(items: items, total: total);
  }

  Future<void> createVocabulary({
    required String word,
    required String translation,
    required String definition,
  }) async {
    final response = await _dio.post<Map<String, dynamic>>(
      '/v1/vocabulary',
      data: {
        'word': word.trim(),
        'translation': translation.trim(),
        'example': definition.trim(),
      },
      options: Options(headers: {'Content-Type': 'application/json'}),
    );
    if (response.statusCode != 201) {
      throw Exception(
        _extractError(response.data, fallback: 'Failed to create vocabulary'),
      );
    }
  }

  String _extractError(dynamic body, {required String fallback}) {
    if (body is Map<String, dynamic>) {
      final message = body['error'] as String?;
      if (message != null && message.isNotEmpty) {
        return message;
      }
    }
    if (body is String && body.trim().isNotEmpty) {
      try {
        final map = jsonDecode(body) as Map<String, dynamic>;
        final message = map['error'] as String?;
        if (message != null && message.isNotEmpty) {
          return message;
        }
      } catch (_) {}
    }
    return fallback;
  }

  Future<void> updateVocabulary({
    required String id,
    required String word,
    required String translation,
    required String definition,
  }) async {
    final response = await _dio.patch<Map<String, dynamic>>(
      '/v1/admin/vocabulary/$id',
      data: {
        'word': word.trim(),
        'translation': translation.trim(),
        'example': definition.trim(),
      },
      options: Options(headers: {'Content-Type': 'application/json'}),
    );
    if (response.statusCode != 200) {
      throw Exception(
        _extractError(response.data, fallback: 'Failed to update vocabulary'),
      );
    }
  }

  Future<void> deleteVocabulary(String id) async {
    final response = await _dio.delete<void>('/v1/admin/vocabulary/$id');
    if (response.statusCode != 204) {
      throw Exception('Failed to delete vocabulary');
    }
  }

  String _extractErrorFromResponse(
    Response<dynamic> response, {
    required String fallback,
  }) {
    return _extractError(response.data, fallback: fallback);
  }

  Future<void> registerNotificationToken(String token) async {
    final response = await _dio.post<Map<String, dynamic>>(
      '/v1/notifications/token',
      data: {'token': token},
      options: Options(headers: {'Content-Type': 'application/json'}),
    );
    if (response.statusCode != 200 &&
        response.statusCode != 201 &&
        response.statusCode != 202) {
      throw Exception(
        _extractErrorFromResponse(
          response,
          fallback: 'Failed to register notification token',
        ),
      );
    }
  }
}
