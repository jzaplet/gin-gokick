package locale

var enUS = Dictionary{
	"brand.name": "Gin GoKick",

	"home.title": "Lightweight full-stack Golang boilerplate",

	"home.description": "Build apps in Go the light, clean and fast way. Gin, Vue\u00a03 and strict rules under which people and AI write equally good code.",

	"home.heading_before": "Lightweight",

	"home.heading_full_stack": "full-stack",

	"home.heading_middle": "",

	"home.heading_highlight": "Golang",

	"home.heading_after": "boilerplate",

	"home.clone_highlight": "Clone the repo",

	"home.clone_rest": "and let AI do the work",

	"home.repositories": "GitHub repositories",

	"home.repo_gokick": "Go to the Gin GoKick repo",

	"home.repo_gin": "Go to Gin",

	"home.features": "What Gin GoKick does",

	"home.group.start": "Getting started",

	"home.group.database": "Database",

	"home.group.deploy": "Deployment",

	"home.group.frontend": "Frontend",

	"home.group.templates": "Go templates",

	"home.group.components": "Components",

	"home.group.languages": "Languages and SEO",

	"home.group.security": "Security",

	"home.group.telemetry": "Sentry and telemetry",

	"home.group.measurement": "Analytics and consent",

	"home.feature.go_gin.title": "Go and Gin",

	"home.feature.go_gin.text": "One binary serves the templates and the frontend build.",

	"home.feature.env.title": "Config from .env",

	"home.feature.env.text": "Variables are checked. A typo stops the start, not production.",

	"home.feature.agents.title": "Rules in AGENTS.md",

	"home.feature.agents.text": "People and AI write by the same rules, and tools enforce them.",

	"home.feature.gate.title": "Pre-commit gate",

	"home.feature.gate.text": "One block runs the lint, types and tests of CI before a commit.",

	"home.feature.push_hook.title": "Pre-push hook",

	"home.feature.push_hook.text": "Lefthook checks the branch name and commit messages before a push.",

	"home.feature.postgres.title": "PostgreSQL 18",

	"home.feature.postgres.text": "The app connects through pgx, one command starts the database.",

	"home.feature.goose.title": "goose migrations",

	"home.feature.goose.text": "They run at start under a lock, so instances never collide.",

	"home.feature.sqlc.title": "sqlc instead of an ORM",

	"home.feature.sqlc.text": "You write SQL, and typed Go functions are generated from it.",

	"home.feature.db_tests.title": "Tests with a database",

	"home.feature.db_tests.text": "Each test gets an empty migrated schema, in CI and locally.",

	"home.feature.docker.title": "Docker image",

	"home.feature.docker.text": "The image runs on Alpine without root and checks the database.",

	"home.feature.dokploy.title": "Dokploy guide",

	"home.feature.dokploy.text": "Guides the first deploy with Postgres, Cloudflare and Grafana.",

	"home.feature.shutdown.title": "Graceful restart",

	"home.feature.shutdown.text": "On shutdown, requests in progress finish within five seconds.",

	"home.feature.mail.title": "Emails over SMTP",

	"home.feature.mail.text": "Mailpit catches them locally, Amazon SES sends them live.",

	"home.feature.vite.title": "Vite the Laravel way",

	"home.feature.vite.text": "HMR runs through Go, so development gets the production CSP.",

	"home.feature.vue_tailwind.title": "Vue\u00a03 and Tailwind\u00a04",

	"home.feature.vue_tailwind.text": "The SPA with vue-router and the Go pages share one style.",

	"home.feature.go_types.title": "Go types and guards",

	"home.feature.go_types.text": "tsgen writes types and guards, so API changes fail to compile.",

	"home.feature.routes.title": "Routes by sign-in",

	"home.feature.routes.text": "Every route says who may open it, and tabs share the sign-in.",

	"home.feature.ssr.title": "SSR with html/template",

	"home.feature.ssr.text": "The server renders the home page, the 404 and the app entry in each language.",

	"home.feature.template_checks.title": "Guarded templates",

	"home.feature.template_checks.text": "Checks catch missing files, untranslated text and unknown keys in templates.",

	"home.feature.mail_templates.title": "Mail templates",

	"home.feature.mail_templates.text": "Mails come from Go templates and reach users in their own language.",

	"home.feature.components.title": "Shared elements",

	"home.feature.components.text": "Icons, dropdowns and tooltips look the same in Vue and Go.",

	"home.feature.grid.title": "Data grid",

	"home.feature.grid.text": "The URL keeps the filters, and a selection runs bulk actions.",

	"home.feature.forms.title": "Forms",

	"home.feature.forms.text": "Text and number fields show the API error right below them.",

	"home.feature.toasts.title": "Toasts and dialogs",

	"home.feature.toasts.text": "Toasts survive navigation, and a modal confirms risky actions.",

	"home.feature.languages.title": "Languages in the URL",

	"home.feature.languages.text": "Languages live in the URL with hreflang, accounts keep theirs.",

	"home.feature.translations.title": "Translation checks",

	"home.feature.translations.text": "tools/i18n catches missing keys, unused texts and bad plurals.",

	"home.feature.icu.title": "ICU in Go and Vue",

	"home.feature.icu.text": "Plurals and numbers look the same in Go and in the browser.",

	"home.feature.titles.title": "Titles and sharing",

	"home.feature.titles.text": "A missing title, description or share image fails the checks.",

	"home.feature.sign_in.title": "Sign-in",

	"home.feature.sign_in.text": "Passwords use argon2id, the attempt limit spans all instances.",

	"home.feature.sessions.title": "Signed-in devices",

	"home.feature.sessions.text": "Users see every device with its last IP and can sign it out.",

	"home.feature.csp.title": "CSP with a nonce",

	"home.feature.csp.text": "A nonce and Trusted Types keep injected scripts from running.",

	"home.feature.csrf.title": "CSRF without tokens",

	"home.feature.csrf.text": "http.CrossOriginProtection refuses requests from other sites.",

	"home.feature.headers.title": "A+ headers",

	"home.feature.headers.text": "HSTS, COOP, CORP and Permissions-Policy come from the server.",

	"home.feature.sentry.title": "Sentry for Go and Vue",

	"home.feature.sentry.text": "Links browser errors to the backend, releases are commit SHAs.",

	"home.feature.metrics.title": "Prometheus metrics",

	"home.feature.metrics.text": "Prometheus gets the metrics and Grafana a ready dashboard.",

	"home.feature.logs.title": "JSON logs",

	"home.feature.logs.text": "slog writes trace_id and user_id, never an email or a name.",

	"home.feature.cookie_banner.title": "Cookie banner",

	"home.feature.cookie_banner.text": "Refusing is as easy as accepting, and the choice can change any time.",

	"home.feature.consent.title": "Analytics and conversions",

	"home.feature.consent.text": "GA4, Ads and Meta measure and count conversions only after consent.",

	"home.feature.hashed_email.title": "Hashed email",

	"home.feature.hashed_email.text": "Only Ads and Meta get the email, and only as SHA-256.",

	"home.feature.tracked_events.title": "Events from Go",

	"home.feature.tracked_events.text": "Go defines the events, and typecheck lets none go unmapped.",

	"home.try_live": "Try it live:",

	"home.try_register": "Sign up",

	"home.try_login": "Sign in",

	"app.title": "App",

	"app.description": "Sign in, sign up and see an overview of your Gin GoKick account.",

	"app.loading": "Loading…",

	"app.loading_slow": "Loading is taking longer than usual.",

	"app.reload": "Reload the page",

	"form.email": "Email",

	"form.password": "Password",

	"login.title": "Sign in",

	"login.description": "Sign in to Gin GoKick with your email and password.",

	"login.submit": "Sign in",

	"login.no_account": "Don’t have an account?",

	"login.to_register": "Sign up",

	"login.signed_in": "You have signed in successfully.",

	"login.signed_in_title": "Signed in",

	"register.title": "Sign up",

	"register.description": "Create a Gin GoKick account with your email and password.",

	"register.submit": "Create account",

	"register.has_account": "Already have an account?",

	"register.to_login": "Sign in",

	"register.signed_up": "You have signed up successfully.",

	"register.signed_up_title": "Signed up",

	"dashboard.title": "Dashboard",

	"dashboard.description": "See an overview of your Gin GoKick account.",

	"dashboard.lead": "Here you’ll find the most important information about your app.",

	"dashboard.empty": "The dashboard has no data yet. It will fill in as the app gains more features.",

	"account.logout": "Sign out",

	"sessions.title": "Sign-in overview",

	"sessions.lead": "The devices and browsers you are signed in on.",

	"sessions.device": "Device",

	"sessions.ip": "IP address",

	"sessions.created_at": "Signed in",

	"sessions.last_seen_at": "Last active",

	"sessions.current": "This device",

	"sessions.unknown_device": "Unknown device",

	"sessions.unknown_ip": "Unknown",

	"sessions.empty": "You are not signed in on any device.",

	"grid.loading": "Loading…",

	"grid.failed": "The list could not be loaded.",

	"grid.retry": "Try again",

	"grid.range": "{from}–{to} of {total}",

	"grid.pages": "Pagination",

	"grid.page": "Page {page}",

	"grid.previous": "Previous page",

	"grid.next": "Next page",

	"grid.no_match": "Nothing matches the filters.",

	"filters.title": "Filters",

	"filters.active": "Filters (active)",

	"filters.clear": "Clear filters",

	"sessions.ip_filter": "The whole address or a part of it",

	"grid.select_page": "Select the whole page",

	"grid.actions": "Actions",

	"bulk.actions": "Bulk actions",

	"bulk.selected": "Selected: {count} / {total}",

	"bulk.selected_all": "Selected: all ({total})",

	"bulk.select_all": "Select all ({total})",

	"bulk.clear": "Clear selection",

	"bulk.action": "{action} ({count}×)",

	"sessions.select": "Select the device {device}",

	"sessions.sign_out": "Sign out",

	"sessions.sign_out_device": "Sign out the device",

	"sessions.sign_out_current": "Sign out this device with the “Sign out” button.",

	"sessions.sign_out_one": "We will sign out {device}, and working on it will need a new sign-in.",

	"sessions.sign_out_selected_title": "Sign out the selected devices",

	"sessions.sign_out_selected": "{count, plural, one {We will sign out # selected device.} other {We will sign out # selected devices.}} This device stays signed in.",

	"sessions.signed_out_title": "Devices signed out",

	"sessions.signed_out": "{count, plural, one {We signed out # device.} other {We signed out # devices.}}",

	"sessions.signed_out_none": "No device was signed out.",

	"sessions.sign_out_failed": "The devices could not be signed out.",

	"account.retry": "Try again",

	"account.signed_out": "You have signed out successfully.",

	"account.signed_out_title": "Signed out",

	"toast.close": "Close the notification",

	"toast.error_title": "Error",

	"modal.cancel": "Cancel",

	"modal.confirm": "Confirm",

	"not_found.status": "404",

	"not_found.title": "Page not found",

	"not_found.description": "There is no page at this address.",

	"navigation.back_home": "Back to home",

	"language.change": "Change language",

	"language.name": "English",

	"footer.gin": "Gin Framework",

	"footer.copyright": "©\u00a0{year} Gin GoKick",

	"mail.home": "Home",

	"mail.welcome.subject": "Welcome to Gin GoKick",

	"mail.welcome.preheader": "Your account is set up and the app is waiting for you.",

	"mail.welcome.heading": "Your account is ready",

	"mail.welcome.account": "This is the address you sign in with.",

	"mail.welcome.intro": "Thanks for signing up. Sign in and the dashboard of the app is waiting for you.",

	"mail.welcome.open_app": "Open the app",

	"mail.welcome.rules": "The rules for people and AI agents and every code check are in the repository.",

	"mail.welcome.rules_link": "Gin GoKick on GitHub",

	"mail.welcome.gin": "The backend runs on the Gin web framework.",

	"mail.welcome.gin_link": "Gin on GitHub",

	"mail.welcome.reason": "You're getting this email because someone signed up for Gin GoKick with this address. If it wasn't you, ignore it.",

	"consent.title": "Cookie consent",

	"consent.lead": "Necessary cookies keep the site running. Other tools set cookies only with your consent. Clicking “{button}” accepts all of them, or you can choose in the detailed settings. You can change your choice at any time.",

	"consent.details": "Detailed settings",

	"consent.necessary_only": "Accept necessary",

	"consent.understood": "Got it",

	"consent.settings": "Cookie settings",

	"consent.settings_lead": "Necessary cookies are always on, the others only with your consent. The arrow next to each group shows what it is for and which tools we use.",

	"consent.close": "Close",

	"consent.necessary": "Necessary cookies",

	"consent.necessary_text": "They keep the site running, for example signing in. The site doesn’t work without them, so they are always on.",

	"consent.always_on": "Always on",

	"consent.analytics": "Analytics cookies",

	"consent.analytics_text": "They measure how visitors use the site. We use {tools}.",

	"consent.marketing": "Advertising and marketing cookies",

	"consent.marketing_text": "They measure the results of ads and help target them. We use {tools}.",

	"consent.save": "Save settings",

	"consent.accept_all": "Accept all",

	"tracking.ga4": "Google Analytics",

	"tracking.google_ads": "Google Ads",

	"tracking.meta_pixel": "Meta Pixel",

	"fetch.network_error": "The server could not be reached. Check your connection and try again.",

	"fetch.malformed_body": "The server sent a response we don’t understand.",

	"fetch.invalid_shape": "The server sent a response in an unexpected shape.",

	"fetch.error_status": "The request failed, the server responded with status {status}.",

	"request.invalid_body": "The request could not be read.",

	"request.invalid_query": "The parameters of the request are not valid.",

	"request.body_too_large": "The request is too large.",

	"request.internal": "Something went wrong on the server. Try again.",

	"request.not_found": "The requested address does not exist.",

	"request.too_many_attempts": "Too many failed attempts. Try again later.",

	"validation.required": "Fill in this field.",

	"validation.email": "Enter a valid email address.",

	"validation.uuid": "The identifier is not in a valid format.",

	"validation.min_length": "Enter at least {min, plural, one {# character} other {# characters}}.",

	"validation.max_length": "Enter at most {max, plural, one {# character} other {# characters}}.",

	"validation.min_items": "Enter at least {min, plural, one {# item} other {# items}}.",

	"validation.max_items": "Enter at most {max, plural, one {# item} other {# items}}.",

	"validation.min": "The smallest allowed value is {min}.",

	"validation.max": "The largest allowed value is {max}.",

	"validation.one_of": "Choose one of the values: {values}.",

	"validation.invalid": "The value is not valid.",

	"auth.login_failed": "The email or password is incorrect.",

	"auth.sign_in_required": "Sign in to continue.",

	"user.email_taken": "An account with this email already exists.",
}

func init() { register("en_US", enUS) }
