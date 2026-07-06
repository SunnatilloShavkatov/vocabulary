import 'package:flutter_riverpod/legacy.dart';

import '../core/session_tokens.dart';
import 'auth_repository.dart';

enum SessionStatus { booting, authenticated, unauthenticated }

class SessionState {
  const SessionState({required this.status, this.tokens, this.errorMessage});

  final SessionStatus status;
  final SessionTokens? tokens;
  final String? errorMessage;

  const SessionState.booting() : this(status: SessionStatus.booting);
  const SessionState.authenticated(SessionTokens tokens)
    : this(status: SessionStatus.authenticated, tokens: tokens);
  const SessionState.unauthenticated([String? error])
    : this(status: SessionStatus.unauthenticated, errorMessage: error);
}

class SessionController extends StateNotifier<SessionState> {
  SessionController(this._authRepository) : super(const SessionState.booting());

  final AuthRepository _authRepository;
  bool _bootstrapped = false;

  Future<void> bootstrap() async {
    if (_bootstrapped) {
      return;
    }
    _bootstrapped = true;
    final tokens = await _authRepository.restoreSession();
    if (tokens == null) {
      state = const SessionState.unauthenticated();
      return;
    }
    state = SessionState.authenticated(tokens);
  }

  Future<void> login({required String email, required String password}) async {
    try {
      final tokens = await _authRepository.login(
        email: email,
        password: password,
      );
      state = SessionState.authenticated(tokens);
    } catch (e) {
      state = SessionState.unauthenticated(e.toString());
      rethrow;
    }
  }

  Future<void> logout() async {
    await _authRepository.logout();
    state = const SessionState.unauthenticated();
  }
}
