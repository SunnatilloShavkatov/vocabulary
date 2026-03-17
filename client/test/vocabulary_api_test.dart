import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:oson_vocabulary_mobile/src/vocabulary/vocabulary_api.dart';

void main() {
  test('fetchVocabulary parses response', () async {
    final api = VocabularyApi(
      client: MockClient((request) async {
        expect(request.url.path, '/v1/vocabulary');
        return http.Response(
          '{"items":[{"id":"1","word":"apple","translation":"olma","example":"apple pie"}],"meta":{"total":1}}',
          200,
          headers: {'content-type': 'application/json'},
        );
      }),
    );

    final result = await api.fetchVocabulary(search: 'apple');

    expect(result.total, 1);
    expect(result.items.length, 1);
    expect(result.items.first.word, 'apple');
    expect(result.items.first.translation, 'olma');
  });

  test('fetchVocabulary throws server error message', () async {
    final api = VocabularyApi(
      client: MockClient((_) async {
        return http.Response('{"error":"bad request"}', 400);
      }),
    );

    expect(
      () => api.fetchVocabulary(),
      throwsA(isA<Exception>()),
    );
  });
}

