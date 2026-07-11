.PHONY: dev build openapi-ui

dev:
	npm run dev

build:
	npm run build

openapi-ui:
	npx openapi-typescript ../coin-api/openapi/v1.yaml -o src/api/schema.d.ts
