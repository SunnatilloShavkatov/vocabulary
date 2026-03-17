module vocabulary/backend/modules/vocabulary

go 1.23

require (
	github.com/jackc/pgx/v5 v5.7.2
	vocabulary/backend/libs/shared v0.0.0
)

replace vocabulary/backend/libs/shared => ../../libs/shared
