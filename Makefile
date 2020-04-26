OPENSSL = /usr/local/opt/openssl@1.1/bin/openssl
KEYS_DIR = ./keys
PRIVATE_KEY = $(KEYS_DIR)/private.pem
PUBLIC_KEY = $(KEYS_DIR)/public.pem

.PHONY: build
build: app.yaml

.PHONY: keys
keys: $(PRIVATE_KEY) $(PUBLIC_KEY)

app.yaml: app.base.json
	cat app.base.json | \
	jq \
		--arg GH_APP_IDENTIFIER "$(GH_APP_IDENTIFIER)" \
		--arg GH_APP_CLIENT_ID "$(GH_APP_CLIENT_ID)" \
		--arg GH_APP_CLIENT_SECRET "$(GH_APP_CLIENT_SECRET)" \
		--arg ADMIN_ORIGIN "$(ADMIN_ORIGIN)" \
		'.env_variables.GH_APP_IDENTIFIER = $$GH_APP_IDENTIFIER | .env_variables.GH_APP_CLIENT_ID = $$GH_APP_CLIENT_ID | .env_variables.GH_APP_CLIENT_SECRET = $$GH_APP_CLIENT_SECRET | .env_variables.ADMIN_ORIGIN = $$ADMIN_ORIGIN' \
	> app.yaml

deploy: app.yaml
	gcloud app deploy --project merge-chance-time

.PHONY: clean
clean:
	rm -f $(PUBLIC_KEY) $(PRIVATE_KEY)

$(PRIVATE_KEY): $(KEYS_DIR)
	$(OPENSSL) genrsa -out $(PRIVATE_KEY) 4096

$(PUBLIC_KEY): $(PRIVATE_KEY) $(KEYS_DIR)
	$(OPENSSL) rsa -in $(PRIVATE_KEY) -pubout -out $(PUBLIC_KEY)

$(KEYS_DIR):
	mkdir -p $(KEYS_DIR)
