import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

void main() {
  runApp(const OsonVocabularyAdminApp());
}

class OsonVocabularyAdminApp extends StatelessWidget {
  const OsonVocabularyAdminApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Oson Vocabulary Admin',
      theme: ThemeData(colorScheme: ColorScheme.fromSeed(seedColor: Colors.indigo)),
      home: const AdminShell(),
    );
  }
}

class AdminShell extends StatefulWidget {
  const AdminShell({super.key});

  @override
  State<AdminShell> createState() => _AdminShellState();
}

class _AdminShellState extends State<AdminShell> {
  final _api = AdminApiClient();
  String? _token;
  String? _error;

  Future<void> _login(String email, String password) async {
    setState(() => _error = null);
    try {
      final token = await _api.login(email: email, password: password);
      setState(() => _token = token);
    } catch (e) {
      setState(() => _error = e.toString());
    }
  }

  void _logout() {
    setState(() {
      _token = null;
      _error = null;
    });
  }

  @override
  Widget build(BuildContext context) {
    if (_token == null) {
      return LoginPage(onLogin: _login, errorText: _error);
    }
    return AdminDashboardPage(api: _api, token: _token!, onLogout: _logout);
  }
}

class LoginPage extends StatefulWidget {
  const LoginPage({super.key, required this.onLogin, this.errorText});

  final Future<void> Function(String email, String password) onLogin;
  final String? errorText;

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  bool _isLoading = false;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() => _isLoading = true);
    await widget.onLogin(_emailController.text.trim(), _passwordController.text);
    if (mounted) {
      setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Oson Vocabulary Admin')),
      body: Center(
        child: SizedBox(
          width: 420,
          child: Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  TextField(controller: _emailController, decoration: const InputDecoration(labelText: 'Email')),
                  const SizedBox(height: 12),
                  TextField(
                    controller: _passwordController,
                    decoration: const InputDecoration(labelText: 'Password'),
                    obscureText: true,
                  ),
                  const SizedBox(height: 16),
                  if (widget.errorText != null)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 8),
                      child: Text(widget.errorText!, style: const TextStyle(color: Colors.red)),
                    ),
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: _isLoading ? null : _submit,
                      child: Text(_isLoading ? 'Signing in...' : 'Login'),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class AdminDashboardPage extends StatefulWidget {
  const AdminDashboardPage({super.key, required this.api, required this.token, required this.onLogout});

  final AdminApiClient api;
  final String token;
  final VoidCallback onLogout;

  @override
  State<AdminDashboardPage> createState() => _AdminDashboardPageState();
}

class _AdminDashboardPageState extends State<AdminDashboardPage> {
  int _tab = 0;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Oson Vocabulary Admin Dashboard'),
        actions: [IconButton(onPressed: widget.onLogout, icon: const Icon(Icons.logout))],
      ),
      body: Row(
        children: [
          NavigationRail(
            selectedIndex: _tab,
            onDestinationSelected: (v) => setState(() => _tab = v),
            labelType: NavigationRailLabelType.all,
            destinations: const [
              NavigationRailDestination(icon: Icon(Icons.translate), label: Text('Add Vocabulary')),
              NavigationRailDestination(icon: Icon(Icons.admin_panel_settings), label: Text('Add Admin')),
            ],
          ),
          const VerticalDivider(width: 1),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: _tab == 0
                  ? AddVocabularyForm(api: widget.api, token: widget.token)
                  : AddAdminForm(api: widget.api, token: widget.token),
            ),
          ),
        ],
      ),
    );
  }
}

class AddVocabularyForm extends StatefulWidget {
  const AddVocabularyForm({super.key, required this.api, required this.token});

  final AdminApiClient api;
  final String token;

  @override
  State<AddVocabularyForm> createState() => _AddVocabularyFormState();
}

class _AddVocabularyFormState extends State<AddVocabularyForm> {
  final _wordController = TextEditingController();
  final _translationController = TextEditingController();
  final _exampleController = TextEditingController();
  String? _message;
  bool _isLoading = false;

