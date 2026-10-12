# First deploy

This guide deploys a new project from this skeleton to a Docker Swarm: Cloudflare, GlitchTip, GitHub Actions with shared workflows, PostgreSQL in the stack, Prometheus, Loki, Tempo and Grafana. The Swarm with its reverse proxy Caddy, Prometheus, Loki and Tempo with Alloy and Grafana already run, and an infra repository describes them; the guide only adds the project to them. A person can follow it by hand, or Claude reads it and does most steps itself through `gh`, the APIs and the infra repository. Skills with a map of the infrastructure or a password manager speed the work up, but the guide doesn't depend on them.

## For Claude

1. Ask which skills and tools the person has: a map of the infrastructure, a password manager with a CLI, signed-in CLIs. Fill in as many inputs of the table below as they give.
2. Show the whole table of inputs to the person for approval. Ask for whatever is missing.
3. Check the tools and access:
   - `curl`, `jq`, `openssl`, `gh` and `ssh` are installed;
   - `gh auth status` shows an account that may set the secrets and variables of `<repo>` and run its workflows;
   - a local clone of `<repo>`, from whose root the guide runs, and a clone of `<infra-repo>`;
   - SSH to `<manager>`;
   - the GlitchTip API answers `GET <glitchtip>/api/0/` with the token of manual item 1. When an access proxy guards GlitchTip, `<infra-repo>` says how to reach its API.
4. List the manual steps of "What the person does", with their links. Items 1 to 4 must be done before step 1, item 5 comes in step 1, item 6 at the end.
5. Go through steps 1 to 7. The last step is the verification.
6. Describe every write to production (the secrets and variables of `<repo>`, a run of Deploy, `<infra-repo>`, SSH to the servers, GlitchTip) beforehand and have it confirmed. A change of `<infra-repo>` goes through its own pull request and deploy, as its README says.
7. Never print tokens and passwords, and never ask for them in the chat.
   - Read tokens from exported environment variables (`GLITCHTIP_TOKEN`, `GRAFANA_TOKEN`) or from the password manager, and pass them into a header through `-H @<(printf …)`.
   - Generate a password and write it to `<repo>` in the same command (`openssl rand -hex 24 | gh secret set …`), so it never shows in the output or in the arguments of a process. GitHub never shows a secret again.
8. Every Bash call is a new shell, so define the helper functions again in each command.
9. The answers of the GlitchTip API and `docker service inspect` contain secrets (keys, the environment of a service). Print only chosen fields of them through `jq` or `--format`.

## Inputs

| Placeholder | What it is | Example |
| --- | --- | --- |
| `<projectname>` | name of the project, which names the database, its role, and the GlitchTip projects | `shop` |
| `<domain>` | domain of the site | `shop.cz` |
| `<repo>` | GitHub repository `owner/name` | `firma/shop` |
| `<stack>` | Swarm stack: the name of `<repo>` in lowercase, unless the Deploy caller sets `stack` | `shop` |
| `<branch>` | branch Build watches and Deploy takes by default | `main` |
| `<locale>` | default language of the site (`DEFAULT_LOCALE`); its language is `<language>` in the SPA addresses | `cs_CZ` |
| `<locales>` | allowed languages (`LOCALES`), `<locale>` included | `cs_CZ,en_US` |
| `<workflows>` | repository `owner/name` with the shared workflows `build.yml` and `deploy.yml` | `firma/workflows` |
| `<registry>` | registry Build pushes the images to | `registry.firma.cz` |
| `<infra-repo>` | path to a clone of the infra repository: DNS, Caddy, Prometheus and the Swarm | `~/infra` |
| `<manager>` | SSH alias of a manager of the Swarm | `ops1` |
| `<grafana>` | address of Grafana | `https://grafana.firma.cz` |
| `<prometheus-datasource>`, `<loki-datasource>` | uids of the Prometheus and Loki datasources in Grafana | `prometheus`, `loki` |
| `<glitchtip>` | address of GlitchTip | `https://glitchtip.firma.cz` |
| `<glitchtip-org>`, `<glitchtip-team>` | slug of the organization and of a team in GlitchTip | `firma`, `firma` |

