package locale

var csCZ = Dictionary{
	"brand.name": "Gin GoKick",

	"home.title": "Lehká full-stack kostra pro Golang",

	"home.description": "Stavte aplikace v\u00a0Go lehce, čistě a\u00a0rychle. Gin, Vue\u00a03 a\u00a0přísná pravidla, podle kterých píše stejně dobrý kód člověk i\u00a0AI.",

	"home.heading_before": "Lehká",

	"home.heading_full_stack": "full-stack",

	"home.heading_middle": "kostra pro",

	"home.heading_highlight": "Golang",

	"home.heading_after": "",

	"home.clone_highlight": "Naklonujte repo",

	"home.clone_rest": "a\u00a0nechte AI pracovat",

	"home.repositories": "Repozitáře na GitHubu",

	"home.repo_gokick": "Přejít na Gin GoKick repo",

	"home.repo_gin": "Přejít na Gin",

	"home.features": "Co Gin GoKick umí",

	"home.group.start": "Rozjetí",

	"home.group.database": "Databáze",

	"home.group.deploy": "Nasazení",

	"home.group.frontend": "Frontend",

	"home.group.templates": "Šablony v\u00a0Go",

	"home.group.components": "Komponenty",

	"home.group.languages": "Jazyky a\u00a0SEO",

	"home.group.security": "Bezpečnost",

	"home.group.telemetry": "Sentry a\u00a0telemetrie",

	"home.group.measurement": "Měření a\u00a0souhlas",

	"home.feature.go_gin.title": "Go a\u00a0Gin",

	"home.feature.go_gin.text": "Gin obslouží web z\u00a0jedné binárky i\u00a0se šablonami a\u00a0frontendem.",

	"home.feature.env.title": "Konfigurace z\u00a0.env",

	"home.feature.env.text": "Proměnné mají kontrolu. Překlep zastaví start, ne produkci.",

	"home.feature.agents.title": "Pravidla v\u00a0AGENTS.md",

	"home.feature.agents.text": "Člověk i\u00a0AI píšou podle stejných pravidel, nástroje je vynutí.",

	"home.feature.gate.title": "Brána před commitem",

	"home.feature.gate.text": "Jeden blok spustí před commitem lint, typy i\u00a0testy jako CI.",

	"home.feature.push_hook.title": "Kontrola před pushem",

	"home.feature.push_hook.text": "Lefthook před pushem ověří název větve i\u00a0zprávy commitů.",

	"home.feature.postgres.title": "PostgreSQL 18",

	"home.feature.postgres.text": "Aplikace se připojí přes pgx, databázi spustí jeden příkaz.",

	"home.feature.goose.title": "Migrace goose",

	"home.feature.goose.text": "Spustí se při startu se zámkem, takže se instance nepoperou.",

	"home.feature.sqlc.title": "sqlc místo ORM",

	"home.feature.sqlc.text": "Dotazy píšete v\u00a0SQL, typované funkce v\u00a0Go se vygenerují.",

	"home.feature.db_tests.title": "Testy s\u00a0databází",

	"home.feature.db_tests.text": "Každý test dostane prázdné schéma s\u00a0migracemi, v\u00a0CI i\u00a0lokálně.",

	"home.feature.docker.title": "Docker image",

	"home.feature.docker.text": "Image běží na Alpine bez roota a\u00a0healthcheck ověří i\u00a0databázi.",

	"home.feature.dokploy.title": "Návod na Dokploy",

	"home.feature.dokploy.text": "Provede prvním nasazením s\u00a0Postgresem, Cloudflarem i\u00a0Grafanou.",

	"home.feature.shutdown.title": "Klidný restart",

	"home.feature.shutdown.text": "Při vypnutí požadavky doběhnou, nejdéle do pěti sekund.",

	"home.feature.mail.title": "E-maily přes SMTP",

	"home.feature.mail.text": "Lokálně je chytí Mailpit, v\u00a0produkci je odešle Amazon SES.",

	"home.feature.vite.title": "Vite po vzoru Laravelu",

	"home.feature.vite.text": "HMR jde přes Go, takže i\u00a0vývoj běží s\u00a0produkční CSP.",

	"home.feature.vue_tailwind.title": "Vue\u00a03 a\u00a0Tailwind\u00a04",

	"home.feature.vue_tailwind.text": "SPA s\u00a0vue-routerem i\u00a0stránky z\u00a0Go sdílejí jeden styl.",

	"home.feature.go_types.title": "Typy a\u00a0guardy z\u00a0Go",

	"home.feature.go_types.text": "tsgen píše typy i\u00a0guardy, změnu API odhalí kompilace.",

	"home.feature.routes.title": "Routy podle přihlášení",

	"home.feature.routes.text": "Každá routa říká, kdo na ni smí, a\u00a0záložky sdílejí přihlášení.",

	"home.feature.ssr.title": "SSR přes html/template",

	"home.feature.ssr.text": "Homepage, stránku 404 i\u00a0vstup do aplikace vykreslí server v\u00a0každém jazyce.",

	"home.feature.template_checks.title": "Hlídané šablony",

	"home.feature.template_checks.text": "Kontroly odhalí v\u00a0šabloně chybějící soubor, nepřeložený text i\u00a0neznámý klíč.",

	"home.feature.mail_templates.title": "Mailové šablony",

	"home.feature.mail_templates.text": "Maily vzniknou z\u00a0Go šablon a\u00a0uživateli přijdou v\u00a0jeho jazyce.",

	"home.feature.components.title": "Sdílené prvky",

	"home.feature.components.text": "Ikony, dropdowny a\u00a0tooltipy vypadají ve Vue i\u00a0v\u00a0Go stejně.",

	"home.feature.grid.title": "Datagrid",

	"home.feature.grid.text": "Filtry i\u00a0stránka zůstanou v\u00a0adrese a\u00a0výběr umí hromadné akce.",

	"home.feature.forms.title": "Formuláře",

	"home.feature.forms.text": "Pole pro text i\u00a0čísla ukážou chybu z\u00a0API přímo pod sebou.",

	"home.feature.toasts.title": "Toasty a\u00a0potvrzení",

	"home.feature.toasts.text": "Toast přežije navigaci a\u00a0před nevratnou akcí se zeptá modal.",

	"home.feature.languages.title": "Jazyky v\u00a0adrese",

	"home.feature.languages.text": "Jazyk má vlastní adresu s\u00a0hreflang a\u00a0účet si ho pamatuje.",

	"home.feature.translations.title": "Kontrola překladů",

	"home.feature.translations.text": "tools/i18n hlídá chybějící klíče, nepoužité texty i\u00a0plurály.",

	"home.feature.icu.title": "ICU v\u00a0Go i\u00a0Vue",

	"home.feature.icu.text": "Plurály a\u00a0čísla vypadají v\u00a0Go i\u00a0v\u00a0prohlížeči stejně.",

	"home.feature.titles.title": "Titulky a\u00a0sdílení",

	"home.feature.titles.text": "Bez titulku, popisu nebo obrázku ke sdílení neprojdou kontroly.",

	"home.feature.sign_in.title": "Přihlášení",

	"home.feature.sign_in.text": "Hesla hashuje argon2id, limit pokusů platí napříč instancemi.",

	"home.feature.sessions.title": "Přehled zařízení",

	"home.feature.sessions.text": "Uživatel vidí svá zařízení s\u00a0poslední IP a\u00a0odhlásí, co nezná.",

	"home.feature.csp.title": "CSP s\u00a0nonce",

	"home.feature.csp.text": "Nonce a\u00a0Trusted Types zastaví cizí skript bez unsafe-inline.",

	"home.feature.csrf.title": "CSRF bez tokenů",

	"home.feature.csrf.text": "Požadavek z\u00a0cizí stránky odmítne http.CrossOriginProtection.",

	"home.feature.headers.title": "Hlavičky na A+",

	"home.feature.headers.text": "HSTS, COOP, CORP i\u00a0Permissions-Policy pošle server sám.",

	"home.feature.sentry.title": "Sentry pro Go i\u00a0Vue",

	"home.feature.sentry.text": "Spojí chybu z\u00a0prohlížeče s\u00a0backendem, release je SHA commitu.",

	"home.feature.metrics.title": "Prometheus a\u00a0Grafana",

	"home.feature.metrics.text": "Hotový dashboard ukáže požadavky, doby odpovědí i\u00a0runtime Go.",

	"home.feature.logs.title": "Logy v\u00a0JSON",

	"home.feature.logs.text": "slog zapíše trace_id i\u00a0user_id, nikdy e-mail ani jméno.",

	"home.feature.cookie_banner.title": "Cookie lišta",

	"home.feature.cookie_banner.text": "Odmítnutí je stejně snadné jako souhlas a\u00a0volbu lze kdykoli změnit.",

	"home.feature.consent.title": "Měření a\u00a0konverze",

	"home.feature.consent.text": "GA4, Ads i\u00a0Meta měří a\u00a0počítají konverze až po souhlasu.",

	"home.feature.hashed_email.title": "E-mail jen jako hash",

	"home.feature.hashed_email.text": "E-mail dostanou jen Ads a\u00a0Meta, a\u00a0to jako hash SHA-256.",

	"home.feature.tracked_events.title": "Události z\u00a0Go",

	"home.feature.tracked_events.text": "Události definuje Go a\u00a0typecheck nepustí žádnou bez mapování.",

	"home.try_live": "Vyzkoušejte živě:",

	"home.try_register": "Registrace",

	"home.try_login": "Přihlásit se",

	"app.title": "Aplikace",

	"app.description": "V aplikaci se přihlásíte, zaregistrujete a uvidíte přehled svého účtu v Gin GoKick.",

	"app.loading": "Načítání…",

	"app.loading_slow": "Načítání trvá déle než obvykle.",

	"app.reload": "Obnovit stránku",

	"form.email": "E-mail",

	"form.password": "Heslo",

	"login.title": "Přihlášení",

	"login.description": "Přihlaste se do Gin GoKick e-mailem a heslem.",

	"login.submit": "Přihlásit se",

	"login.no_account": "Nemáte účet?",

	"login.to_register": "Zaregistrujte se",

	"login.signed_in": "Přihlášení proběhlo úspěšně.",

	"login.signed_in_title": "Přihlášení",

	"register.title": "Registrace",

	"register.description": "Založte si účet v Gin GoKick e-mailem a heslem.",

	"register.submit": "Vytvořit účet",

	"register.has_account": "Už máte účet?",

	"register.to_login": "Přihlaste se",

	"register.signed_up": "Registrace proběhla úspěšně.",

	"register.signed_up_title": "Registrace",

	"dashboard.title": "Přehled",

	"dashboard.description": "Zde vidíte přehled svého účtu v Gin GoKick.",

	"dashboard.lead": "Zde najdete nejdůležitější informace o své aplikaci.",

	"dashboard.empty": "Přehled zatím neobsahuje žádné údaje. Přibudou postupně s dalšími funkcemi aplikace.",

	"account.logout": "Odhlásit se",

	"sessions.title": "Přehled přihlášení",

	"sessions.lead": "Zařízení a prohlížeče, ve kterých jste přihlášeni.",

	"sessions.device": "Zařízení",

	"sessions.ip": "IP adresa",

	"sessions.created_at": "Přihlášeno",

	"sessions.last_seen_at": "Naposledy aktivní",

	"sessions.current": "Toto zařízení",

	"sessions.unknown_device": "Neznámé zařízení",

	"sessions.unknown_ip": "Neznámá",

	"sessions.empty": "Nejste přihlášeni na žádném zařízení.",

	"grid.loading": "Načítání…",

	"grid.failed": "Seznam se nepodařilo načíst.",

	"grid.retry": "Zkusit znovu",

	"grid.range": "{from}–{to} z {total}",

	"grid.pages": "Stránkování",

	"grid.page": "Stránka {page}",

	"grid.previous": "Předchozí stránka",

	"grid.next": "Další stránka",

	"grid.no_match": "Filtrům neodpovídá žádná položka.",

	"filters.title": "Filtry",

	"filters.active": "Filtry (aktivní)",

	"filters.clear": "Zrušit filtry",

	"sessions.ip_filter": "Celá adresa nebo její část",

	"grid.select_page": "Vybrat celou stránku",

	"grid.actions": "Akce",

	"bulk.actions": "Hromadné akce",

	"bulk.selected": "Vybráno: {count} / {total}",

	"bulk.selected_all": "Vybráno: vše ({total})",

	"bulk.select_all": "Vybrat vše ({total})",

	"bulk.clear": "Zrušit výběr",

	"bulk.action": "{action} ({count}×)",

	"sessions.select": "Vybrat zařízení {device}",

	"sessions.sign_out": "Odhlásit",

	"sessions.sign_out_device": "Odhlásit zařízení",

	"sessions.sign_out_current": "Toto zařízení odhlásíte tlačítkem „Odhlásit se“.",

	"sessions.sign_out_one": "Zařízení {device} odhlásíme a pro další práci se na něm bude nutné znovu přihlásit.",

	"sessions.sign_out_selected_title": "Odhlásit vybraná zařízení",

	"sessions.sign_out_selected": "{count, plural, one {Odhlásíme # vybrané zařízení.} few {Odhlásíme # vybraná zařízení.} many {Odhlásíme # vybraného zařízení.} other {Odhlásíme # vybraných zařízení.}} Toto zařízení zůstane přihlášené.",

	"sessions.signed_out_title": "Odhlášení zařízení",

	"sessions.signed_out": "{count, plural, one {Odhlásili jsme # zařízení.} few {Odhlásili jsme # zařízení.} many {Odhlásili jsme # zařízení.} other {Odhlásili jsme # zařízení.}}",

	"sessions.signed_out_none": "Žádné zařízení jsme neodhlásili.",

	"sessions.sign_out_failed": "Zařízení se nepodařilo odhlásit.",

	"account.retry": "Zkusit znovu",

	"account.signed_out": "Odhlášení proběhlo úspěšně.",

	"account.signed_out_title": "Odhlášení",

	"toast.close": "Zavřít oznámení",

	"toast.error_title": "Chyba",

	"modal.cancel": "Zrušit",

	"modal.confirm": "Potvrdit",

	"not_found.status": "404",

	"not_found.title": "Stránka neexistuje",

	"not_found.description": "Stránka na této adrese neexistuje.",

	"navigation.back_home": "Zpět na úvod",

	"language.change": "Změnit jazyk",

	"language.name": "Čeština",

	"footer.gin": "Gin Framework",

	"footer.copyright": "©\u00a0{year} Gin GoKick",

	"mail.home": "Úvod",

	"mail.welcome.subject": "Vítejte v\u00a0Gin GoKick",

	"mail.welcome.preheader": "Účet je založený a\u00a0aplikace na vás čeká.",

	"mail.welcome.heading": "Účet je připravený",

	"mail.welcome.account": "Touto adresou se přihlásíte.",

	"mail.welcome.intro": "Děkujeme za registraci. Po přihlášení vás čeká přehled aplikace.",

	"mail.welcome.open_app": "Otevřít aplikaci",

	"mail.welcome.rules": "Pravidla pro lidi i\u00a0AI agenty a\u00a0všechny kontroly kódu najdete v\u00a0repozitáři.",

	"mail.welcome.rules_link": "Gin GoKick na GitHubu",

	"mail.welcome.gin": "Backend stojí na webovém frameworku Gin.",

	"mail.welcome.gin_link": "Gin na GitHubu",

	"mail.welcome.reason": "E-mail vám přišel, protože si někdo s\u00a0touto adresou založil účet v\u00a0Gin GoKick. Pokud jste to nebyli vy, e-mail ignorujte.",

	"consent.title": "Souhlas s cookies",

	"consent.lead": "Nezbytné cookies zajišťují chod webu. Ostatní nástroje ukládají cookies až s vaším souhlasem. Tlačítkem „{button}“ souhlasíte se všemi, v podrobném nastavení si vyberete. Volbu můžete kdykoli změnit.",

	"consent.details": "Podrobné nastavení",

	"consent.necessary_only": "Přijmout nezbytné",

	"consent.understood": "Rozumím",

	"consent.settings": "Nastavení cookies",

	"consent.settings_lead": "Nezbytné cookies jsou zapnuté vždy, ostatní jen s vaším souhlasem. Šipka u každé skupiny ukáže, k čemu slouží a jaké nástroje používáme.",

	"consent.close": "Zavřít",

	"consent.necessary": "Nezbytné cookies",

	"consent.necessary_text": "Zajišťují chod webu, například přihlášení. Bez nich web nefunguje, a proto jsou vždy zapnuté.",

	"consent.always_on": "Vždy zapnuté",

	"consent.analytics": "Analytické cookies",

	"consent.analytics_text": "Měří, jak návštěvníci web používají. Používáme {tools}.",

	"consent.marketing": "Reklamní a marketingové cookies",

	"consent.marketing_text": "Měří výsledky reklam a pomáhají je cílit. Používáme {tools}.",

	"consent.save": "Uložit nastavení",

	"consent.accept_all": "Přijmout vše",

	"tracking.ga4": "Google Analytics",

	"tracking.google_ads": "Google Ads",

	"tracking.meta_pixel": "Meta Pixel",

	"fetch.network_error": "Server se nepodařilo zastihnout. Zkontrolujte připojení a zkuste to znovu.",

	"fetch.malformed_body": "Server poslal odpověď, které nerozumíme.",

	"fetch.invalid_shape": "Server poslal odpověď v nečekaném tvaru.",

	"fetch.error_status": "Požadavek se nepodařil, server odpověděl stavem {status}.",

	"request.invalid_body": "Požadavek se nepodařilo přečíst.",

	"request.invalid_query": "Parametry požadavku nejsou platné.",

	"request.body_too_large": "Požadavek je příliš velký.",

	"request.internal": "Na serveru nastala chyba. Zkuste to znovu.",

	"request.not_found": "Požadovaná adresa neexistuje.",

	"request.too_many_attempts": "Příliš mnoho neúspěšných pokusů. Zkuste to znovu později.",

	"validation.required": "Vyplňte toto pole.",

	"validation.email": "Zadejte platnou e-mailovou adresu.",

	"validation.uuid": "Identifikátor nemá platný tvar.",

	"validation.min_length": "Zadejte alespoň {min, plural, one {# znak} few {# znaky} many {# znaku} other {# znaků}}.",

	"validation.max_length": "Zadejte nejvýše {max, plural, one {# znak} few {# znaky} many {# znaku} other {# znaků}}.",

	"validation.min_items": "Zadejte alespoň {min, plural, one {# položku} few {# položky} many {# položky} other {# položek}}.",

	"validation.max_items": "Zadejte nejvýše {max, plural, one {# položku} few {# položky} many {# položky} other {# položek}}.",

	"validation.min": "Nejmenší povolená hodnota je {min}.",

	"validation.max": "Největší povolená hodnota je {max}.",

	"validation.one_of": "Vyberte jednu z hodnot: {values}.",

	"validation.invalid": "Hodnota není platná.",

	"auth.login_failed": "E-mail nebo heslo nesouhlasí.",

	"auth.sign_in_required": "Pro pokračování se přihlaste.",

	"user.email_taken": "Účet s tímto e-mailem už existuje.",
}

func init() { register("cs_CZ", csCZ) }
