import 'dart:convert';

import 'package:http/http.dart' as http;

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
  VocabularyApi({http.Client? client}) : _client = client ?? http.Client();

  final http.Client _client;
  final String _baseUrl = const String.fromEnvironment('API_BASE_URL', defaultValue: 'http://localhost:8080');

  Future<VocabularyListResult> fetchVocabulary({String search = '', int page = 1, int limit = 20}) async {
    final query = <String, String>{
      'search': search,
      'page': page.toString(),
      'limit': limit.toString(),
    };
    final uri = Uri.parse('$_baseUrl/v1/vocabulary').replace(queryParameters: query);
    final response = await _client.get(uri);

    if (response.statusCode != 200) {
      throw Exception(_extractError(response.body, fallback: 'Failed to fetch vocabulary'));
    }

    final map = jsonDecode(response.body) as Map<String, dynamic>;
    final rawItems = map['items'] as List<dynamic>? ?? const [];
    final items = rawItems
        .whereType<Map<String, dynamic>>()
        .map(VocabularyItem.fromJson)
        .toList(growable: false);

    final meta = map['meta'] as Map<String, dynamic>? ?? const {};
    final total = (meta['total'] as num?)?.toInt() ?? items.length;
    return VocabularyListResult(items: items, total: total);
  }

  String _extractError(String body, {required String fallback}) {
    try {
      final map = jsonDecode(body) as Map<String, dynamic>;
      final message = map['error'] as String?;
      if (message != null && message.isNotEmpty) {
        return message;
      }
    } catch (_) {}
    return fallback;
  }
}

