module vocabulary/backend/gateway

go 1.23

require (
	github.com/golang-jwt/jwt/v5 v5.2.3
	github.com/jackc/pgx/v5 v5.7.2
	google.golang.org/grpc v1.67.1
	vocabulary/backend/libs/shared v0.0.0
	vocabulary/backend/proto v0.0.0
	vocabulary/backend/modules/notification v0.0.0
	vocabulary/backend/modules/auth v0.0.0
	vocabulary/backend/modules/users v0.0.0
	vocabulary/backend/modules/vocabulary v0.0.0
)

replace vocabulary/backend/libs/shared => ../libs/shared
replace vocabulary/backend/proto => ../proto
replace vocabulary/backend/modules/notification => ../modules/notification
replace vocabulary/backend/modules/auth => ../modules/auth
replace vocabulary/backend/modules/users => ../modules/users
replace vocabulary/backend/modules/vocabulary => ../modules/vocabulary

