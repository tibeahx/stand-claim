POSTGRES_DSN=postgres://postgres:postgres@db:5435/stands?sslmode=disable
MIGRATION_DOWN_FLAG=-all

migration-up:
	migrate -source file://migrations -database  $(POSTGRES_DSN) up

migration-down:
	migrate -source file://migrations -database  $(POSTGRES_DSN) down $(MIGRATION_DOWN_FLAG)

make config-up:
	if [ ! -f ./config/config.yaml ]; then cp ./config/config.yaml.example ./config/config.yaml; fi