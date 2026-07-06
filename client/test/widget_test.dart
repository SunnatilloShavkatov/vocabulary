import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:oson_vocabulary_mobile/src/auth/login_page.dart';

void main() {
  testWidgets('Shows login page', (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(home: LoginPage(onLogin: (_, __) async {})),
    );

    expect(find.text('Oson Vocabulary Login'), findsOneWidget);
    expect(find.text('Login'), findsOneWidget);
  });
}
