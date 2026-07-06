class SessionTokens {
  const SessionTokens({
    required this.accessToken,
    required this.refreshToken,
    required this.name,
  });

  final String accessToken;
  final String refreshToken;
  final String name;

  bool get hasRefreshToken => refreshToken.trim().isNotEmpty;

  SessionTokens copyWith({
    String? accessToken,
    String? refreshToken,
    String? name,
  }) {
    return SessionTokens(
      accessToken: accessToken ?? this.accessToken,
      refreshToken: refreshToken ?? this.refreshToken,
      name: name ?? this.name,
    );
  }
}
