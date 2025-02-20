#!/usr/bin/env bash

hurl --very-verbose \
    --variable prot_addr_input="$(./scripts/rnd_email.py)" \
    --variable ext_sender_input="$(./scripts/rnd_email.py)" \
    ./scripts/api_test.hurl
