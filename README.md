# Ovoo – Privacy Mail Gateway

Ovoo is your personal email privacy guardian that you can host yourself. Imagine creating unlimited unique email addresses that all forward to your real inbox - without ever revealing your actual email address to anyone. Just like premium email privacy services, but completely under your control.

Whether you're signing up for a newsletter, shopping online, or managing business communications, Ovoo lets you create disposable or permanent email aliases that protect your privacy while ensuring you never miss an important message.

## Overview

Ovoo works with two simple concepts:
* `Protected address` - Your real email address where you want to receive messages
* `Alias` - A randomly generated email address that forwards messages to your protected address. Even when you reply, your real email stays hidden

Setting up Ovoo is straightforward - you just need one server running the Ovoo software and an email server (MTA) to handle the actual email delivery.

You can find example configurations for the Postfix email server in [examples/postfix](./docs/examples/postfix).

## Components

Ovoo has a simple architecture with just two main parts that run from the same program:

### Ovoo API

This is the central brain of Ovoo, providing a REST API service built with Go. It allows you to:
* Create and manage your email aliases that forward to your real inbox
* Control user access through modern authentication methods like OpenID Connect (OIDC), API Keys, or simple username/password
* Use a friendly web interface built with Vue.js to manage everything

Want to integrate with other tools? Check out the full API documentation in [openapi.yaml](./openapi.yaml).

#### REST API Overview

| Endpoints group         | Description                                                                  |
| ----------------------- | ---------------------------------------------------------------------------- |
| /api/v1/aliases         | Allows to manage `Alias` entities for all users                              |
| /api/v1/users           | Allows to manage `User`s of the system (only available to `admin` users)     |
| /api/v1/users/profile   | Retrieves the current authenticated user profile                             |
| /api/v1/users/apitokens | Provides ability to manage API keys for authentication                       |
| /api/v1/praddrs         | Allows managing `Protected address` entities for all users                   |
| /private/api/v1/chains  | Manage email chains identifying each message flow (only user by Ovoo Milter) |

### Ovoo Milter

Ovoo Milter implements [Sendmail Milter](https://www.postfix.org/MILTER_README.html)
protocol and is aimed to use as filtering layer for any MTA supporting this protocol (for now it's only tested
with [Postfix](https://postfix.org))

Ovoo Milter is responsible for receiving emails from MTA and checking if the destination address belongs to the
Ovoo ecosystem, in other words if it can find an `Alias` in the database, it will rewrite incoming email headers to securely forward it to the matching `Protected Address`.

<p align="left">
    <img width="100%" src="./docs/assets/overview/ovoo_overview.svg" alt="Ovoo overview diagram" />
</p>


## Roadmap

- [x] REST API and core mail flow logic
- [x] Milter and integration with Postfix
- [x] WebUI for simple access
- [ ] NoSQL databases support
- [ ] Chrome browser plugin
- [ ] Safari browser plugin
- [ ] IaC for easy deployments to: GCP, AWS, VPS
