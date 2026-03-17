import 'package:flutter/material.dart';
import 'src/vocabulary/vocabulary_api.dart';
import 'src/vocabulary/vocabulary_page.dart';

void main() {
  runApp(const OsonVocabularyMobileApp());
}

class OsonVocabularyMobileApp extends StatelessWidget {
  const OsonVocabularyMobileApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Oson Vocabulary',
      theme: ThemeData(colorScheme: ColorScheme.fromSeed(seedColor: Colors.teal)),
      home: const WelcomePage(),
    );
  }
}

class WelcomePage extends StatelessWidget {
  const WelcomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Oson Vocabulary')),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Text(
                'Welcome to Oson Vocabulary',
                style: TextStyle(fontSize: 24, fontWeight: FontWeight.w600),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 12),
              const Text(
                'Search and browse saved vocabulary words.',
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 20),
              ElevatedButton(
                onPressed: () {
                  Navigator.of(context).push(
                    MaterialPageRoute<void>(
                      builder: (_) => VocabularyPage(api: VocabularyApi()),
                    ),
                  );
                },
                child: const Text('Open Vocabulary'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

