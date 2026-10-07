# AGENTS.md

Gin GoKick is a full-stack Go boilerplate:
- a gin backend in `app/`;
- Vue 3 and Tailwind 4 in `assets/`;
- Go templates in `views/`;
- PostgreSQL through pgx, goose and sqlc.

The Go module `gokick` sits at the repository root.

This file says how to change the code. [docs/README.md](docs/README.md) says how each part works and why, so read its chapter before you change a part. Keep each fact in one of the two files. A rule goes here, and a description goes there.

## Workflow

- **Branches.**
  - A change is one pull request on its own branch from the current `origin/main`.
  - A branch name starts with `feat/`, `fix/`, `refactor/`, `perf/`, `docs/` or `test/` (`refactor/core-ping-metrics`). The pre-push hook refuses any other name except `main`. A tool that names the branch `feature/…` needs a `git branch -m` before the push.
  - A follow-up fix goes to the branch of the open pull request. Don't create helper branches, and don't stack a branch on an unmerged one.
  - Never rename a branch with an open pull request, because GitHub closes it.
- **Commits.** Commit messages follow Conventional Commits with the rules of `commitlint.config.ts`.
  - `yarn exec commitlint --from origin/main` checks a branch.
  - Lefthook runs the same check before every push, once installed as [Quick start](docs/README.md#quick-start) says.
  - CI checks every pull request.
- **Merging and deploying.** The maintainer reviews and merges, so don't merge. Don't deploy either: a merge into `main` deploys nothing, and the maintainer deploys by hand.
- **Worktrees.** Work in the checkout you were given, and make a git worktree only when asked. Inside a worktree, run golangci-lint with its own cache (`GOLANGCI_LINT_CACHE=<dir>`), or it reports the paths of the main checkout and `--fix` rewrites its files.

### Before a commit

```bash
yarn install --immutable &&
  go run ./tools/tsgen &&
  go run ./tools/i18n &&
  yarn lint &&
  yarn typecheck &&
  yarn test &&
  yarn build

go run ./tools/codestyle &&
  go tool -modfile=tools/sqlc/go.mod sqlc diff &&
  golangci-lint run --fix ./... &&
  go vet ./... &&
  CI=1 go test -race ./... &&
  go build ./... &&
  go mod tidy -diff

docker build -f docker/production/Dockerfile . &&
  yarn exec commitlint --from origin/main
```

- **The same as CI.** The block runs what the jobs of [CI](docs/README.md#checks) run. golangci-lint has to be the version `validate.yml` pins.
- **Order.** `yarn build` runs before the Go block. Go embeds `public/build`, and `TestNewRouterLoadsTheBuiltAssets` checks the templates against it.
- **Generators and fixes.** `tsgen`, `i18n` and `codestyle` are programs in `tools/`.
  - Run them without flags and golangci-lint with `--fix`, and commit what they write, because CI only checks.
  - No Go test runs them over the module, so run them yourself.
- **`CI=1` for the Go tests.** The block runs them with `CI=1`, as CI does. [Checks](docs/README.md#checks) says what it changes, and [Database](docs/README.md#database) where the tests find their database.

## Ground rules

- **Strict tooling stays strict.** Never loosen a lint, type or Go config to get through. That means no file exception, no `off` override, no disable comment and no wider allow-list.
  - Find a design that keeps the rule. Move the code where the rule already allows it, or build a factory there (`createApiFetch` in `shared/Fetch` for `authFetch`).
  - If no such design exists, ask the maintainer first. List every loosening in the pull request.
- **No comments**, unless the next reader would get the code wrong without one and neither a name nor a test can say it. That means no doc comments, no package comments and no comments in config files. There are three exceptions:
  - `.env.example`, where every variable has its section, allowed values, default and a short explanation;
  - the commented-out `ports` of `compose.yaml`, which a machine without OrbStack uncomments;
  - a page drawn twice, once in a Go template and once in Vue. Each file starts with a comment that names the other. Today these are `views/errors/not_found.html` with `NotFoundView.vue`, `views/shared/languages.html` with `shared/I18n/Components/LanguageSwitcher.vue`, and `views/shared/footer.html` with `shared/Layout/Components/AppFooter.vue`.
- **No logical negation** in Go or TypeScript. Compare explicitly instead:
  - `== false` in Go;
  - `=== false` for a boolean, `!== true` for one that may be missing, `=== undefined` or `=== null` for a missing value in TypeScript.

  A condition that only guards the rest of its block is written positively, with an early `return` or `continue`. `go run ./tools/codestyle` rewrites `!` in Go, and ESLint refuses it in scripts and templates.
- **No language and no project name in code.** `DEFAULT_LOCALE` is required, and `LOCALES` defaults to it. A test may name its own locales.
  - Before a commit, grep the code outside the tests for `cs_CZ`, `cs-CZ`, `'cs'`, `"cs"` and `/cs/`. The expected hits are `.env.example`, the dictionary `locale/cs_CZ.go` with the generated loaders `shared/I18n/Dictionary/dictionaries.ts`, the test router `routertest.go`, and the Czech preposition rule of `tools/i18n`.
  - Names of the project stay the neutral `gokick` (`gokick.local`, the database `gokick`).
- **Every text comes from the dictionaries** in `locale/`:
  - `{{t "…"}}` in a template;
  - `t('…')` in Vue;
  - an `api.Key` constant in an API error.

  A key is written whole, never built at runtime. A key per value goes into a map, like `toolNames` in `useCookieConsent`.
- **Configuration is typed code.** Configuration the frontend reads (a form, a survey) is TypeScript with literal types. JSON is only a generated file with a test against drift.
- **Keep the backend light.** A new Go dependency needs an entry in the depguard allow-list of `.golangci.yml`. A dependency only a tool needs also gets a depguard rule that denies it everywhere else (`golang.org/x/net/html`).
- **Secrets stay out of the query.** A token in a link goes into the fragment, and the page that reads it removes it from the address first ([Measurement and consent](docs/README.md#measurement-and-consent) says why).
- **A test guards a failure that would go unnoticed and do harm.** [Tests](#tests) says how to write one.
  - A failure that the first load, the first click, the log, the start, `vue-tsc` or ESLint shows at once needs no test, and neither does one that harms nothing.
  - Unnoticed is what nobody tries by hand or sees in their own setup: keyboard and screen readers, error and network paths, the dark scheme, reduced motion, blocked storage, a single language, a production build, mails.
  - Tested even when the failure shows or harms nothing: security, consent before any measurement, personal data sent to other hosts, the API contract (status, envelope, error keys, field errors), the checks at start and those this file describes, the tools in `tools/` and the rules of the ESLint plugin `codestyle`.
  - Coverage alone earns no test, and neither do a library's own behaviour (gin, `html/template`, pgx, `v-model`) and a state the types or the wiring rule out. Code without logic of its own, like a wrapper that only passes its arguments on or a constant, is checked through its effect in the test of its user.

## Go

- **`app/main.go` only starts.** It reads the config, builds the parts and passes them on.
  - A decision, like which mail sender to use, belongs to a package (`mail.New`).
  - Objects built from the config are built in `main.go` and passed to `router.NewRouter` as parameters. These are the logger, the reporter, the Sentry options of the browser, the pool, the languages, the public files and the mail sender.
- **`app/core/` is the reusable core.** It takes plain values, never the app config, and never imports `app/internal/` (depguard). A big core package becomes a facade over subpackages, like `vite` (`manifest`, `devserver`, `tags`), `locale` and `reporting`.
- **`app/internal/<domain>/` holds only packages:**
  - `handler`, when the domain has routes. It holds one file per route and, when the route needs one, its test, nothing else; the helpers of the tests live in the test files.
  - `db`, when the domain has tables. It holds the SQL and the code sqlc generates from it, and every importer names it `<domain>db`.
  - One package per job of the domain (`auth/session`, `user/account`), up to about 200 lines; a bigger one splits into subpackages. Such a package:
    - takes plain values;
    - returns sentinel errors (`account.ErrLoginFailed`);
    - wraps every other error with what failed;
    - imports neither gin nor `app/core/api` (depguard).
  - What belongs to no domain sits in `app/internal/shared` ([Repository](docs/README.md#repository)).
- **Handlers.**
  - A handler is a function of what it needs (`homehandler.Home(renderer)`, `authhandler.Login(pool, locales)`). It builds the domain's packages inside its closure.
  - A handler only orchestrates. It binds the request, calls the domain, turns the sentinel errors into the envelope, and writes cookies and the status. A blank line separates its steps.
  - A link that leaves the response (a mail, a stored or shared URL) starts with `APP_URL`, never with the request's `Host`, which a client can forge. `.Origin` from the request is only for the page itself, sent `private, no-cache`.
- **Errors and logs.**
  - An unexpected error goes to `reporting.Error(c, err)`, and the handler returns. Never use `c.Error` (forbidigo).
  - An error the request survives goes to `reporting.Warning(c, err)`.
  - Every reported problem reaches the injected log too, not only Sentry, so it shows where `SENTRY_DSN` is empty. A new reporting path keeps that.
  - Log only through the injected `*slog.Logger` with a context (`InfoContext`, `LogAttrs`), with typed `slog.Attr` values and the key constants of `app/core/logging`.
- **Undo.** When a later step of a handler fails, undo the earlier one through that domain's own function (`accounts.Delete` in `authhandler.Register`).
  - Run the undo on `context.WithoutCancel` with its own timeout.
  - Report both errors, the failed step first.
  - No transaction stays open across a password hash.
- **Routes in `app/internal/shared/router/routes.go`.**
  - Every route has its own line with its full path, so a search for the path finds it.
  - An API path starts with its domain (`/api/auth/login`).
  - A GET, HEAD or OPTIONS route changes nothing a foreign site could abuse ([CSRF](docs/README.md#security)).
  - A group only adds middleware (`router.Group("", api.InternalErrors())`), never a prefix.
  - When gin can't match a list of values, use a pattern with a middleware that checks the parameter (`/:lang/app/*path` behind `locales.Prefix(notFound)`). A loop is only for what a pattern would catch too widely: the home page of each language.
  - Before you design around a gin limit, try the behaviour in a small probe program, and show the maintainer the variants with what each costs.
- **Formatting.**
  - No gofmt columns.
  - A keyed literal stays on one line until it is wider than 120 columns. Then `go run ./tools/codestyle` gives each field its own line, with a blank line between them.
  - Struct fields and consts are split by blank lines or share one type.
  - Blank lines between statements follow wsl, and `golangci-lint run --fix ./...` writes them:
    - a statement sits right above an `if`, a loop, a `switch` or a call only when they share a variable, and an assignment only among assignments;
    - a blank line follows a block statement, `defer` and `go` unless it ends its block or `case`.
  - A longer test case is a `t.Run` with a helper.
  - A parameter list too long for one line takes one parameter per line with a trailing comma (`NewRouter`, `routes`).

## Frontend

- **Folders.**
  - `assets/app/<Domain>/` holds a domain's screens, with only the subfolders it needs (`Views`, `Components`, `Composables`, `types`, and a folder of plain functions like `Auth/Device`).
  - `assets/shared/<Module>/` holds code shared across the app.
  - `assets/standalone/` holds the extra Vite entries the Go templates load. Each only starts pieces of `shared/`.
  - `assets/router/` holds vue-router.
- **Components and composables.** A component's script calls its composable and binds what it returns. A screen's state and logic live in one composable per file in `app/<Domain>/Composables/` (`useLoginForm`). The composable:
  - creates its refs and returns state the component only reads through `readonly()`;
  - declares its functions and computeds as consts above the `return` and returns them by name: `return { form, errors, sending, submit }` (`codestyle/return-by-name`).

  A stateless operation stays a plain function (`completeLogin`, `authFetch`).
- **Code rules.**
  - A function is an arrow function bound to a `const`, never the `function` keyword.
  - No type assertions except `as const`.
  - Imports in `assets/` go through `@/`.
  - No inline `style`, which the CSP blocks.
  - No `v-html` and no `innerHTML` binding, which only ESLint stops ([Content Security Policy](docs/README.md#security)).
  - No `<svg>` in a Vue template outside `shared/Icons`.
  - Browser storage goes only through `appStorage()`.
  - A script that JS adds at runtime from another host loads only through `loadScript`, with a `ScriptURL` from Go. A script a Go template renders takes the nonce instead.
- **Texts.** A text follows a switch of the language only from the template or a `computed`. A text taken once in `setup` stays in the old language.
  - A key held in a variable or a prop is typed `PlainMessageKey` (a text without params; `isPlainMessageKey` checks one at runtime) or `MessageKeyWith<{ … }>` (exactly those params), so `t` accepts it without an assertion.
- **The router singleton.** `shared/Auth` (`authFetch`, `logout`, `completeLogin`) imports `router` from `@/router`, because it also runs outside `setup()`, where `useRouter()` doesn't exist. Nothing in `shared/Auth/Session` imports `@/router`, because the router imports it.
- **Redirects pass `lang`.** A redirect from a router guard and a target in another language set `params.lang` themselves (`sessionRedirect`, `redirectAfterLogin`): on the first load, the route a guard leaves has no params.
- **API calls.**
  - Call the API with `apiFetch` from `@/shared/Fetch`, and a route behind the session with `authFetch`, where a 401 leads to the login.
  - Every call passes a guard: `validate` for the success (`isNull` for an empty body) and `validateError` for the field errors of a form.
  - Render an `ApiMessage` with `tm`.
- **Types from Go.** The types the API sends and receives come from Go through tsgen. Never write them by hand, and never edit a file that starts with `// Code generated`. The exceptions are the generic `GridResult<T>`, since tsgen has no generics, and the envelope types in `shared/Fetch/Envelope`.
- **Layout of the code.** Lines stay within 120 columns. Lists follow `codestyle/list-layout`, which `yarn eslint --fix .` writes.

## Tests

- **One rule, one test.** Before you write a test, find the one that already guards the rule. A new case becomes a row of its table (`t.Run`, `it.each`), and when a test covers one half of a behaviour, the other half joins it, like the idle and the busy state of a button.
- **The lowest level that catches it.** A rule is tested in the domain package, or in a composable through its component. A route test checks only the wiring (binding, status, envelope, cookies) with one case that tells its failures apart: `en_GB` among `cs_CZ` and `en_US` becomes `en_US`, which neither the raw value nor the default gives.
- **Behaviour, not looks.** A test pins no order of elements, no copy and no Tailwind class, unless it passes them in itself or they are the behaviour (`sr-only`, `animate-spin`, a tooltip made only of CSS).
  - Texts come from the dictionaries: `t()` or `tm()` in Vitest, `All()` of `locale/` in Go. A test writes a text out only as its own input or as the output of a formatter it tests.
  - A test that switches the language compares a text with the dictionary of the new language, or takes it before the switch and checks that it changed. `t()` in the test switches too, so a failed switch would pass otherwise.
  - A test checks only what it is about: no whole struct that gains unrelated fields (the config), no restated constant, an error through `errors.Is` and one name it carries, not its whole list.
- **No list kept by hand.** A test derives what the code already lists: the files the templates name (`testkit.TemplateLiterals`), the variables of `.env.example`, a folder through `import.meta.glob`.
- **It fails when the rule breaks.** Before you commit a test, break the code it guards once and see it fail.
- **Deleting a test** needs a named test of the same property, or a rule that is gone. A check that passes by accident covers nothing, and docs that name the test change with it.
- **Go.**
  - Shared helpers sit behind the facade of `app/core/testkit`, and the router comes from `routertest`. Never copy them into a package.
  - A query is tested through the job that runs it, or in the `db` package when only its SQL holds the rule (`lower()`, the `WHERE` of a query meant for one row). The `db` package also tests the constraints of the schema. A query meant for one row runs beside a second row that must stay as it was.
  - A problem the request survives (`reporting.Warning`) is checked as the `WARN` line of its access log, through `routertest.WithLogs` or `testkit.CaptureLogs`.
- **Vitest.**
  - Tests of `assets/` live in `tests/assets/<area>/`.
  - A composable is tested through the component that uses it: what the user sees, not its refs.
  - Between tests, Vitest restores spies and stubbed globals, not module state or timers. A test resets those it touches in `afterEach` (`forgetSession`, `clearToasts`, `vi.useRealTimers()`) or takes fresh modules through `vi.resetModules()`.
  - A test that opens the app does it through `tests/assets/app/openApp.ts`, never with its own `mount(App, …)` or stub of the signed-in API, and finds the calls to a URL with `calls(path)`, never by their position.
  - A rule of this file that only the types hold is tested with `expectTypeOf`, which `yarn typecheck` checks.

## Recipes

### An API route
1. Write `app/internal/<domain>/handler/<route>.go`: a function of what the handler needs that returns `gin.HandlerFunc`.
   - Read the body with `api.Bind(c, &req)` into a DTO with one `json` tag and its `binding` rules per field (`dive` for a slice of structs). Read the query of a grid with `api.BindListing`.
   - Answer with `api.JSON(c, status, body)`. Answer errors with `api.Invalid` for fields (422) or with `api.Fail(c, status, key)` for a general error.
   - A JSON key or query parameter of several words is camelCase (`userAgent`, `perPage`).
2. A message key is a top-level `const KeyX api.Key = "<domain>.<what>"`. Give it a text in every `locale/*.go` and run `go run ./tools/i18n`.
3. Mark the request and answer types with `//tsgen:assets/app/<Domain>/types/<Name>.ts <Name>`, adding `request` for a body. Then run `go run ./tools/tsgen`.
4. Add the route to `routes.go` on its own line, in the group `apiRoutes`.
   - That group carries `api.InternalErrors()`, which turns `reporting.Error` into the 500 envelope. Outside it, `reporting.Error` only records the error, the response stays an empty 200, and Sentry never hears of it.
   - Put the route in the `signedIn` group, which sits inside `apiRoutes`, when it needs the session. The handler then reads the user with `authn.UserID(c)`.
   - To limit failed attempts, add `ratelimit.Middleware` with a store of its own scope (`failures.New(pool, "<scope>")`), its limit and window. The handler records each failure with `ratelimit.Failed(c)`. Without that call the limit never trips, and a shared scope would delete the other limit's attempts.
5. Test it in `<route>_test.go` (package `handler_test`) through its real path. Check the status, the envelope and the cookies, and set up state through the routes where they can make it.
   - `routertest.New(t)` gives the router on a migrated database. `WithLogs` adds the captured log, `WithLocales` more languages, and `WithMailbox` a mailbox for the mails.
   - A test without the database builds the router with `routertest.Build(t, &routertest.Parts{…})`. A zero field takes its default, and `routertest.Config()` gives the default config to change.
   - A new parameter of `router.NewRouter` becomes a field of `routertest.Parts` with its default in `Build`, so the tests stay as they are.

### A job of a domain
1. Create `app/internal/<domain>/<job>/` with a constructor (`account.New(pool, params)`) and sentinel errors. A test beside it comes when the job has logic beyond its query.
2. Test it directly on `testkit.Pool(t, migrations.FS)`. Pass cheap `password.Params` when it hashes.
3. Call it from the handler's closure.

### An environment variable
1. Add it to its section of `.env.example`, with allowed values, default and a short explanation, in the language of the file.
2. Read and check it in the package of its section under `app/internal/shared/config/`.
   - A new section gets a new package with `Read` and its tests, joined in `config.go`.
3. Add it to the configuration table in `docs/README.md`.

### A migration and a query
1. Write a new file `migrations/<yyyymmddhhmmss>_<name>.sql` with `-- +goose Up` and `-- +goose Down`. Once a migration has reached production, it never changes.
   - A migration must work with the rights of the app's role ([Database](docs/README.md#database)).
2. Put the queries in the domain's `db` package (`app/internal/user/db/users.sql`). Each starts with `-- name: <Name> :one`, `:many`, `:exec` or `:execrows`.
   - A new `db` package gets its entry in `sqlc.yaml`. Copy the `auth` entry, not the first one, which defines the anchor: it sets its own `queries`, and under `gen.go` merges `<<: *go` and sets its own `out`.
3. Run `go tool -modfile=tools/sqlc/go.mod sqlc generate`, and never edit the generated files. Upgrade sqlc with `go -C tools/sqlc get -tool github.com/sqlc-dev/sqlc/cmd/sqlc@<version>` and `go -C tools/sqlc mod tidy`, never with `go get` at the root, which would add it to the app's `go.mod`.
4. A domain writes only its own tables. It reads another domain's tables in a join, and changes them only through that domain's functions.

### A text or a language
- **A new key** needs its first use (`t('…')`, `{{t "…"}}` or an `api.Key` constant) before `go run ./tools/i18n` takes it. Then add its text to every dictionary and run the tool, which writes the TypeScript.
- **Key names** are two or more snake_case words joined by dots (`login.title`, `not_found.status`). A key starts with its screen or part (`form.email`), and an API key with its domain.
- **Messages** are ICU MessageFormat in the subset of [Languages and texts](docs/README.md#languages-and-texts). A plural needs every CLDR category of its language.
- **A highlighted word** inside a sentence gets its own key, and so do the parts around it. A fixed space is `\u00a0` in the dictionary.
- **A new language** is:
  - a new `locale/<xx_YY>.go` that registers itself in a one-line `init`;
  - its entry in `LOCALES`;
  - a flag `assets/img/flags/<country>.svg`;
  - a share image `assets/img/og/<xx_YY>.png` of 1200×630.

  Then run `go run ./tools/i18n`, which writes its dictionary module and its sample of numbers, and `yarn build`. `goNumbers.test.ts` then shows whether `Intl` formats its numbers like Go.

### A page rendered by Go
1. Write `views/<folder>/<page>.html`. It starts with `{{template "shared/layout.html" .}}` and defines `title`, `description`, `assets` (`{{vite .Nonce "assets/app.css" …}}`) and `content`.
   - On a subpage, `title` is `{{t "<page>.title"}} | {{t "brand.name"}}`.
   - A bigger page splits into `views/<folder>/components/`, which are called by their path and define no template.
2. The handler calls `renderer.Page(c, status, "<folder>/<page>.html", params)`. Its route needs the locale middleware: `locales.Home()` for `/`, `locales.Prefix(notFound)` for a path with a language.
3. Every file `{{asset "…"}}` names must be in the Vite manifest, or the start refuses it. A file only a template uses gets there through an import in `assets/app.ts`.
4. Add the texts, then run `go run ./tools/i18n`, which refuses an unknown key in a template.
5. Run `go run ./tools/codestyle`, which lays the template out.
6. Add the page to the table of `TestEveryPageRendersInEveryLanguage`, which checks every Go page in every language, from the status to the cookie settings. A test of the page itself checks only what the page adds, through `testkit.Markup(res)`.

Three behaviours of the templates catch people out:
- **Empty blocks.** An empty `{{define}}` doesn't drop a block of the layout, because `text/template` keeps the old one. Replace the block with content instead, as the 404 replaces the share card with `noindex`.
- **Whitespace.** The layout tool assumes the default `white-space`. Text whose whitespace matters goes into `pre` or `textarea`. Inline items without spaces between them stay glued, so a flex row needs spaces in the template.
- **Dropdowns.** A page with a `<details data-dropdown>` loads `assets/standalone/dropdowns.ts`, or the dropdown never closes.

### A screen in the SPA
1. Add an `AppRoute` to `assets/router/routes.ts`:
   - the full path `/:lang/app/<path>`;
   - a `name`;
   - a `meta` with `access` (`Guest`, `User` or `Public`), `title` and `description` keys.
2. Put the view in `app/<Domain>/Views/` and its state in a composable in `app/<Domain>/Composables/`. Build it from the pieces in `shared/`: fields, buttons, the modal, toasts and the grid.
3. Add the texts to the dictionaries and run `go run ./tools/i18n`.
4. Test what the screen does that the ground rules ask a test for. The test lives in `tests/assets/<area>/`, opens the screen with `openApp` or `openSignedIn` from `tests/assets/app/openApp.ts` and checks what the user sees. A new API call of a signed-in screen gets its default answer in `answerApi` there.

### A grid
1. **Go, domain package.** Declare the sort columns as a string type with `//tsgen:… <Name> union` (`session.Sort`), and the grid as `listing.Grid[Sort]` beside the list function (`session.Grid`).
2. **Go, query.** The list takes `listing.Query[Sort]`, whose `Offset()` and `Limit()` go into the SQL. The SQL sorts through a `CASE` per column and direction, with the ID last.
3. **Go, handler.** It reads the query with `api.BindListing(c, grid, &filters)`, where the filters are a struct with `form` and `binding` tags. It answers `listing.Result[Item]`, or `listing.SelectableResult` when some rows can't be selected.
4. **Frontend.** Call `useGrid({ name, url, item, isSort, sort, filters, isFilter, perPage, selectable })` and render it with `DataGrid`, `GridPagination` and `GridFrame`, plus `FilterPanel` and `BulkActionBar` when needed.
   - `isFilter` keeps the limits of the Go `binding` tags.
   - `name` prefixes the grid's keys in the address.

   The model is `useSessionsGrid` with `SessionsSection`.
5. **Bulk actions.** A bulk action sends either the chosen IDs or, for "all", `grid.appliedFilters`, never what the filter fields hold.
   - Its Go route applies the same filter with the same limit and query as the grid, so "all" means exactly what the grid shows.
   - The route never acts on rows the user can't select.
   - The screen asks through `ConfirmModal` first, and reports the count the server answers.

### A mail
1. Write `views/mail/<name>.html`. It calls no layout and defines the block `subject`.
   - It follows the rules of mail clients: tables with `role="presentation"`, inline `style`, PNG images with `width`, `height` and `alt`, and `align` only on a `td`.
   - Every link to the app starts with `{{.Origin}}`.
2. In the domain's package `mail`, write a function that calls `mailer.Send(ctx, &coremail.Template{Name, To, Locale, Params})`, as `usermail.Welcome` does. The handler gets the `mailer` from `routes.go`, and a failed send goes to `reporting.Warning`.
3. Test the mail through `routertest.Renderer` and `testkit.MarkupOf` (`welcome_test.go`). Test the handler that sends it through `routertest.WithMailbox`.

### A tracked event
1. Add a constant to `app/internal/tracking/events`: a `TrackedEvent`, and an `AdsConversion` listed in `AdsConversions` when Google Ads counts it. Run `go run ./tools/tsgen`.
2. Map the event in `trackedEvents` of `shared/Tracking/trackedEvents.ts`; typecheck refuses an event without a row. Call `trackEvent(TrackedEvent.X)`.
   - The label of an Ads conversion goes into `GOOGLE_ADS_CONVERSIONS`, never into the code.

### A script or service from another host
1. Add a `csp.Service` with its script and frame hosts and its Trusted Types policy names.
   - A service any project can use goes into `app/core/csp/services.go`.
   - A host of this project goes into `Policy` in `app/internal/shared/csp`, which switches the services on.
2. A script JS adds at runtime gets a `ScriptURL` constant in `app/internal/shared/csp`, which tsgen writes, and loads through `loadScript`. A Go test wants its origin in `script-src`.
3. A tool that sets cookies waits for consent through `followChoice`.

### An icon
1. Add `assets/img/icons/<name>.svg` with a 16×16 `viewBox`, the stroke attributes on the root only and the drawing in one element `id="icon"`.
2. Add a component `shared/Icons/Icon<Name>.vue` that draws `<use href="@/img/icons/<name>.svg#icon" />` inside `BaseIcon`.
3. A Go template draws the icon as `<svg viewBox="0 0 16 16" …><use href="{{asset "assets/img/icons/<name>.svg"}}#icon"/></svg>` with the attributes of `BaseIcon`. Then `assets/app.ts` imports the file.

## Settled decisions

These were chosen with their costs in mind. Don't propose to undo them.

- **Sign-in.**
  - Sessions live in PostgreSQL behind an HttpOnly cookie, not in a JWT. A session can be revoked at once, and a JWT would save only one indexed query. A mobile app can later send the same token as `Authorization: Bearer`.
  - Passwords are hashed with argon2id.
  - The limit of failed attempts lives in the table `auth_failures`. Accepted cost: a parallel burst can pass the limit by a few attempts, because the middleware counts before the hash and the failure is recorded after it.
  - Roles and permissions don't exist yet, because they depend on the project and on multitenancy.
- **Database.** PostgreSQL through pgx, goose migrations at start under a lock, and queries in SQL through sqlc. No ORM and no schema tool; migrations are written by hand.
- **Undo.** A failed later step of a handler is undone by a domain delete reported to Sentry, not by a transaction across a hash.
- **Operations.**
  - `/metrics` runs on the app's port behind basic auth, not on its own port.
  - The access log keeps `/ping`, `/readyz` and `/metrics`.
  - Logs are slog on stdout, JSON by default. There is no OpenTelemetry and no tracing backend: the trace ID comes from Sentry's `sentry-trace`.
- **CSP and measurement.**
  - The CSP comes from our own middleware, with nonces, `strict-dynamic` and Trusted Types.
  - `img-src` and `connect-src` allow `https:`.
  - Measurement tools are added one by one, without Google Tag Manager, and Hotjar is out.
- **Code shape.**
  - gofmt stays. Blank lines between split fields replace its columns.
  - A handler is a function of its dependencies, not a method of a struct.
  - One tool covers one concept. A new kind of check gets its own tool, not a mode of `tools/i18n`.

## Traps

- **The manifest checks.** `yarn dev` writes `public/.hot`, and while it exists the templates may name any file. The checks against the manifest run only against a build.
- **Streaming.** The server cuts a write after 10 s. A handler that streams or sends a long download extends its own deadline through `http.ResponseController`.
- **`public/` is embedded.** A file added there shows only after a new build of the binary. Its name must be one `go:embed` accepts.
- **The Dockerfile.** A new build input or Tailwind `@source` outside `assets/` and `views/` needs a `COPY` in its `assets` stage.
- **Pages drawn twice** in Go and in Vue (see the comment exception) change together.
- **The failure floor.** A slower password hash needs a higher `account.FailureFloor` ([Passwords](docs/README.md#sign-in-and-sessions)).
- **Frontend tests.** The router keeps the route of the previous test, so a test that mounts a routed component navigates first.