## How the deploy works

`<repo>` adds two short workflows that call the shared ones of `<workflows>`. These run on self-hosted runners, which hold the login to `<registry>` and the Docker socket of a manager of the Swarm. The skeleton itself carries neither caller, because its public repository would queue jobs for runners it has no access to.

- **Build** runs on every push to `<branch>`. It builds `docker/production/Dockerfile` and pushes `<registry>/<owner>/<repo>:<commit>`, in lowercase and with the whole SHA. A repository variable named by an `ARG` of the Dockerfile becomes a build argument, and the repository secret `GLITCHTIP_AUTH_TOKEN`, which the Dockerfile mounts, becomes a build secret.
- **Deploy** starts by the Run workflow button in the Actions tab. It deploys `docker/production/compose.yaml` of the chosen commit with the image of the same commit, as the stack `<stack>`. An empty version takes the head of the branch, and an older commit rolls back.
- **What the compose gets:** `IMAGE`, `VERSION` (the commit), `STACK`, `ROLE` (the services run on the nodes with `node.labels.role == ${ROLE}`, by default `app`), and every repository variable and secret by its name for `${…}`. A value written into `environment` shows in `docker service inspect`, so only configuration goes there.
- **Secrets.** A top-level secret with `x-environment: <NAME>` becomes a Docker secret with the value of the repository secret `<NAME>`. Its name carries a hash of the value, so a new value deploys as a new secret.
  - `x-optional: true` beside it leaves the secret out when the repository lacks it, and the variable stays unset.
  - The entrypoint of `app` exports every file of `/run/secrets` as a variable of its name and then starts the app, so the app needs no `_FILE` support. `db` reads its passwords through `POSTGRES_PASSWORD_FILE`, `DB_PASSWORD_FILE` and `DB_MONITORING_PASSWORD_FILE`, which the PostgreSQL image and the scripts of `docker/postgres` support, and `db-exporter` through `DATA_SOURCE_PASS_FILE`.
  - The deploy refuses a stack in whose rendered compose the value of any repository secret appears. The compose therefore interpolates only repository variables, and `GLITCHTIP_DSN` or `SMTP_USERNAME` are never secrets.
- **A broken version.** The app updates `start-first` with `failure_action: rollback`. When the new version doesn't start or doesn't pass the healthcheck of the image on `/readyz`, Swarm brings the previous version back and the run fails. The database updates `stop-first`, since two PostgreSQL can't share one volume, and a new app waits up to a minute for it.
- **The database roles.** The first start of `db` on an empty volume runs `docker/postgres/initdb/app-role.sh`, which creates the role `DB_USERNAME` with `DB_PASSWORD` and the database `DB_NAME`. Then `docker/postgres/monitoring-role.sh` creates the role `monitoring` with `DB_MONITORING_PASSWORD` and `pg_monitor`, and the extension `pg_stat_statements` in `DB_NAME`, for `db-exporter`.
- **Logs** need no step. Alloy of the infra sends the stdout of every container to Loki with the labels `service_namespace` (the stack) and `service_name` (the name of the Swarm service, `<stack>_<service>`), so the app logs under `{service_namespace="<stack>", service_name="<stack>_app"}`.
- **Traces and metrics** need no step either, when `<infra-repo>` runs an OTLP receiver named `otel` on `edge` that passes metrics to Prometheus. The compose sends both there with `service.namespace=<stack>` and the name of the Swarm service, `<stack>_app`, as `service.name`, so traces, metrics and logs share one name. Tempo finds a trace by the `X-Trace-Id` of any response, and Prometheus labels the metrics `service_namespace="<stack>"`. The repository variable `OTEL_EXPORTER_OTLP_ENDPOINT` points them to another receiver, `OTEL_TRACES_SAMPLER_ARG` sends only a share of the traces, and the optional secret `OTEL_EXPORTER_OTLP_HEADERS` carries its key.

## What the person does

These steps can't go through an API, or the token would pass through the chat. Give Claude the tokens as environment variables exported before it starts, or through a password manager with a CLI.

