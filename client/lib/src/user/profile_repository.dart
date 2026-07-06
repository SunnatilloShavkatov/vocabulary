import 'package:dio/dio.dart';

class UserProfile {
  const UserProfile({required this.id, required this.role, required this.name});

  final String id;
  final String role;
  final String name;

  factory UserProfile.fromJson(Map<String, dynamic> json) {
    return UserProfile(
      id: json['id'] as String? ?? '',
      role: json['role'] as String? ?? '',
      name: json['name'] as String? ?? '',
    );
  }
}

class ProfileRepository {
  ProfileRepository({required this.dio});

  final Dio dio;

  Future<UserProfile> me() async {
    final response = await dio.get<Map<String, dynamic>>('/v1/users/me');
    final body = response.data ?? const <String, dynamic>{};
    return UserProfile.fromJson(body);
  }
}
