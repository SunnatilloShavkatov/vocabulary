import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:oson_vocabulary_mobile/src/vocabulary/vocabulary_api.dart';

void main() {
  test('fetchVocabulary parses response', () async {
    final dio = Dio(BaseOptions());
    dio.httpClientAdapter = _MockAdapter(
      onRequest: (options) {
        expect(options.path, '/v1/vocabulary');
        return ResponseBody.fromString(
          '{"items":[{"id":"1","word":"apple","translation":"olma","example":"apple pie"}],"meta":{"total":1}}',
          200,
          headers: {
            Headers.contentTypeHeader: [Headers.jsonContentType],
          },
        );
      },
    );

    final api = VocabularyApi(dio: dio);

    final result = await api.fetchVocabulary(search: 'apple');

    expect(result.total, 1);
    expect(result.items.length, 1);
    expect(result.items.first.word, 'apple');
    expect(result.items.first.translation, 'olma');
  });

  test('fetchVocabulary throws server error message', () async {
    final dio = Dio(BaseOptions());
    dio.httpClientAdapter = _MockAdapter(
      onRequest: (_) => ResponseBody.fromString(
        '{"error":"bad request"}',
        400,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      ),
    );
    final api = VocabularyApi(dio: dio);

    expect(() => api.fetchVocabulary(), throwsA(isA<Exception>()));
  });
}

class _MockAdapter implements HttpClientAdapter {
  _MockAdapter({required this.onRequest});

  final ResponseBody Function(RequestOptions options) onRequest;

  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return onRequest(options);
  }
}
