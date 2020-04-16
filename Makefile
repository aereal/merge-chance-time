build: app.yaml

app.yaml: app.base.json
	cat app.base.json | \
	jq \
		--arg GH_APP_IDENTIFIER "$(GH_APP_IDENTIFIER)" \
		--arg GH_APP_CLIENT_ID "$(GH_APP_CLIENT_ID)" \
		--arg GH_APP_CLIENT_SECRET "$(GH_APP_CLIENT_SECRET)" \
		'.env_variables.GH_APP_IDENTIFIER = $$GH_APP_IDENTIFIER | .env_variables.GH_APP_CLIENT_ID = $$GH_APP_CLIENT_ID | .env_variables.GH_APP_CLIENT_SECRET = $$GH_APP_CLIENT_SECRET' \
	> app.yaml