1. **GlitchTip, token for the configuration**, under Profile → Auth Tokens on `<glitchtip>`, into `GLITCHTIP_TOKEN`. Scopes:
   - `project:read`, `project:write` and `project:admin`, for the projects, their keys and alerts;
   - `team:read`, and `team:write` when `<glitchtip-org>` has no team yet;
   - `org:read`;
   - `event:read` and `event:write`, to read and resolve the test error.

   What the token can do shows in `GET <glitchtip>/api/0/`, in the field `auth.scopes`.
2. **GlitchTip, token for the build** with the single scope `project:releases`, which uploads the source maps. The person pastes it into `gh secret set GLITCHTIP_AUTH_TOKEN --repo <repo>`, which asks for it without showing it. The token belongs to its person, so a new one is needed when the person leaves the organization.
3. **Grafana:** a service account token allowed to import dashboards and query the datasources (Administration → Service accounts) into `GRAFANA_TOKEN`.
4. **GitHub:** an owner of `<workflows>` lets `<repo>` call its workflows (Settings → Actions → General → Access) and use the self-hosted runners (the repository access of their runner group).
5. **Cloudflare:** the redirect rule of step 1 is done in the panel.
6. **hstspreload.org:** submit the domain by hand once no subdomain runs on HTTP only.

## Helper functions

The steps below rely on this function (bash or zsh):

```bash
glitchtip_api() { curl -sS --fail-with-body -H @<(printf 'Authorization: Bearer %s' "$GLITCHTIP_TOKEN") -H 'Content-Type: application/json' "$@"; }
```

## 1. Cloudflare and Caddy

In `<infra-repo>`, wherever it keeps them:

1. DNS: records `@` and `www` of `<domain>` to the public address of the edge, with the proxy of Cloudflare on.
2. TLS: the SSL/TLS mode Full, with the certificate the infra gives Caddy for the zone.
3. Caddy: a site that proxies `<domain>` to the app over the network `edge`, with the TLS and the Cloudflare-only rule of the other sites:

   ```caddy
   <domain> {
   	reverse_proxy <stack>_app:8020
   }
   ```

   Caddy passes `CF-Connecting-IP` unchanged, and gin takes the client IP from it.

In the Cloudflare panel for the zone `<domain>`, Rules → Redirect Rules → Single Redirect:
- the condition `http.host eq "www.<domain>"`;
- the dynamic target `concat("https://<domain>", http.request.uri.path)`;
- 301 and Preserve query string.

## 2. GlitchTip

GlitchTip must be ready before step 3, which writes the DSNs of both projects into the repository variables. The Sentry SDKs of the backend and the frontend send there, because GlitchTip takes the protocol of Sentry.

1. Two projects, the backend and the frontend. A project belongs to a team; when `<glitchtip-org>` has none, the first command creates `<glitchtip-team>` with the owner of the token in it:

   ```bash
   glitchtip_api -X POST -d '{"slug":"<glitchtip-team>"}' "<glitchtip>/api/0/organizations/<glitchtip-org>/teams/" | jq '{slug}'
   glitchtip_api -X POST -d '{"name":"<projectname>-go","platform":"go"}' "<glitchtip>/api/0/teams/<glitchtip-org>/<glitchtip-team>/projects/" | jq '{slug, id}'
   glitchtip_api -X POST -d '{"name":"<projectname>-vue","platform":"javascript-vue"}' "<glitchtip>/api/0/teams/<glitchtip-org>/<glitchtip-team>/projects/" | jq '{slug, id}'
   ```

2. The alert. A new project has none. This one emails the members of the project's teams once for each new issue, after its first event:

   ```bash
   for p in <projectname>-go <projectname>-vue; do
     glitchtip_api -X POST -d '{"name":"Every new issue","timespanMinutes":1,"quantity":1,"uptime":false,"alertRecipients":[{"recipientType":"email"}]}' \
       "<glitchtip>/api/0/projects/<glitchtip-org>/$p/alerts/" | jq -c '{id, name}'
   done
   ```

Compared with Sentry, GlitchTip loads no source code from GitHub and links no commits to a release. A frame of Go shows its file and line, and a frame of Vue its source line from the source maps.

## 3. The repository: workflows, variables and secrets

