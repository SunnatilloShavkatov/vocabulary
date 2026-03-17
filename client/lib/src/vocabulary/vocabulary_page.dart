import 'package:flutter/material.dart';

import 'vocabulary_api.dart';

class VocabularyPage extends StatefulWidget {
  const VocabularyPage({super.key, required this.api});

  final VocabularyApi api;

  @override
  State<VocabularyPage> createState() => _VocabularyPageState();
}

class _VocabularyPageState extends State<VocabularyPage> {
  final _searchController = TextEditingController();
  bool _isLoading = true;
  String? _error;
  List<VocabularyItem> _items = const [];
  int _total = 0;

  @override
  void initState() {
    super.initState();
    _load();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final result = await widget.api.fetchVocabulary(search: _searchController.text.trim());
      if (!mounted) {
        return;
      }
      setState(() {
        _items = result.items;
        _total = result.total;
      });
    } catch (e) {
      if (!mounted) {
        return;
      }
      setState(() => _error = e.toString());
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Vocabulary List')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
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
                    onSubmitted: (_) => _load(),
                  ),
                ),
                const SizedBox(width: 8),
                ElevatedButton(onPressed: _isLoading ? null : _load, child: const Text('Search')),
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
      return Center(child: Text(_error!, style: const TextStyle(color: Colors.red)));
    }
    if (_items.isEmpty) {
      return const Center(child: Text('No vocabulary found.'));
    }

    return ListView.separated(
      itemCount: _items.length,
      separatorBuilder: (_, _) => const Divider(height: 1),
      itemBuilder: (context, index) {
        final item = _items[index];
        return ListTile(
          title: Text(item.word),
          subtitle: Text(item.translation),
          trailing: item.example.isEmpty ? null : Tooltip(message: item.example, child: const Icon(Icons.info_outline)),
        );
      },
    );
  }
}

