#!/usr/bin/env bash

hurl -k --very-verbose \
    --variable prot_addr_input="$(./scripts/rnd_email.py)" \
    --variable ext_sender_input="$(./scripts/rnd_email.py)" \
    --user "$(jq -r '.api.default_admin.login' config.json):$(jq -r '.api.default_admin.password' config.json)" \
    ./scripts/api_test.hurl
