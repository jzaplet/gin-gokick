# Gin GoKick

Gin GoKick is a lightweight full-stack boilerplate for web apps in Go. One binary serves the API, the server-rendered pages, the Vue 3 SPA and its build. It runs on PostgreSQL and ships with sign-in, languages, emails, a strict CSP, error reporting, metrics and measurement with cookie consent.

**Why it exists**
- **Go keeps hosting cheap.** One small server runs many apps.
- **gin, not a framework of our own.** Solved problems are taken over, not written and maintained again.
- **One skeleton for every project.** The reusable core in `app/core/` takes plain values and never imports the app, so it can later move to its own module. Only its test helpers read the environment and `.env` themselves.
- **Strict rules for people and AI alike.** [AGENTS.md](../AGENTS.md) holds them, and the tools enforce most of them.

This document describes how the parts work and why. AGENTS.md says how to change them.

**Contents**

- [Quick start](#quick-start)
- [Repository](#repository)
- [Configuration](#configuration)
- [Routes and requests](#routes-and-requests)
- [API](#api)
- [Sign-in and sessions](#sign-in-and-sessions)
- [Database](#database)
- [Pages and templates](#pages-and-templates)
- [Emails](#emails)
- [Languages and texts](#languages-and-texts)
- [Frontend](#frontend)
- [Security](#security)
- [Logs, errors and metrics](#logs-errors-and-metrics)
- [Measurement and consent](#measurement-and-consent)
- [Production](#production)
- [Tests](#tests)
- [Checks](#checks)

## Quick start

You need Go (the version of `go.mod`), Node 24 with corepack, and Docker. Local domains like `db.<APP_DOMAIN>` come from OrbStack. Without OrbStack, uncomment the `ports` in `compose.yaml` and set `DB_HOST=127.0.0.1` with `DB_PORT=54320`.

```bash
cp .env.example .env    # APP_DOMAIN, DB_*, DEFAULT_LOCALE…
docker compose up -d db # PostgreSQL 18 on db.<APP_DOMAIN>
corepack enable         # once per machine: yarn 4 from package.json
yarn install
yarn dev                # Vite with HMR, keep it running
go run ./app            # in a second terminal: http://localhost:8020
```

- **Start Vite first.** Go decides at start whether Vite runs (`public/.hot`). Without Vite it needs a build from `yarn build`, or it doesn't start.
- **Open the site through Go**, on `localhost` or `127.0.0.1`, also in development.
- **The git hooks.** `yarn exec lefthook install` installs them once, and [Checks](#checks) says what they run. They run the lefthook of `package.json`, never an older one on the `PATH`. Yarn runs no install scripts, so they never install themselves.
- **The whole app in a container.** `docker compose up -d --build` runs it on `https://<APP_DOMAIN>`.
- **Mailpit.** `docker compose up -d mailpit` catches the mails on `https://mail.<APP_DOMAIN>`. Their links and images lead to `APP_URL`. With `go run`, set `APP_URL=http://localhost:8020` and open Mailpit on `http://mail.<APP_DOMAIN>`, because a page on HTTPS loads no images over HTTP.

## Repository

| Path | What it holds |
| --- | --- |
| `app/main.go` | the start: config, logger, languages and their dictionaries, mail sender, Sentry, database with migrations, router, server with graceful shutdown |
| `app/core/` | the reusable core: HTTP server, middleware, CSP, API envelope, grids, views, page URLs, Vite, static files, locales, ICU, mail, database, password hashing, session cookie and signed-in user, rate limit, logging, Sentry, metrics, health checks, test helpers |
| `app/internal/` | this app, one folder per domain (`auth`, `user`, `home`, `tracking`) plus `shared`: config, router, templates, CSP services, the handlers of no domain, the locale images and the test router |
| `locale/` | the dictionaries, one Go map per locale |
| `views/` | Go templates: the pages, their components and the mails |
| `assets/` | the frontend: Tailwind, Vue screens per domain in `app/`, shared modules in `shared/`, extra entries in `standalone/`, vue-router in `router/` |
| `public/` | the web root, with the Vite build in `public/build` |
| `migrations/` | goose migrations in SQL |
| `tools/` | `tsgen` and `i18n` with their shared `codegen`, `codestyle`, the ESLint plugin `codestyle` in `tools/eslint`, and sqlc pinned in its own module |
| `tests/` | the Vitest tests of the frontend and of the ESLint plugin |
| `docker/` | the production Dockerfile and the script that creates the database role |
| `docs/` | this document, the [first deploy](first-deploy.md) and a Grafana dashboard |

`views/`, `public/` and `migrations/` are embedded into the binary, and `locale/` is compiled into it. `app/main.go` stays under `app/`, because Go lets only packages under `app/` import `app/internal`.

## Configuration

The app reads the environment, and `.env` fills in what the environment lacks. `.env.example` describes every variable with its allowed values and default. A value it can't parse stops the start, and `TestEnvExampleIsValid` checks that the file itself loads.

| Variable | Default | Meaning |
| --- | --- | --- |
| `APP_URL` | required | public address without a path (`https://gokick.dev`); mails build their links and images from it |
| `DEFAULT_LOCALE` | required | default language as a POSIX locale; `/` is always in it |
| `LOCALES` | `DEFAULT_LOCALE` | allowed locales, comma-separated (`cs_CZ,en_US`) |
| `PORT` | `8020` | HTTP port |
| `GIN_MODE` | `release` | `debug`, `release` or `test` |
| `TRUSTED_PROXIES` | empty | IPs and CIDRs trusted for `X-Forwarded-For` |
| `LOG_FORMAT`, `LOG_LEVEL` | `json`, `info` | log format (`json`, `text`) and lowest level (`debug` to `error`) |
| `DB_HOST`, `DB_NAME`, `DB_USERNAME`, `DB_PASSWORD` | required | PostgreSQL server, database, role and password of the app |
| `DB_PORT`, `DB_PARAMS` | `5432`, empty | port, and connection parameters as a query (`?sslmode=disable&pool_max_conns=10`) |
| `DB_TEST_NAME`, `DB_TEST_PARAMS` | empty | test database for `go test`; compose requires `DB_TEST_NAME` and creates it |
| `SMTP_ENABLED` | `false` | `true` sends mails; off, the start logs a warning and reads no other `SMTP_*` |
| `SMTP_HOST`, `SMTP_SENDER_EMAIL` | required when on | SMTP server and sender address |
| `SMTP_PORT` | `587` | SMTP port; port 465 (TLS from the first byte) is refused |
| `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_SENDER_NAME` | empty | SMTP account (both or neither) and sender name |
| `SENTRY_DSN`, `SENTRY_FRONTEND_DSN` | empty | Sentry for the backend and the browser; empty turns each off |
| `SENTRY_ENVIRONMENT`, `SENTRY_RELEASE` | `production`, git revision of the build or `dev` | environment and release in Sentry |
| `METRICS_USER`, `METRICS_PASSWORD` | empty | account for `/metrics`, both or neither; empty makes `/metrics` answer 404 |
| `UMAMI_WEBSITE_ID` | empty | Umami, which needs no consent |
| `GA4_MEASUREMENT_ID`, `GOOGLE_ADS_ID`, `META_PIXEL_ID` | empty | tools that wait for cookie consent |
| `GOOGLE_ADS_CONVERSIONS` | empty | conversion labels (`sign_up=AbC-D_efG`); an unknown conversion stops the start |
| `APP_DOMAIN`, `POSTGRES_USER`, `POSTGRES_PASSWORD` | — | not read by the app: the local domain for `docker compose` (`.env.example` builds `APP_URL`, `DB_HOST`, `SMTP_HOST` and `SMTP_SENDER_EMAIL` from it) and the PostgreSQL superuser |
| `SENTRY_ORG`, `SENTRY_PROJECT`, `SENTRY_REPOSITORY`, `SENTRY_AUTH_TOKEN` | — | only for the Docker build: source maps and the release |

`app/internal/shared/config` puts `Config` together from one package per section of `.env.example`. The exceptions are `TRUSTED_PROXIES`, which the server package reads, and the Sentry build section, which the app never reads.

## Routes and requests

The routes sit in `app/internal/shared/router/routes.go`, and the Vite proxy and the files of `public/` register themselves. `/ping` answers that the app runs, `/readyz` checks the database, and `/metrics` serves Prometheus.

**The language is part of the address.**
- `/` is always `DEFAULT_LOCALE`.
- Every other page has its language as the first segment (`/en`, `/cs/app/login`).
- The home of the default language under its prefix (`/cs`) redirects to `/` with 301 and keeps the query. Its other pages keep the prefix (`/cs/app/login`).
- The SPA is the pattern `/:lang/app` and `/:lang/app/*path` behind `locales.Prefix(notFound)`, which hands a language outside `LOCALES` to the 404.
- The home page of every other language has its own route, because a pattern of one segment would also catch `/favicon.ico`.
- The API, `/ping`, `/readyz`, `/metrics` and the files of `public/` have no language.

**Client IP and limits.**
- gin takes the client IP from Cloudflare's `CF-Connecting-IP` first and trusts it whatever `TRUSTED_PROXIES` says, so the server must be reachable only through Cloudflare.
- Every request gets a 5 s deadline on its context. A query that got `c.Request.Context()` stops with it, but a handler that ignores the context keeps running.
- The server closes a connection that reads or writes longer than 10 s, and an idle keep-alive after 120 s. That is longer than Traefik keeps an idle connection, so Traefik never sends into a closing one.
- On SIGTERM the server stops taking connections and gives running requests 5 s to finish.

**Public files.** `public/` is the web root, like nginx `try_files`. A file there is served at the same path unless an app route takes it.
- Dotfiles, `.go` files and directory listings never go out, and only GET and HEAD are answered.
- The hashed files in `/build/assets/` are cached for a year (`immutable`). Everything else gets `no-cache` with an ETag.
- Every 404 answers `no-store`, except `/metrics` without an account. A client that asks for JSON first gets the envelope, and anyone else the 404 page.

## API

An API handler reads its body with `api.Bind` and answers with `api.JSON`, which sends a nil slice as `[]` and a nil map as `{}`. Gin's own binding is forbidden outside `app/core/api`, and its JSON writers outside `app/core/api` and `app/core/health` (forbidigo, depguard), so no API handler binds or writes JSON past the envelope.

Every error goes out in one envelope. Its keys are the JSON path of a field (`email`, `billing.street`, `items[1].name`), or `general`:

```json
{ "email": { "key": "validation.email" }, "password": { "key": "validation.min_length", "params": { "min": 12 } } }
```

| Status | When | Body |
| --- | --- | --- |
| 400 | the body or the query can't be read | `general`: `request.invalid_body` or `request.invalid_query` |
| 413 | the body is larger than 1 MiB | `general`: `request.body_too_large` |
| 422 | the `binding` tags or the handler refuse fields | one message per field |
| 401, 404, 429 | a handler or middleware refuses the request | `general` with its key |
| 500 | an unexpected error, reported to Sentry | `general`: `request.internal` |

- **Validation messages.** `binding` tags map to message keys: `required`, `email`, `uuid` (also `uuid4`, `uuid7`), `min`, `max` and `oneof`. Every other tag maps to `validation.invalid`.
  - `min` and `max` give the length of a text as a number for a plural (`validation.min_length`), the count of a slice, array or map (`validation.min_items`), and on anything else the limit as a string (`validation.min`).
- **A slice of structs** needs `dive`, or its items go unchecked.
- **Bodies without the envelope.** The CSRF check (403) and recovery (500) answer an empty body, which the frontend turns into a general error.
- **Message keys** are `api.Key` constants, which tsgen collects into `ApiMessageKey` for the frontend.

**Grids.** A grid is a list the frontend pages, sorts and filters, built on `app/core/listing`.
- `api.BindListing` reads `page` (default 1), `perPage` (up to 100, default of the grid), `sortBy` (one of the sorts of the grid), `sortDir` (`asc` or `desc`) and the grid's filters. A value it can't parse (`page=abc`) answers 400, and every other error answers together in one 422.
- The answer is `{ items, total }`. A grid where some rows can't be selected adds `selectable`.
- A page past the end answers empty `items` with the `total`, and the frontend moves back.

## Sign-in and sessions

| Route | What it does |
| --- | --- |
| `POST /api/auth/register` | creates an account from `{ email, password, locale }`, signs it in and sends the welcome mail; 201 `{ id, locale }` |
| `POST /api/auth/login` | signs in with `{ email, password }`; 200 `{ id, locale }` |
| `POST /api/auth/logout` | ends the session and deletes its cookie; 204 |
| `GET /api/auth/session` | `{ id }` of a live session, else 401 |
| `GET /api/auth/sessions` | the grid of the user's live sessions |
| `POST /api/auth/sessions/end` | ends chosen sessions (`ids`, up to 100) or all the filter lists (`all`); `{ ended }` |
| `GET /api/user/me` | `{ id, email }` of the signed-in user |
| `PUT /api/user/locale` | stores the user's language; 204 |

**Passwords.**
- A password has 12 to 128 characters.
- The email is stored in lowercase and compared without regard to case.
- Passwords are hashed with argon2id at OWASP parameters (19 MiB, 2 passes, 1 thread) in the PHC format. When the parameters change, the next login rehashes the password. A hash that fails to store doesn't fail the login; it goes to the log and Sentry as a warning.
- `Verify` refuses a stored hash above 256 MiB or 16 passes.
- A failed login always takes at least 300 ms (`account.FailureFloor`), and an unknown email is hashed too, so the timing doesn't reveal whether an account exists.
- Registration reveals a taken email on purpose, so it isn't delayed.

**Sessions.**
- A session is a random 130-bit token in the cookie `__Host-session` (`HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/`), valid for 30 days.
- The table `sessions` keeps only the token's SHA-256, so a leaked database reveals no session.
- A new session deletes the expired ones. A login or registration ends the session the browser sent, but only after the new one exists.
- If the session doesn't start after the account was created, the registration deletes the account again and answers 500.
- The session stores the sign-in's `User-Agent` (valid UTF-8 without NUL, cut to 512 characters) and the time and IP of the last request. The middleware rewrites the time and IP after 5 minutes, or at once when the address changes, so a stolen cookie used elsewhere shows up immediately.
- An address that doesn't parse is stored as `NULL` and never replaces a known one. An IPv4 address inside IPv6 is stored as IPv4.

**Routes behind the session.**
- They sit in the group with `authn.Middleware`. It finds the owner of the cookie's live session, answers 401 with `auth.sign_in_required` without one, and puts the user ID into the context for the log and Sentry. Neither gets the email.
- A cookie with an unknown token is not deleted, because an older response could otherwise delete the cookie of a newer sign-in from another tab.
- The two session routes work only on the sessions of the signed-in user and never end the current one; only the sign-out and a new sign-in, which replaces it, do. The `ip` filter (up to 45 characters) keeps the sessions whose last address contains the text, in any case and literally, so `%` and `_` stay text.

**Failed attempts.**
- A failed login and a registration with a taken email count as failed attempts. Validation errors don't.
- After 10 in 15 minutes from one network, the login and the registration answer 429 until the oldest attempt leaves the window. A network is an IPv4 /32 or an IPv6 /64.
- The middleware checks the limit before it reads the request.
- The attempts live in PostgreSQL (`auth_failures`), so the limit holds across instances.
- Each limit has its own scope, and a successful login doesn't reset the count.

**The language of the user.** `users.locale` holds it. The registration stores the locale of its form, the login answers it, and `PUT /api/user/locale` changes it. All three pass through `locales.Nearest`, so a locale that left `LOCALES` falls back to its language or to `DEFAULT_LOCALE` and never leads to a 404.

## Database

**Connection and migrations.**
- The app connects to PostgreSQL 18 through pgx. `database.Address` builds the URL from the `DB_*` parts, so a password with `@` or `/` can't break it.
- Without a database the app doesn't start.
- Right after connecting it runs the new goose migrations from `migrations/` under a PostgreSQL lock, so two instances never migrate at once. A failed migration stops the start.

**The app's role.**
- The app connects as a role that owns its database but is no superuser, so it can create tables and run migrations, but no roles or databases.
- `docker/postgres/app-role.sql` creates the role and the databases. Compose runs it at the first start of `db` and CI runs it too; in production it runs once by hand.
- A local change of `DB_NAME`, `DB_TEST_NAME`, `DB_USERNAME` or `DB_PASSWORD` needs `docker compose down -v`, which deletes the data.

**Queries.**
- Queries are SQL in the domain's `db` package. `go tool -modfile=tools/sqlc/go.mod sqlc generate` turns them into typed Go functions according to `sqlc.yaml`.
- The schema comes from the migrations.
- `uuid` becomes `uuid.UUID`, `timestamptz` `time.Time`, `interval` `time.Duration`, `inet` `netip.Addr`, `cidr` `netip.Prefix`, and a nullable column a pointer. A `:many` without rows returns an empty slice, which JSON sends as `[]`.
- The generated files are committed, so a build doesn't need sqlc. sqlc lives in its own module in `tools/sqlc`, so its dependencies stay out of the app.

**Tests with PostgreSQL.**
- `testkit.Database` creates an empty schema in the test database and drops it afterwards. `testkit.Pool(t, migrations.FS)` opens a pool on it with the migrations applied.
- Tests read `DB_HOST`, `DB_PORT`, `DB_USERNAME`, `DB_PASSWORD`, `DB_TEST_NAME` and `DB_TEST_PARAMS`: all from the environment once it has `DB_TEST_NAME`, else all from `.env`.
- Without a database they skip. In CI, where `CI` is set, they fail.

## Pages and templates

**Pages and the layout.**
- The templates in `views/` are parsed at start, so a broken template stops the app.
- A page is named by its path (`home/home.html`) and inherits `shared/layout.html`. It must define the blocks `title` and `description`, usually defines `assets` and `content`, and can add tags through `head`.
- `view.Parse` clones every page for each locale with its own `t` and escapes every clone once at start. A missing block or text that breaks its HTML context stops the start.
- A handler renders a page with `renderer.Page(c, status, "home/home.html", params)`. The template gets:
  - `.Nonce`, `.Locale` and `.Params`;
  - `.Origin` and `.URL` of the page without its query;
  - `.Home`, `.Homes`, `.Alternates` and `.Languages` for the language links.
- `renderer.Page` renders into memory first. On an error it answers 500 without a half-written page and reports it.
- `renderer.Page` sends `Cache-Control: private, no-cache` unless the handler set its own, so no shared cache keeps a page with a nonce.

**Components.** A component in `views/<folder>/components/` shares one set with each page of that folder, so only those pages can call it (`{{template "home/components/intro.html" .}}`). The start refuses a component that defines a template, a folder inside `components/` and components without a page beside them.

**Texts in templates.** Every text comes from `{{t "key" "param" value}}`. An unknown key or params that don't fit their message fail the page with 500 and never show a bare key.

**Head and share card.**
- The layout writes `<html lang>`, the title, the description and a share card: Open Graph and `twitter:card` with the image of the locale (`assets/img/og/<locale>.png`, 1200×630).
- It also writes `<meta name="locale">` with the home of each language for the SPA, and with two or more languages the `hreflang` links.
- `pageurl` builds absolute addresses from `Host` and `X-Forwarded-Proto`.

**The 404.** `sharedhandler.NotFound` answers every 404, except the empty one of `/metrics` without an account: an unknown path, a missing file or an unknown language. The page takes its language from the first path segment and has `noindex` instead of the share card. The SPA draws the same page in `NotFoundView.vue`.

**Vite in templates.**
- `{{vite .Nonce "assets/app.css" …}}` writes the entries with their preloads and nonce.
- `{{asset "…"}}` gives the URL of another built file.
- `{{dictionary .Nonce .Locale}}` preloads the dictionary module of the page.
- A file reaches the manifest only through an import. `assets/app.ts` therefore imports every image a template names, and the flags come through a glob.
- With a build, the start refuses a template that names an entry or file the manifest lacks.

**The SPA shell.** `views/app/app.html` holds a loader that Vue replaces. Its spinner shows after 300 ms, and after 10 s a note with a reload link appears. Both work without JavaScript.

## Emails

**Templates.**
- A mail is a template in `views/mail/` that calls no layout. Its block `subject` is the subject.
- `renderer.Mail(name, locale, origin, params)` renders it. The template gets only `.Locale`, `.Origin` and `.Params`. `.Origin` is `APP_URL`, because a mail may start outside a request.
- The mail logo `assets/img/mail/mark.png` sits on an opaque tile, so it stays visible in dark mode.
- Images under `/build/`, the built ones and those of the Vite proxy, go out with `Cross-Origin-Resource-Policy: cross-origin`, so clients that load them in the browser show them.

**Sending.**
- `mail.Mailer` renders a template in the recipient's locale and sends it within 3 s, through `mail.Sender`.
- `mail.New` returns an SMTP sender, or `mail.Discard` unless `SMTP_ENABLED` is `true`. The SMTP sender writes quoted-printable HTML, always takes the STARTTLS the server offers, and signs in with PLAIN when `SMTP_USERNAME` is set. Away from localhost it sends the password only over an encrypted connection.
- Locally the mails go to Mailpit (`mail.<APP_DOMAIN>`, SMTP on 1025, no account), in production to any SMTP server, such as Amazon SES on 587.

**The welcome mail.** The registration sends it right after the session starts, in the account's language. A mail that fails goes to `reporting.Warning`, and the registration still answers 201.

**Testing mails.** This test sends the welcome mail of every locale to Mailpit, through the `SMTP_*` and `APP_URL` of the environment, else of `.env`:

```bash
go test -tags mailpit -run TestSendSamplesToMailpit ./app/internal/user/mail
```

The test needs a build. Its images come from `APP_URL`, so the app must run there with the current code.

## Languages and texts

**Locales.**
- A locale is POSIX (`en_US`), and the address carries only the language (`en`).
- With several dialects of one language in `LOCALES` (`en_US,en_GB`), `Accept-Language` picks the dialect, related ones included (`en-AU` gets `en_GB`). Every page under a language prefix sends `Vary: Accept-Language`.
- The default language has one dialect, because `/` picks none.
- `locale.New` refuses an unknown or deprecated code, a duplicate, and a default outside the list.

**Dictionaries.**
- Texts live in `locale/`, one Go map per locale (`locale/cs_CZ.go`). Each file registers itself in a one-line `init`.
- The start refuses a locale without a dictionary. Against a build it also refuses a locale without its dictionary module or share image and, with two or more languages, a language whose first dialect has no flag (`assets/img/flags/<country>.svg`).

**ICU MessageFormat.** Messages use a subset, parsed by `app/core/i18n/icu`:
- `{arg}` for a text;
- `select` on a text;
- `plural` on a number with `#`, `=N` and `offset`.

A number is a signed integer, a `float64` (up to 3 decimals) or an `icu.Decimal` with a fixed count of 0 to 15 decimals. Integer parts above 2^53−1 are refused.
- **What the parser refuses.**
  - Every other argument type.
  - Constructs that libraries read differently: an apostrophe before `<` or `>`, an unclosed quote, a `}` without a pair, repeated cases, unknown plural categories and `=01`.
  - Bad names: numeric names, names outside ASCII letters, digits and `_`, one name used in two ways, and the `select` case `__proto__`.
- **Rounding.** A plural rounds like the browser, halves away from zero. Its category follows the number as shown, so `1,0` is `many` in Czech.

**`go run ./tools/i18n`** checks the dictionaries and writes them for the frontend. It refuses:
- a key that isn't lowercase words joined by dots, or that a dictionary lacks;
- a message outside the subset, or whose arguments differ in name or kind between the dictionaries;
- a plural without every CLDR category of its language, or with one the language never picks;
- in Czech, a vocalizing preposition (s/se, k/ke, v/ve, z/ze) right before `#`, since only the number decides between „s 5“ and „se 7“;
- in `views/`:
  - a `t` whose key is no literal or is unknown, or whose params don't fit;
  - text outside `t`, between tags and in the text attributes (`title`, `alt`, `placeholder`, `aria-label`…);
- in the Go code, an `api.Key` that doesn't come from a constant of that type, a constant without a text, and one declared inside a function;
- a dictionary without a `language.name`;
- a key nothing uses. A key used only in `tests/` counts as unused;
- in `assets/`, a key built at runtime (`` `tracking.${tool}` ``), and any `.tsx` or `.jsx` file.

The tool refuses a key nothing uses before it writes anything. It reads the frontend with its own lexer, which needs no types, so a use can come before them. The tool writes into `assets/shared/I18n/Dictionary`:
- the types `MessageKey` and `MessageParams`;
- one module per locale;
- the loaders, with the name of each locale's language (`language.name`) and a version that changes with its module.

It also writes a sample of numbers to `tests/assets/i18n/numbers`, against which a Vitest test checks that `Intl` formats every locale like Go.

## Frontend

**Build.** Vite builds Tailwind 4 and Vue 3 in TypeScript, managed by yarn 4 through corepack.
- **Entries.** They are `assets/app.css`, `assets/app.ts` and every `.ts` file in `assets/standalone/`.
- **The brand.** It lives in the `@theme` of `assets/app.css`: the colors `brand-*` and `ink-*`, Tailwind's `slate-*` for neutrals, and Plus Jakarta Sans served from the app's own files (Latin and Latin Extended).
- **The brand files.** `logo.svg` and `logo-dark.svg`, `mark.svg` and `mark-dark.svg`, `mail/mark.png` and `og/<locale>.png` in `assets/img/`, and `public/favicon.ico`, are Gin GoKick's. A project replaces them with its own. The dark logo serves the root README in the dark theme of GitHub.
- **Production.** `yarn build` writes hashed files and `manifest.json` to `public/build`, which Go embeds.
  - The build never inlines a font, an SVG or a PNG as `data:`.
  - It renames files to names `go:embed` accepts.
  - It deletes a half-written `public/build` when it fails.
- **Development.**
  - `yarn dev` writes `public/.hot`. A Go started after it (it reads the file only at start) writes the dev sources and proxies `/build/`, the HMR websocket included, to Vite for local clients only. The browser talks only to Go, so the CSP is the same as in production.
  - After a restart of `yarn dev`, reload the page by hand.

**Structure.**
- `assets/app/<Domain>/` holds the screens of a domain, and `shared/` everything shared: Fetch, Auth, I18n, Grid, the form fields, buttons, modal, toasts, icons and the rest.
- The routes in `assets/router/routes.ts` are flat. Each writes its whole path (`/:lang/app/login`) and its `meta`: `access` (`Guest`, `User`, `Public`), `title` and `description`.
- A guard sends a signed-in user from a `Guest` route to the dashboard and a signed-out one from a `User` route to the login, without calling the API.
- After each navigation, the title and description of the page are written for SEO and measurement. A navigation that only changes the query of a grid is skipped.

**Layouts.**
- `App.vue` picks the layout by `access`: `PublicLayout` for the login, the registration and the 404, `AuthLayout` for the dashboard.
- `AuthLayout` loads the user from `/api/user/me` before it renders a screen. Until then it shows a spinner, and it shows a retry button when the load fails.

**Session in the browser.**
- `shared/Auth/Session` keeps only the user's ID, in localStorage, and gives Sentry the same ID.
- At start `restoreSession` checks it once with `/api/auth/session`, and other tabs follow a sign-in or sign-out through the `storage` event.
- `completeLogin` opens the target from `redirect` (only a path inside the SPA) in the account's language, without a reload.

**Calling the API.**
- `apiFetch` from `@/shared/Fetch` answers `{ success, status, data }`. A response that fails its guard becomes a general error and goes to Sentry, because it means the frontend and the API disagree.
- `authFetch` has the same signature. On a 401 it forgets the session and opens the login with the page in `redirect`.
- The sign-in routes stay on `apiFetch`, since a wrong password also answers 401.

**Types from Go.**
- `go run ./tools/tsgen` writes a TypeScript type and a guard for every Go type marked `//tsgen:<path> <Name>`. The options are:
  - `request`, for a body the frontend sends: the type plus its field errors;
  - `noguard`;
  - `union`, for a string type whose constants become a const object.
- It covers text, numbers, `bool`, `time.Time`, `uuid.UUID` and `netip.Addr` as text, slices, pointers (`| null`), `omitempty` and `omitzero` (optional), `map[string]any`, and other marked types. For anything else it fails rather than write `any`.

**Texts in the browser.**
- The page loads its own dictionary at start, and another one when the visitor switches. The switch loads it on hover or focus already.
- A dictionary is cached in localStorage under its version.
- `t('key', params)` takes exactly the params of its key, so vue-tsc refuses an unknown key or a wrong param, in templates too.
- `formatMessage` fills a message like Go's `icu.Formatter`, through `Intl`. A missing text or a param that doesn't fit is reported to Sentry, and the page shows the key.

**The language switcher.**
- It shows the flag of the page and the languages of `LOCALES`, each in its own name.
- In the SPA it switches without a reload. On the screens behind the sign-in it also stores the language.

**Shared pieces.**
- The form fields (`BaseInput`, `BaseNumberInput`, `BaseNullableInput`, `BaseSelect`, `BaseCheckbox`, `BaseSwitch`) and `BaseButton`.
- `ConfirmModal`, a native `<dialog>` that fades in and out, in Firefox too.
- Toasts with their keys, so they follow a language switch.
- `BaseDropdown` and `BaseTooltip`.
- The icons: SVG files drawn through `<use>`.
- The Go pages draw the same dropdown and icons, and their dropdowns open without JavaScript.

**The grid.** `useGrid` with `DataGrid` and `GridPagination` loads a page through `authFetch`.
- It sorts by a click on a header.
- It filters 400 ms after the last keystroke.
- It keeps its page, sort and filters in the address under its name (`?sessionsPage=2&sessionsIp=203.0`), so a link restores it.
- It selects rows across pages or "all the filter lists" for bulk actions.
- Only the answer to the last request counts. A failed load keeps the rows shown.

**The Go home page.**
- It shows the project with an animated mark and moving feature cards (`assets/standalone/showcase.ts`).
- `PublicLayout` and the 404 draw drifting dots (`constellation.ts`).
- Both stay still for a visitor who asked for less motion, and both can go away in one piece.

## Security

**Headers.** Every response carries these, the 404 and the 500 included:

| Header | Value |
| --- | --- |
| `Content-Security-Policy` | see below |
| `Strict-Transport-Security` | `max-age=63072000; includeSubDomains; preload` |
| `Cross-Origin-Opener-Policy` | `same-origin` |
| `Cross-Origin-Resource-Policy` | `same-origin`; `cross-origin` for the images under `/build/`, for mails |
| `Permissions-Policy` | no camera, microphone, geolocation, screen sharing, payment, USB or sensors |
| `Referrer-Policy` | `strict-origin-when-cross-origin` |
| `X-Content-Type-Options` | `nosniff` |
| `X-Frame-Options` | `DENY` |

**Content Security Policy.** `app/core/csp` builds it from `default-src 'none'`:
- **Scripts** run only with the nonce of the request. `'strict-dynamic'` lets them load others.
- **Styles** come from the own domain or a `<style>` with the nonce, and fonts only from the own domain.
- **Images and connections** may go to any HTTPS host. Google measures through the national domain of each visitor, and CSP has no wildcard for a top-level domain.
- **Trusted Types** (`require-trusted-types-for 'script'`) block text written into `innerHTML` or `script.src` without a named policy.
  - Vue writes through its policy `vue`.
  - The script loader writes through `script-loader`, which lets through only the `ScriptURL` constants from Go.
  - `v-html` and `innerHTML` bindings are left to ESLint, which refuses them.
- **Services.** `Policy` in `app/internal/shared/csp` allows Turnstile and Vue always, and the hosts of a measurement tool only when its ID is set.

**CSRF.** `middleware.CSRF()` refuses a cross-origin POST, PUT, PATCH or DELETE with 403, through `http.CrossOriginProtection`, so forms need no tokens. It lets GET, HEAD and OPTIONS through.

## Logs, errors and metrics

**Logs.**
- The app logs through one `*slog.Logger` to stdout, as JSON in production.
- Every request gets a `trace_id`, taken from the frontend's `sentry-trace` or generated, and returned in `X-Trace-Id`. A route behind the session adds `user_id`.
- **The access log** writes the method, the route pattern in `path`, the status, the IP, the size, `duration_ms`, the reported errors and the warnings. A token in the path of a matched route stays out. A request no route matched (an unknown path, a public file) logs its real path. The query is never logged.
  - 5xx and panics log at `error`, and a warning without 5xx at `warn`.
- **Lint keeps one path.** It forbids:
  - `slog.New` outside `app/core/logging`;
  - the global logger and the `log` package;
  - `fmt.Print` and the `print` builtins;
  - `os.Stdout` and `os.Stderr` outside the `main.go` files;
  - gin's logger and recovery;
  - logs without a context;
  - keys without a constant.

**Sentry.**
- `SENTRY_DSN` turns it on. Every panic and every 5xx with a reported error goes there, and several errors of one request form one event.
- The event carries:
  - the route pattern, the `User-Agent` and the query;
  - masked `Authorization` and `Cookie`;
  - the user ID.

  It carries no body, no cookies and no IP.
- `reporting.Error(c, err)` keeps the stack of its call, so Sentry shows the handler line. `reporting.Warning(c, err)` reports an error the request survives.
- **In the browser** `SENTRY_FRONTEND_DSN` starts `@sentry/vue` in the SPA.
  - It reports uncaught and Vue errors, and every CSP violation outside browser extensions as a warning.
  - It drops the fragment from every URL.
  - It sends `sentry-trace` and `baggage` with each own-origin `fetch`, so a backend error joins the frontend's trace.
  - It measures no performance and never infers the visitor's IP.
- **Source maps** are uploaded by the Docker build, only with the build secret `SENTRY_AUTH_TOKEN`, and then deleted from the image.

**Metrics.** `/metrics` serves Prometheus on the app's port behind basic auth (`METRICS_USER`, `METRICS_PASSWORD`). Without an account it answers an empty 404.

| Metric | Type | Labels |
| --- | --- | --- |
| `http_requests_total` | counter | `method`, `route`, `status` |
| `http_request_duration_seconds` | histogram | `method`, `route` |
| `http_requests_in_flight` | gauge | none |
| `go_*`, `process_*` | runtime of Go and of the process | — |

`route` is the route pattern, so tokens stay out and the series stay few. Unknown paths and public files count as `unmatched`.

**Health.** `/readyz` checks the database within 2 s and answers 200, or 503 with the state of each check without the error text.

## Measurement and consent

Each tool is switched on by its ID in the environment and runs without Google Tag Manager, so the CSP stays strict and events are called from code.

- **Umami** runs without cookies, so it needs no consent. The layout loads its tag with the nonce on every page, and the fragment of the address stays out.
- **GA4, Google Ads and Meta Pixel** set cookies, so they wait for consent.
  - The layout then writes their IDs and the Ads labels into `<meta name="tracking">` and loads `assets/standalone/consent.ts`.
  - The banner offers "accept all", "only necessary" and detailed settings with one switch per category of the enabled tools, all off at first.
  - The choice is stored with the tools and their IDs, and it is asked again when one of them changes.
  - A cookie settings button appears in the footers whenever one of them is on. A withdrawn consent reloads the page.
- **Google.** GA4 and Ads share one gtag.js, loaded after the first grant.
  - Consent Mode starts all denied and grants only the categories of Google's own tools.
  - Page views are sent by the router, the same address twice in a row once.
  - The GA4 stream must keep "Page changes based on browser history events" off.
- **Meta Pixel** queues `init` and `PageView` until fbevents.js loads, and counts SPA pages itself. Meta reads the address with its fragment.
- **Custom events** are listed once in Go (`app/internal/tracking/events`). `trackedEvents` maps each to its GA4 name, Ads conversion and Meta event, and `trackEvent` sends it only to tools that have consent and run.
  - The sign-up sends the email hashed with SHA-256 to Ads and Meta, and never in plain text.
  - In Google Ads, the conversion needs "Enhanced conversions" on.
  - The privacy policy must mention it.
- **Queries reach every tool.** The query of an address goes to every tool and to Sentry.
- **Hotjar** is not supported, because it can't run under Trusted Types.

## Production

**The Docker image.** `docker/production/Dockerfile` builds the frontend in `node:24-alpine` and the binary in `golang:1.26-alpine`, and runs it on Alpine as the user `gokick` (uid 1000) on port 8020.
- The build needs the `.git` folder, a shallow clone is enough, because the binary carries the commit as its Sentry release.
- Its `assets` stage copies only what the frontend build needs.
- The healthcheck asks `/readyz` every 30 s.

**Dokploy.** The app runs behind Traefik in Dokploy. Traefik passes `CF-Connecting-IP` unchanged. For requests without it, set `TRUSTED_PROXIES=127.0.0.1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16` so gin skips Traefik's address.

**The first deploy.** [first-deploy.md](first-deploy.md) walks through it: Cloudflare, Sentry, PostgreSQL, Dokploy, Prometheus, Loki, Grafana and the checks.

## Tests

A test catches a failure nobody would notice before it does harm. [AGENTS.md](../AGENTS.md#tests) holds the rules, and four questions stand behind them:
- **Impact.** What breaks with the behaviour. Security, consent and personal data, stored data, the API contract and the start weigh most, looks and demo content least.
- **Silence.** Whether anyone sees the failure before it does harm. What shows at once gets fixed at once.
- **Uniqueness.** Whether a tool of [Checks](#checks), a check at start or another test catches it already. A check at start counts only when a test runs it, as `TestNewRouterLoadsTheBuiltAssets` runs the templates against the build.
- **Cost.** How often an unrelated change breaks the test: classes, markup, copy, order, restated constants and lists kept by hand.

So the colors and sizes of `BaseButton` have no test, because they show at once. Its busy state has one: the locked button prevents a double submit, and nobody sees `aria-busy`.

**Go.**
- `testkit` is the facade of the shared helpers: requests and responses (`Request`, `JSONRequest`, `Serve`, `Markup`, `MarkupOf`, `Cookie`, `Nonce`), logs (`DiscardLogs`, `CaptureLogs`, `LogEntries`), `Mailbox`, `.env` (`LoadEnv`), the test database (`Database`, `Pool`, `Count`, see [Database](#database)) and a fake build (`Public`, `TemplateLiterals`).
- `routertest` builds the router through `router.NewRouter` with defaults for a test: the test mode of gin, a discarded log, Czech alone, a fake build of the files the templates and the languages name, and a sender that discards the mails. `Config` returns the default config, `Locales` builds a set of languages, `Files` and `Public` give the file names and the fake build for such a set, and `Renderer` renders the templates of a mail.
- The tests of `tools/` are the specification of the tools. `tsgen` and `i18n` compare their output with golden files in `testdata/`.

**Vitest.**
- The tests run in Vitest with jsdom (`yarn test`). The setup files add what jsdom lacks for `<dialog>`, and before every test they show the Czech dictionary and set the page to Czech alone.
- Routed components mount with the app's real router.
- The rules of the ESLint plugin `codestyle` and the ban of `innerHTML` are tested in `tests/tools/eslint/`.

## Checks

| Command | What it checks |
| --- | --- |
| `yarn install --immutable` | `yarn.lock` matches `package.json` |
| `go run ./tools/tsgen` | TypeScript types and guards match the Go types |
| `go run ./tools/i18n` | dictionaries, texts of the templates and the code, unused keys, generated TypeScript |
| `go run ./tools/codestyle` | no gofmt columns, split keyed literals, no `!`, layout of the templates |
| `go tool -modfile=tools/sqlc/go.mod sqlc diff` | generated sqlc code matches the queries |
| `golangci-lint run --fix ./...` | lint and format by `.golangci.yml` (gofumpt, goimports, the blank lines of wsl) |
| `go vet ./...`, `CI=1 go test -race ./...`, `go build ./...` | the Go code; `CI=1` makes a test without its database or the Vite build fail instead of skipping |
| `go mod tidy -diff` | `go.mod` and `go.sum` hold only what the code needs |
| `yarn lint`, `yarn typecheck`, `yarn test`, `yarn build` | ESLint, vue-tsc, Vitest, the production build |
| `docker build -f docker/production/Dockerfile .` | the production image builds |
| `yarn exec commitlint --from origin/main` | commit messages |

`codestyle` and golangci-lint with `--fix` rewrite what they can fix and report the rest. wsl runs without `decl`, whose fix joins `var` lines into a block that gofmt aligns into columns. `tsgen` and `i18n` write their files only when every check passes. CI runs the three tools with `-check` and golangci-lint without `--fix`, so it only reports.

**Git hooks** (`lefthook.yml`):
- **pre-commit** runs the commands of the table except the tests, the builds and commitlint, and stops at the first failure.
  - Each runs only when the commit changes the files it reads (vue-tsc also after a change of Go, which `tsgen` turns into TypeScript), and ESLint lints only the staged files.
  - The fixes of the staged files and the files `tsgen` and `i18n` write join the commit.
  - A commit of named paths (`git commit -- <path>`, as JetBrains IDEs commit) gets the fixes too, but the index keeps such a file unfixed until the next `git add`.
  - A code file staged only in part stops the hook. Lefthook hides its unstaged lines during the hook and puts them back by their line numbers, which a fix would shift.
  - golangci-lint keeps its cache in the git directory of the checkout ([Worktrees](../AGENTS.md#workflow)).
- **pre-push** checks the branch name and the commit messages.

**CI** (`.github/workflows/`) runs four jobs:
- **Go**: the Vite build first, then `go mod tidy -diff`, `sqlc diff`, the tools with `-check`, `go vet`, lint, tests on a PostgreSQL service and `go build`;
- **Frontend**;
- **Docker image**, so a broken Dockerfile shows in the pull request;
- **Commit messages**, on pull requests only.
