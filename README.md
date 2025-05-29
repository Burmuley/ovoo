# Ovoo – Privacy Mail Gateway

Ovoo is a self-hosted privacy mail gateway that enables users to deploy a secure and private email forwarding service on their own DNS domain. It allows you to manage disposable or long-term email aliases that forward to one or more personal mailboxes—keeping your real email address private.

Designed for individuals, small businesses, and communities, Ovoo empowers users to take control of their email privacy without relying on third-party services.

## Overview

The Ovoo operates a few major entities:
* `Protected address` - is an email address configured by a user aimed to receive incoming email
* `Alias` - a randomly generated email configured to forward all incoming emails to one of the Protected addresses keeping it hidden even if you reply to an incoming email

To operate Ovoo in a minimal setup only need to have one host running both Ovoo components along with an MTA
repsonsinble for handling SMTP data transfers.

Example Postfix MTA configuration you can find in [examples/postfix](./docs/examples/postfix) subdir.

## Components

Architecture of the Ovoo is quite simple and only involves two components that is running from the same binary.

### Ovoo API

This component is a simple REST API service implemented in Go and plays the core role in the Ovoo ecosystem.

You need Ovoo API to:
* Create and manage email aliases mapped for forwarding to real mailboxes
* Manage users and access to the API with the following supported options: OpenID Connect (OIDC), API Keys
or the simplest one - HTTP Basic Authentication
* Provide a user-friendly web UI built with Vue.js for your users (or yourself only)

The OpenAPI definition you can find in the [openapi.yaml](./openapi.yaml) file in the root of the repository.

<p align="left">
    <img width="100%" src="./docs/assets/overview/ovoo_overview.svg" alt="Ovoo overview diagram" />
</p>

## REST API Overview

| Endpoints group         | Description                                                                  |
| ----------------------- | ---------------------------------------------------------------------------- |
| /api/v1/aliases         | Allows to manage `Alias` entities for all users                              |
| /api/v1/users           | Allows to manage `User`s of the system (only available to `admin` users)     |
| /api/v1/users/profile   | Retrieves the current authenticated user profile                             |
| /api/v1/users/apitokens | Provides ability to manage API keys for authentication                       |
| /api/v1/praddrs         | Allows managing `Protected address` entities for all users                   |
| /private/api/v1/chains  | Manage email chains identifying each message flow (only user by Ovoo Milter) |

### Ovoo Milter

Ovoo Milter on the other hand implements [Sendmail Milter](https://www.postfix.org/MILTER_README.html)
protocol and is aimed to use as filtering layer for any MTA supporting this protocol (for now it's only tested
with [Postfix](https://postfix.org))

Ovoo Milter is responsible for receiving emails from MTA and checking if the destination address belongs to the
Ovoo ecosystem, in other words if it can find an `Alias` in the database, it will rewrite incoming email headers to securely forward it to the matching `Protected Address`.

## Roadmap

- [x] REST API and core mail flow logic
- [x] Milter and integration with Postfix
- [x] WebUI for simple access
- [ ] NoSQL databases support
- [ ] Chrome browser plugin
- [ ] Safari browser plugin
- [ ] IaC for easy deployments to: GCP, AWS, VPS