1. **The callers.** Add both files to `.github/workflows/` of `<repo>` and merge them like any other change. Build then runs on the merge.

   `build.yml`:

   ```yaml
   name: Build

   on:
     push:
       branches:
         - <branch>
     workflow_dispatch:

   permissions:
     contents: read

   jobs:
     build:
       uses: <workflows>/.github/workflows/build.yml@v1
       secrets: inherit
   ```

   `deploy.yml`:

   ```yaml
   name: Deploy

   on:
     workflow_dispatch:
       inputs:
         version:
           description: Commit to deploy, empty for the head of the branch
           required: false

   permissions:
     contents: read

   jobs:
     deploy:
       uses: <workflows>/.github/workflows/deploy.yml@v1
       with:
         version: ${{ inputs.version }}
       secrets: inherit
   ```

   Build also takes `dockerfile` and `context`, and Deploy `compose`, `stack` and `role`. The defaults fit this skeleton.
2. **The variables.** The compose names every variable it takes. It requires the ones written `${NAME:?}`, and an empty one means the default of the app. `.env.example` says what each means. The compose fixes the rest: release mode, JSON logs at `info`, the environment `production`, the release `VERSION`, the database `db`, `TRUSTED_PROXIES` and `APP_URL` from `APP_DOMAIN`.

   ```bash
   dsn() { glitchtip_api "<glitchtip>/api/0/projects/<glitchtip-org>/$1/keys/" | jq -r '.[0].dsn.public'; }
   gh variable set APP_DOMAIN --repo <repo> --body '<domain>'
   gh variable set DEFAULT_LOCALE --repo <repo> --body '<locale>'
   gh variable set LOCALES --repo <repo> --body '<locales>'
   gh variable set DB_NAME --repo <repo> --body '<projectname>'
   gh variable set DB_USERNAME --repo <repo> --body '<projectname>'
   gh variable set GLITCHTIP_DSN --repo <repo> --body "$(dsn <projectname>-go)"
   gh variable set GLITCHTIP_FRONTEND_DSN --repo <repo> --body "$(dsn <projectname>-vue)"
   gh variable set GLITCHTIP_URL --repo <repo> --body '<glitchtip>'
   gh variable set GLITCHTIP_ORG --repo <repo> --body '<glitchtip-org>'
   gh variable set GLITCHTIP_PROJECT --repo <repo> --body '<projectname>-vue'
   gh variable list --repo <repo> --json name --jq '.[].name'
   ```

   Mails stay off until `SMTP_ENABLED=true` with `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME` and `SMTP_SENDER_EMAIL`. The measurement tools turn on with their IDs.
3. **The secrets.** The compose mounts `DB_PASSWORD` into the app and the database, `POSTGRES_PASSWORD` into the database and `DB_MONITORING_PASSWORD` into the database and its exporter, and all three are required. `SMTP_PASSWORD` and `OTEL_EXPORTER_OTLP_HEADERS` are optional. None of them is ever a variable. `GLITCHTIP_AUTH_TOKEN` from manual item 2 goes only to Build; with it, every build uploads the source maps, and a failed upload fails the build.

   ```bash
   openssl rand -hex 24 | tr -d '\n' | gh secret set DB_PASSWORD --repo <repo>
   openssl rand -hex 24 | tr -d '\n' | gh secret set POSTGRES_PASSWORD --repo <repo>
   openssl rand -hex 24 | tr -d '\n' | gh secret set DB_MONITORING_PASSWORD --repo <repo>
   gh secret list --repo <repo> --json name --jq '.[].name'
   ```

   The SMTP password comes from the person or the password manager, piped the same way into `gh secret set SMTP_PASSWORD`.

## 4. Build and deploy

1. Wait for Build of the merged callers. It must end green:

   ```bash
   run=$(gh run list --repo <repo> --workflow build.yml --branch <branch> --limit 1 --json databaseId --jq '.[0].databaseId')
   gh run watch "$run" --repo <repo> --exit-status
   ```

