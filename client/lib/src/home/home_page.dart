import 'package:flutter/material.dart';

import '../user/profile_repository.dart';
import '../vocabulary/vocabulary_api.dart';

class HomePage extends StatefulWidget {
  const HomePage({
    super.key,
    required this.vocabularyApi,
    required this.profileRepository,
    required this.displayName,
    required this.onLogout,
  });

  final VocabularyApi vocabularyApi;
  final ProfileRepository profileRepository;
  final String displayName;
  final Future<void> Function() onLogout;

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  final _searchController = TextEditingController();
  bool _isLoading = true;
  bool _isProfileLoading = true;
  String? _error;
  String? _profileError;
  String _name = '';
  List<VocabularyItem> _items = const [];
  int _total = 0;

  @override
  void initState() {
    super.initState();
    _name = widget.displayName;
    _loadProfile();
    _loadVocabulary();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadProfile() async {
    setState(() {
      _isProfileLoading = true;
      _profileError = null;
    });
    try {
      final profile = await widget.profileRepository.me();
      if (!mounted) {
        return;
      }
      setState(() {
        if (profile.name.trim().isNotEmpty) {
          _name = profile.name.trim();
        }
      });
    } catch (e) {
      if (mounted) {
        setState(() => _profileError = e.toString());
      }
    } finally {
      if (mounted) {
        setState(() => _isProfileLoading = false);
      }
    }
  }

  Future<void> _loadVocabulary() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final result = await widget.vocabularyApi.fetchVocabulary(
        search: _searchController.text.trim(),
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _items = result.items;
        _total = result.total;
      });
    } catch (e) {
      if (mounted) {
        setState(() => _error = e.toString());
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  Future<void> _openAddWordDialog() async {
    final added = await showDialog<bool>(
      context: context,
      builder: (context) => AddWordDialog(vocabularyApi: widget.vocabularyApi),
    );
    if (added == true) {
      await _loadVocabulary();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(
          _name.trim().isEmpty ? 'Oson Vocabulary Home' : 'Hello, $_name',
        ),
        actions: [
          IconButton(
            onPressed: widget.onLogout,
            icon: const Icon(Icons.logout),
            tooltip: 'Logout',
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _openAddWordDialog,
        child: const Icon(Icons.add),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (_profileError != null)
              Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: Text(
                  _profileError!,
                  style: const TextStyle(color: Colors.red),
                ),
              ),
            if (_isProfileLoading)
              const Padding(
                padding: EdgeInsets.only(bottom: 8),
                child: LinearProgressIndicator(),
              ),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _searchController,
                    decoration: const InputDecoration(
                      labelText: 'Search',
                      hintText: 'word or translation',
                      border: OutlineInputBorder(),
                    ),
                    onChanged: (_) => _loadVocabulary(),
                  ),
                ),
                const SizedBox(width: 8),
                ElevatedButton(
                  onPressed: _isLoading ? null : _loadVocabulary,
                  child: const Text('Refresh'),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Text('Total: $_total'),
            const SizedBox(height: 12),
            Expanded(child: _buildContent()),
          ],
        ),
      ),
    );
  }

  Widget _buildContent() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_error != null) {
      return Center(
        child: Text(_error!, style: const TextStyle(color: Colors.red)),
      );
    }
    if (_items.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: const [
            Icon(Icons.menu_book_outlined, size: 56),
            SizedBox(height: 12),
            Text('Lugat bosh. FAB bosib birinchi soz qoshing.'),
          ],
        ),
      );
    }

    return ListView.separated(
      itemCount: _items.length,
      separatorBuilder: (_, _) => const Divider(height: 1),
      itemBuilder: (context, index) {
        final item = _items[index];
        return ListTile(
          title: Text(item.word),
          subtitle: Text(item.translation),
          trailing: item.example.isEmpty
              ? null
              : Tooltip(
                  message: item.example,
                  child: const Icon(Icons.info_outline),
                ),
        );
      },
    );
  }
}

class AddWordDialog extends StatefulWidget {
  const AddWordDialog({super.key, required this.vocabularyApi});

  final VocabularyApi vocabularyApi;

  @override
  State<AddWordDialog> createState() => _AddWordDialogState();
}

class _AddWordDialogState extends State<AddWordDialog> {
  final _wordController = TextEditingController();
  final _translationController = TextEditingController();
  final _definitionController = TextEditingController();
  bool _isSaving = false;
  String? _error;

  @override
  void dispose() {
    _wordController.dispose();
    _translationController.dispose();
    _definitionController.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final word = _wordController.text.trim();
    final translation = _translationController.text.trim();
    final definition = _definitionController.text.trim();

    if (word.isEmpty || translation.isEmpty) {
      setState(() => _error = 'Word va translation required');
      return;
    }

    setState(() {
      _isSaving = true;
      _error = null;
    });

    try {
      await widget.vocabularyApi.createVocabulary(
        word: word,
        translation: translation,
        definition: definition,
      );
      if (mounted) {
        Navigator.of(context).pop(true);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _error = e.toString());
      }
    } finally {
      if (mounted) {
        setState(() => _isSaving = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Add Word'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: _wordController,
              decoration: const InputDecoration(labelText: 'Word'),
            ),
            const SizedBox(height: 10),
            TextField(
              controller: _translationController,
              decoration: const InputDecoration(labelText: 'Translation'),
            ),
            const SizedBox(height: 10),
            TextField(
              controller: _definitionController,
              decoration: const InputDecoration(labelText: 'Definition'),
            ),
            if (_error != null) ...[
              const SizedBox(height: 10),
              Text(_error!, style: const TextStyle(color: Colors.red)),
            ],
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: _isSaving ? null : () => Navigator.of(context).pop(false),
          child: const Text('Cancel'),
        ),
        ElevatedButton(
          onPressed: _isSaving ? null : _save,
          child: Text(_isSaving ? 'Saving...' : 'Save'),
        ),
      ],
    );
  }
}
