import 'package:flutter_test/flutter_test.dart';

import 'package:oson_vocabulary_mobile/main.dart';

void main() {
  testWidgets('Shows welcome page by default', (WidgetTester tester) async {
    await tester.pumpWidget(const OsonVocabularyMobileApp());

    expect(find.text('Oson Vocabulary'), findsOneWidget);
    expect(find.text('Welcome to Oson Vocabulary'), findsOneWidget);
    expect(find.text('Open Vocabulary'), findsOneWidget);
  });
}