2. Start Deploy and wait for it. Its summary lists the services with their replicas, and it fails when a service rolled back or runs fewer replicas than it should.

   ```bash
   gh workflow run deploy.yml --repo <repo> --ref <branch>
   sleep 5
   run=$(gh run list --repo <repo> --workflow deploy.yml --limit 1 --json databaseId --jq '.[0].databaseId')
   gh run watch "$run" --repo <repo> --exit-status
   ssh <manager> 'docker service logs --since 10m <stack>_app 2>&1' | grep -cE 'migration applied|listening'
   ```

   The log must show `migration applied` and `listening`, and `https://<domain>/readyz` must answer 200 once step 1 is deployed.
3. **Later deploys.** Every push to `<branch>` builds an image, and Deploy runs by hand. To roll back, run Deploy with the commit of the last good version.
4. **Optional: change the database password.** The password of the role lives in the secret `DB_PASSWORD`, and Postgres keeps its hash. `app-role.sh` runs only on an empty volume, so a change is an `ALTER ROLE` and a new secret in the same command, then a run of Deploy. `<db-server>` is the server `ssh <manager> "docker service ps <stack>_db --filter desired-state=running --format '{{.Node}}'"` names.
   - Until the deploy, open connections keep working, and new ones of the running app can't connect.
   - `db` mounts the same secret, so the deploy restarts it too, and every connection drops for a moment.

   ```bash
   PW=$(openssl rand -hex 24)
   printf "ALTER ROLE \"<projectname>\" PASSWORD '%s';\n" "$PW" \
     | ssh <db-server> 'docker exec -i $(docker ps -q -f label=com.docker.swarm.service.name=<stack>_db) psql -v ON_ERROR_STOP=1 -U postgres -d postgres -f -' \
     && printf '%s' "$PW" | gh secret set DB_PASSWORD --repo <repo>
   ```

   `POSTGRES_PASSWORD` counts only on an empty volume too. A new value of the secret restarts `db` on the next deploy and leaves the password of the superuser as it was.
5. **A database from before the exporter.** A volume created by an older compose has neither the role `monitoring` nor `pg_stat_statements`, so `db-exporter` can't sign in. The deploy of this compose restarts `db` with the library loaded and without utility statements in it, so the password of the role never reaches the statistics. Then the script of the compose adds both, as the superuser over the socket of the container:

   ```bash
   ssh <db-server> 'docker exec $(docker ps -q -f label=com.docker.swarm.service.name=<stack>_db) /docker-entrypoint-initdb.d/monitoring-role.sh'
   ```

   It prints `CREATE ROLE`, `GRANT ROLE` and `CREATE EXTENSION`, and the exporter connects at its next scrape.

## 5. Backups

The label `backup: pgbackrest` under `deploy.labels` of `db` asks the deploy for point-in-time backups.
- The deploy sets `PGBACKREST_REPO1_PATH` to `/<stack>/db` and mounts the shared backup keys into `db` only. The keys must exist in the Swarm, or the deploy fails; a Swarm without them needs the label removed.
- Only a PostgreSQL image with pgBackRest uses them. The default `postgres:18.6-alpine` ignores them, and the database runs without backups.

To switch them on, set the variable `POSTGRES_IMAGE` to the PostgreSQL image with pgBackRest that `<infra-repo>` builds and run Deploy. That image starts pgBackRest from its own entrypoint and leaves the secrets with a `_FILE` variable to PostgreSQL.

## 6. Grafana

