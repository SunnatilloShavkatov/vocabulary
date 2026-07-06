import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'auth/login_page.dart';
import 'auth/session_controller.dart';
import 'core/providers.dart';
import 'home/home_page.dart';

class OsonVocabularyMobileApp extends ConsumerStatefulWidget {
  const OsonVocabularyMobileApp({super.key});

  @override
  ConsumerState<OsonVocabularyMobileApp> createState() =>
      _OsonVocabularyMobileAppState();
}

class _OsonVocabularyMobileAppState
    extends ConsumerState<OsonVocabularyMobileApp> {
  @override
  void initState() {
    super.initState();
    Future<void>.microtask(
      () => ref.read(sessionControllerProvider.notifier).bootstrap(),
    );
  }

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(sessionControllerProvider);

    return MaterialApp(
      title: 'Oson Vocabulary',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.teal),
      ),
      home: switch (session.status) {
        SessionStatus.booting => const Scaffold(
          body: Center(child: CircularProgressIndicator()),
        ),
        SessionStatus.authenticated => HomePage(
          vocabularyApi: ref.watch(vocabularyApiProvider),
          profileRepository: ref.watch(profileRepositoryProvider),
          displayName: session.tokens?.name ?? '',
          onLogout: () => ref.read(sessionControllerProvider.notifier).logout(),
        ),
        SessionStatus.unauthenticated => LoginPage(
          errorText: session.errorMessage,
          onLogin: (email, password) => ref
              .read(sessionControllerProvider.notifier)
              .login(email: email, password: password),
        ),
      },
    );
  }
}
