module vocabulary/backend/modules/auth

go 1.23

require (
	github.com/golang-jwt/jwt/v5 v5.2.3
	vocabulary/backend/libs/shared v0.0.0
)

replace vocabulary/backend/libs/shared => ../../libs/shared