[grafana/](grafana/) holds the six boards of a project, which [README](README.md#production) describes. They have placeholders instead of the project:
- `<projectname>` in the `uid`, in the tag that links the boards to each other, and in the names of the GlitchTip projects;
- `<stack>` in the default of the variable `project`, which picks the stack in Prometheus, Loki and Tempo;
- `<domain>` in the titles and in the variable `domain`, which picks the site in the metrics of Caddy;
- `<glitchtip>` and `<glitchtip-org>` in the link to the issues.

**What the boards need from `<infra-repo>`**, besides the metrics, logs and traces of the app:
- Prometheus with `--query.lookback-delta=2m` and its Grafana datasource with `timeInterval: 30s`. Series pushed over OTLP get no staleness marker, so the lookback drops a stopped replica after 4 pushes, and `$__rate_interval` spans at least 4 of them.
- Prometheus with `--enable-feature=exemplar-storage`, and the datasource with an `exemplarTraceIdDestinations` named `trace_id` that leads to Tempo.
- A blackbox probe of `http://<stack>_app:8020/readyz` with the label `probe="<stack>"`.
- A scrape of `<stack>_db-exporter:9187` on `edge` with the labels `service_namespace="<stack>"` and `service_name="<stack>_db"`.
- cAdvisor with `service_namespace` and `service_name` on the containers, the metrics `swarm_service_*` by `service_name`, which mark the deploys too, and `pgbackrest_last_backup_timestamp_seconds` and `pgbackrest_last_archive_timestamp_seconds` by `path="/<stack>/db"`.
- The metrics of Caddy by `host`, and `glitchtip_unresolved_issues` and `glitchtip_issues_total` by `project` from GlitchTip.
- Tempo with the processors `span-metrics` and `service-graphs`, which write to Prometheus, and the datasource uids `prometheus`, `loki` and `tempo` as the defaults of the variables.

The files in the repository stay as they are, and the import fills in the values. The `sed` patterns are written as `[<]…[>]`, so nobody mistakes them for placeholders of this guide.

```bash
for board in docs/grafana/*.json; do
  dashboard=$(sed -e 's/[<]projectname[>]/<projectname>/g' -e 's/[<]stack[>]/<stack>/g' -e 's/[<]domain[>]/<domain>/g' \
    -e 's|[<]glitchtip[>]|<glitchtip>|g' -e 's/[<]glitchtip-org[>]/<glitchtip-org>/g' "$board")
  grep -cE '[<](projectname|stack|domain|glitchtip|glitchtip-org)[>]' <<<"$dashboard"
  jq '{dashboard: (. + {id: null}), overwrite: true, folderUid: ""}' <<<"$dashboard" \
    | curl -sS --fail-with-body -X POST -H @<(printf 'Authorization: Bearer %s' "$GRAFANA_TOKEN") -H 'Content-Type: application/json' \
        --data-binary @- <grafana>/api/dashboards/db | jq -c '{status, url}'
done
```

Every `grep -c` must print 0. A Grafana that loads its dashboards from the files of `<infra-repo>` takes the six filled files there instead, in one folder of the project.

## 7. Verification

Claude checks these itself:

- `https://<domain>/` answers 200 with `cf-cache-status: DYNAMIC`, `/readyz` 200, and `https://www.<domain>/` 301 to the apex.
- `<meta name="sentry">` on `https://<domain>/<language>/app` has `data-environment="production"` and a `data-release` equal to the commit of the last run of Deploy.
- The stack runs every service:

  ```bash
  ssh <manager> 'docker service ls --filter label=com.docker.stack.namespace=<stack> --format "{{.Name}} {{.Replicas}}"'
  ```

- Prometheus has the metrics of the app. They arrive every 30 s, and the healthcheck counts too, so the number must be above 0:

  ```bash
  curl -sS --fail-with-body -G -H @<(printf 'Authorization: Bearer %s' "$GRAFANA_TOKEN") \
    "<grafana>/api/datasources/proxy/uid/<prometheus-datasource>/api/v1/query" \
    --data-urlencode 'query=sum(http_server_request_duration_seconds_count{service_namespace="<stack>"})' | jq -r '.data.result[0].value[1]'
  ```

- Loki has the logs of the app. The healthcheck calls `/readyz` every 30 s, so the number must be above 0:

  ```bash
  curl -sS --fail-with-body -G -H @<(printf 'Authorization: Bearer %s' "$GRAFANA_TOKEN") \
    "<grafana>/api/datasources/proxy/uid/<loki-datasource>/loki/api/v1/query" \
    --data-urlencode 'query=sum(count_over_time({service_namespace="<stack>", service_name="<stack>_app"}[5m]))' | jq -r '.data.result[0].value[1]'
  ```

- The exporter reads PostgreSQL, so the query below prints 1:

  ```bash
  curl -sS --fail-with-body -G -H @<(printf 'Authorization: Bearer %s' "$GRAFANA_TOKEN") \
    "<grafana>/api/datasources/proxy/uid/<prometheus-datasource>/api/v1/query" \
    --data-urlencode 'query=min(pg_up{service_namespace="<stack>"})' | jq -r '.data.result[0].value[1]'
  ```

- securityheaders.com and MDN Observatory show A+.
- The live bundle carries a debug ID, and its map never leaks out:

  ```bash
  js=$(curl -sS https://<domain>/<language>/app | grep -oE '/build/assets/app-[A-Za-z0-9_-]+\.js' | head -1)
  curl -sS "https://<domain>$js" | grep -c 'sentry-dbid-'
  curl -s -o /dev/null -w '%{http_code}\n' "https://<domain>$js.map"
  ```

  The first command must print at least 1, the second 404.

With the person's permission, because an issue is created and an email arrives:

- **A test error of the frontend.** On `https://<domain>/<language>/app`, run the code below in the browser console. A Claude with a browser does it itself, otherwise the person does. The error arises in `assets/shared/Sentry/startSentry.ts`, so it tests the source maps on our own code.

  ```js
  const probe = new Event('securitypolicyviolation');
  Object.defineProperty(probe, 'sourceFile', { get() { throw new Error('GlitchTip probe'); } });
  setTimeout(() => document.dispatchEvent(probe));
  ```

  Then check the event and close the issue. The second command must print the frame `../assets/shared/Sentry/startSentry.ts` with `"source":true`, and the email of the alert must arrive:

  ```bash
  issue=$(glitchtip_api "<glitchtip>/api/0/projects/<glitchtip-org>/<projectname>-vue/issues/?query=GlitchTip%20probe" | jq -r '.[0].id')
  glitchtip_api "<glitchtip>/api/0/organizations/<glitchtip-org>/issues/$issue/events/latest/" \
    | jq -c '.entries[] | select(.type == "exception") | .data.values[-1].stacktrace.frames[] | select(.filename | endswith("startSentry.ts")) | {filename, lineNo, source: ((.context // []) | length > 0)}'
  glitchtip_api -X PUT -d '{"status":"resolved"}' "<glitchtip>/api/0/issues/$issue/" | jq '{status}'
  ```

- **A backend error joined with its log and trace.** It needs a short outage of the database, so only on request. Take the login from the role of the app for a second and end its connections, with the login given back by the `trap` whatever happens, and call two routes straight in the app, past the edge. `<db-server>` is the server `ssh <manager> "docker service ps <stack>_db --filter desired-state=running --format '{{.Node}}'"` names, and the app must run there too.

  ```bash
  trace=$(openssl rand -hex 16)
  ssh <db-server> bash -s -- "$trace" <<'SH'
  db=$(docker ps -q -f label=com.docker.swarm.service.name=<stack>_db); app=$(docker ps -qf name=<stack>_app | head -1)
  trap 'docker exec "$db" psql -U postgres -qAt -c "ALTER ROLE \"<projectname>\" LOGIN"' EXIT
  docker exec "$db" psql -U postgres -qAt -v ON_ERROR_STOP=1 -c 'ALTER ROLE "<projectname>" NOLOGIN' \
    -c "SELECT count(pg_terminate_backend(pid)) FROM pg_stat_activity WHERE usename = '<projectname>'"
  docker exec "$app" wget -q -S -O /dev/null --header "traceparent: 00-$1-1a2b3c4d5e6f7a8b-01" --header 'Cookie: __Host-session=probe' http://127.0.0.1:8020/api/user/me 2>&1 | grep 'HTTP/'
  docker exec "$app" wget -q -S -O /dev/null --header "traceparent: 00-$1-2b3c4d5e6f7a8b9c-01" http://127.0.0.1:8020/readyz 2>&1 | grep 'HTTP/'
  SH
  echo "$trace"
  ```

  The routes answer 500 and 503, and `https://<domain>/readyz` answers 200 again right after. The trace ID must then be:
  - in two events of `<projectname>-go`, whose culprits are `app/core/authn/authn.go` and `app/core/health/health.go`;
  - in two `ERROR` lines of the log in Loki;
  - in Tempo, where the span `GET /api/user/me` is failed and has an `exception` event with the stack (`/readyz` sends no spans).

  Then resolve both issues as above.