  @override
  void dispose() {
    _wordController.dispose();
    _translationController.dispose();
    _exampleController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _isLoading = true;
      _message = null;
    });
    try {
      await widget.api.createVocabulary(
        token: widget.token,
        word: _wordController.text,
        translation: _translationController.text,
        example: _exampleController.text,
      );
      setState(() => _message = 'Vocabulary added successfully.');
      _wordController.clear();
      _translationController.clear();
      _exampleController.clear();
    } catch (e) {
      setState(() => _message = e.toString());
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        TextField(controller: _wordController, decoration: const InputDecoration(labelText: 'Word')),
        const SizedBox(height: 12),
        TextField(controller: _translationController, decoration: const InputDecoration(labelText: 'Translation')),
        const SizedBox(height: 12),
        TextField(controller: _exampleController, decoration: const InputDecoration(labelText: 'Example')),
        const SizedBox(height: 16),
        ElevatedButton(onPressed: _isLoading ? null : _submit, child: Text(_isLoading ? 'Saving...' : 'Save')),
        if (_message != null) ...[
          const SizedBox(height: 12),
          Text(_message!),
        ],
      ],
    );
  }
}

class AddAdminForm extends StatefulWidget {
  const AddAdminForm({super.key, required this.api, required this.token});

  final AdminApiClient api;
  final String token;

  @override
  State<AddAdminForm> createState() => _AddAdminFormState();
}

class _AddAdminFormState extends State<AddAdminForm> {
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  String? _message;
  bool _isLoading = false;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _isLoading = true;
      _message = null;
    });
    try {
      await widget.api.createAdmin(token: widget.token, email: _emailController.text, password: _passwordController.text);
      setState(() => _message = 'Admin added successfully.');
      _emailController.clear();
      _passwordController.clear();
    } catch (e) {
      setState(() => _message = e.toString());
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        TextField(controller: _emailController, decoration: const InputDecoration(labelText: 'Admin email')),
        const SizedBox(height: 12),
        TextField(
          controller: _passwordController,
          decoration: const InputDecoration(labelText: 'Admin password'),
          obscureText: true,
        ),
        const SizedBox(height: 16),
        ElevatedButton(onPressed: _isLoading ? null : _submit, child: Text(_isLoading ? 'Saving...' : 'Save')),
        if (_message != null) ...[
          const SizedBox(height: 12),
          Text(_message!),
        ],
      ],
    );
  }
}

class AdminApiClient {
  AdminApiClient({http.Client? client}) : _client = client ?? http.Client();

  final http.Client _client;
  final String _baseUrl = const String.fromEnvironment('API_BASE_URL', defaultValue: 'http://localhost:8080');

  Future<String> login({required String email, required String password}) async {
    final res = await _client.post(
      Uri.parse('$_baseUrl/v1/auth/login'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'email': email, 'password': password}),
    );
    if (res.statusCode != 200) {
      throw Exception(_extractError(res.body, fallback: 'Login failed'));
    }
    final map = jsonDecode(res.body) as Map<String, dynamic>;
    final token = map['access_token'] as String?;
    if (token == null || token.isEmpty) {
      throw Exception('Login response does not contain access_token');
    }
    return token;
  }

  Future<void> createVocabulary({required String token, required String word, required String translation, String? example}) async {
    final res = await _client.post(
      Uri.parse('$_baseUrl/v1/vocabulary'),
      headers: {'Content-Type': 'application/json', 'Authorization': 'Bearer $token'},
      body: jsonEncode({'word': word, 'translation': translation, 'example': example ?? ''}),
    );
    if (res.statusCode != 201) {
      throw Exception(_extractError(res.body, fallback: 'Create vocabulary failed'));
    }
  }

  Future<void> createAdmin({required String token, required String email, required String password}) async {
    final res = await _client.post(
      Uri.parse('$_baseUrl/v1/admins'),
      headers: {'Content-Type': 'application/json', 'Authorization': 'Bearer $token'},
      body: jsonEncode({'email': email, 'password': password}),
    );
    if (res.statusCode != 201) {
      throw Exception(_extractError(res.body, fallback: 'Create admin failed'));
    }
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

