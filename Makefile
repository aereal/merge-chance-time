build: app.yaml

app.yaml: app.base.json
	cat app.base.json | \
	jq \
		--arg GITHUB_APP_IDENTIFIER "$(GITHUB_APP_IDENTIFIER)" \
		--arg GITHUB_APP_CLIENT_ID "$(GITHUB_APP_CLIENT_ID)" \
		--arg GITHUB_APP_CLIENT_SECRET "$(GITHUB_APP_CLIENT_SECRET)" \
		'.env_variables.GITHUB_APP_IDENTIFIER = $$GITHUB_APP_IDENTIFIER | .env_variables.GITHUB_APP_CLIENT_ID = $$GITHUB_APP_CLIENT_ID | .env_variables.GITHUB_APP_CLIENT_SECRET = $$GITHUB_APP_CLIENT_SECRET' \
	> app.yaml
