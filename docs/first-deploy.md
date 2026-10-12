# First deploy

This guide deploys a new project from this skeleton to a Docker Swarm: Cloudflare, Sentry, GitHub Actions with shared workflows, PostgreSQL in the stack, Prometheus, Loki and Grafana. The Swarm with its reverse proxy Caddy, Prometheus, Loki with Alloy and Grafana already run, and an infra repository describes them; the guide only adds the project to them. A person can follow it by hand, or Claude reads it and does most steps itself through `gh`, the APIs and the infra repository. Skills with a map of the infrastructure or a password manager speed the work up, but the guide doesn't depend on them.

## For Claude

1. Ask which skills and tools the person has: a map of the infrastructure, a password manager with a CLI, signed-in CLIs. Fill in as many inputs of the table below as they give.
2. Show the whole table of inputs to the person for approval. Ask for whatever is missing.
3. Check the tools and access:
   - `curl`, `jq`, `openssl`, `gh` and `ssh` are installed;
   - `gh auth status` shows an account that may set the secrets and variables of `<repo>` and run its workflows;
   - a local clone of `<repo>`, from whose root the guide runs, and a clone of `<infra-repo>`;
   - SSH to `<manager>`;
   - the Sentry API answers `GET <sentry-api>/api/0/`.
4. List the manual steps of "What the person does", with their links. Items 1 to 4 must be done before step 1, item 5 comes in step 1, item 6 at the end.
5. Go through steps 1 to 7. The last step is the verification.
6. Describe every write to production (the secrets and variables of `<repo>`, a run of Deploy, `<infra-repo>`, SSH to the servers, Sentry) beforehand and have it confirmed. A change of `<infra-repo>` goes through its own pull request and deploy, as its README says.
7. Never print tokens and passwords, and never ask for them in the chat.
   - Read tokens from exported environment variables (`SENTRY_TOKEN`, `GRAFANA_TOKEN`) or from the password manager, and pass them into a header through `-H @<(printf …)`.
   - Generate a password and write it to `<repo>` in the same command (`openssl rand -hex 24 | gh secret set …`), so it never shows in the output or in the arguments of a process. GitHub never shows a secret again.
8. Every Bash call is a new shell, so define the helper functions again in each command.
9. The answers of the Sentry API and `docker service inspect` contain secrets (keys, the environment of a service). Print only chosen fields of them through `jq` or `--format`.

## Inputs

| Placeholder | What it is | Example |
| --- | --- | --- |
| `<projectname>` | name of the project, which names the database, its role, and the Sentry projects | `shop` |
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
| `<sentry-org>`, `<sentry-team>` | slug of the organization and of the team in Sentry | `firma`, `firma` |
| `<sentry-api>` | the Sentry API, `https://de.sentry.io` for organizations in the EU | `https://sentry.io` |

The Sentry links below hold for sentry.io. A self-hosted Sentry has the same pages at its own address.

## How the deploy works

`<repo>` adds two short workflows that call the shared ones of `<workflows>`. These run on self-hosted runners, which hold the login to `<registry>` and the Docker socket of a manager of the Swarm. The skeleton itself carries neither caller, because its public repository would queue jobs for runners it has no access to.

- **Build** runs on every push to `<branch>`. It builds `docker/production/Dockerfile` and pushes `<registry>/<owner>/<repo>:<commit>`, in lowercase and with the whole SHA.
- **Deploy** starts by the Run workflow button in the Actions tab. It deploys `docker/production/compose.yaml` of the chosen commit with the image of the same commit, as the stack `<stack>`. An empty version takes the head of the branch, and an older commit rolls back.
- **What the compose gets:** `IMAGE`, `VERSION` (the commit), `STACK`, `ROLE` (the services run on the nodes with `node.labels.role == ${ROLE}`, by default `app`), and every repository variable and secret by its name for `${…}`. A value written into `environment` shows in `docker service inspect`, so only configuration goes there.
- **Secrets.** A top-level secret with `x-environment: <NAME>` becomes a Docker secret with the value of the repository secret `<NAME>`. Its name carries a hash of the value, so a new value deploys as a new secret.
  - `x-optional: true` beside it leaves the secret out when the repository lacks it, and the variable stays unset.
  - The entrypoint of `app` exports every file of `/run/secrets` as a variable of its name and then starts the app, so the app needs no `_FILE` support. `db` reads its passwords through `POSTGRES_PASSWORD_FILE` and `DB_PASSWORD_FILE`, which the PostgreSQL image and `app-role.sh` support.
  - The deploy refuses a stack in whose rendered compose the value of any repository secret appears. The compose therefore interpolates only repository variables, and `SENTRY_DSN` or `SMTP_USERNAME` are never secrets.
- **A broken version.** The app updates `start-first` with `failure_action: rollback`. When the new version doesn't start or doesn't pass the healthcheck of the image on `/readyz`, Swarm brings the previous version back and the run fails. The database updates `stop-first`, since two PostgreSQL can't share one volume, and a new app waits up to a minute for it.
- **The database role.** The first start of `db` on an empty volume runs `docker/postgres/initdb/app-role.sh`, which creates the role `DB_USERNAME` with `DB_PASSWORD` and the database `DB_NAME`.
- **Traces, metrics and logs** need no step, when `<infra-repo>` runs an OTLP receiver named `otel` on `edge` that passes metrics to Prometheus and logs to Loki. The compose sends all three there with `service.namespace=<stack>`, the name of the Swarm service, `<stack>_app`, as `service.name`, and the server as `host.name`, so traces, metrics and logs share one name.
  - Tempo finds a trace by the `X-Trace-Id` of any response, and Prometheus labels the metrics `service_namespace="<stack>"`.
  - Loki keeps the logs under `{service_namespace="<stack>", service_name="<stack>_app"}`. The infra turns `host.name` into the label `server`, which the logs of the other containers carry.
  - Alloy of the infra sends the stdout of every other container to Loki. The label `logs: otlp` of `app` tells it to skip the app, whose lines would come twice.
  - The repository variable `OTEL_EXPORTER_OTLP_ENDPOINT` points them to another receiver, `OTEL_TRACES_SAMPLER_ARG` sends only a share of the traces, and the optional secret `OTEL_EXPORTER_OTLP_HEADERS` carries its key.

## What the person does

These steps can't go through an API, or the token would pass through the chat. Give Claude the tokens as environment variables exported before it starts, or through a password manager with a CLI.

1. **Sentry, personal token** for the configuration, on `https://<sentry-org>.sentry.io/settings/account/api/auth-tokens/`, into `SENTRY_TOKEN`. Permissions:
   - Project: Admin, to create projects;
   - Release: Admin;
   - Alerts: Write;
   - Organization: Read and Integrations, for the code mappings;
   - Issue & Event: Write, to resolve the test error.

   What the token can really do shows in `GET <sentry-api>/api/0/`, in the field `auth.scopes`.
2. **Sentry, GitHub:** `https://<sentry-org>.sentry.io/settings/integrations/github/` → Add Installation for the account that owns `<repo>`.
3. **Grafana:** a service account token allowed to import dashboards and query the datasources (Administration → Service accounts) into `GRAFANA_TOKEN`.
4. **GitHub:** an owner of `<workflows>` lets `<repo>` call its workflows (Settings → Actions → General → Access) and use the self-hosted runners (the repository access of their runner group).
5. **Cloudflare:** the redirect rule of step 1 is done in the panel.
6. **hstspreload.org:** submit the domain by hand once no subdomain runs on HTTP only.

## Helper functions

The steps below rely on this function (bash or zsh):

```bash
sentry_api() { curl -sS --fail-with-body -H @<(printf 'Authorization: Bearer %s' "$SENTRY_TOKEN") -H 'Content-Type: application/json' "$@"; }
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

## 2. Sentry

Sentry must be ready before step 3, which writes the DSNs of both projects into the repository variables.

The shared Build passes the Dockerfile no build arguments and no build secret, so the image uploads no source maps and creates no release with commits and a deploy. A build that passes `SENTRY_ORG`, `SENTRY_PROJECT` and `SENTRY_REPOSITORY` with the build secret `SENTRY_AUTH_TOKEN` (an organization token) uploads them. Until then the frames of the frontend stay minified, and the Vue code mapping of item 3 waits for them.

1. Two projects, the backend and the frontend:

   ```bash
   sentry_api -X POST -d '{"name":"<projectname>-go","platform":"go"}' "<sentry-api>/api/0/teams/<sentry-org>/<sentry-team>/projects/" | jq '{slug, id}'
   sentry_api -X POST -d '{"name":"<projectname>-vue","platform":"javascript-vue"}' "<sentry-api>/api/0/teams/<sentry-org>/<sentry-team>/projects/" | jq '{slug, id}'
   ```

2. The alert. The default rule of a new project reports only high-priority issues. Replace it with this one:
   - a new issue, a regression, a reappeared issue, a new or escalating high-priority issue;
   - level warning and above, every environment;
   - an email to the owner of the token;
   - at most once per 30 minutes.

   In the API, alerts are workflows; the old `projects/…/rules/` answers 404. Find the rule of a project through its detector with the project's `projectId`, never by the type of its action. The organization-wide rule of Seer has the same actions and would be overwritten.

   ```bash
   me=$(sentry_api "<sentry-api>/api/0/" | jq -r .user.id)
   for p in <projectname>-go <projectname>-vue; do
     pid=$(sentry_api "<sentry-api>/api/0/projects/<sentry-org>/$p/" | jq -r .id)
     det=$(sentry_api "<sentry-api>/api/0/organizations/<sentry-org>/detectors/" | jq -r --arg p "$pid" '.[] | select(.projectId == $p and .type == "issue_stream") | .id')
     wf=$(sentry_api "<sentry-api>/api/0/organizations/<sentry-org>/workflows/" | jq -r --arg d "$det" '.[] | select(.detectorIds | index($d)) | .id')
     [ -n "$det" ] && [ -n "$wf" ] || { echo "$p: rule not found"; exit 1; }
     jq -n --arg me "$me" --arg det "$det" '{name: "Send all issues", enabled: true, environment: null, config: {frequency: 30}, detectorIds: [$det],
       triggers: {logicType: "any-short", conditions: [
         {type: "first_seen_event", comparison: true, conditionResult: true},
         {type: "regression_event", comparison: true, conditionResult: true},
         {type: "reappeared_event", comparison: true, conditionResult: true},
         {type: "new_high_priority_issue", comparison: true, conditionResult: true},
         {type: "existing_high_priority_issue", comparison: true, conditionResult: true}]},
       actionFilters: [{logicType: "any-short",
         conditions: [{type: "level", comparison: {level: 30, match: "gte"}, conditionResult: true}],
         actions: [{type: "email", data: {}, config: {targetType: "user", targetIdentifier: $me}}]}]}' \
       | sentry_api -X PUT --data-binary @- "<sentry-api>/api/0/organizations/<sentry-org>/workflows/$wf/" | jq -c '{id, name, frequency: .config.frequency}'
   done
   ```

3. Code mappings, so frames lead to the release's commit on GitHub. The repository must show in Sentry once GitHub is installed; otherwise stop and go back to manual item 2.
   - **Go:** the binary is built with `-trimpath`, and its frames have the path `<Go module>/app/…`. The stack root is therefore the module from `go.mod` with a slash, and the source root is empty.
   - **Vue:** after the source maps, the frames have the path `../assets/…`, so the stack root is `../`. Check it after the first error of a build with source maps.

   ```bash
   repos=$(sentry_api "<sentry-api>/api/0/organizations/<sentry-org>/repos/?query=$(basename <repo>)")
   repo=$(jq -r '.[] | select(.name == "<repo>") | .id' <<<"$repos")
   integration=$(jq -r '.[] | select(.name == "<repo>") | .integrationId' <<<"$repos")
   [ -n "$repo" ] || { echo "<repo> is not in Sentry"; exit 1; }
   module=$(awk '/^module /{print $2}' go.mod)
   for pair in "<projectname>-go:$module/" "<projectname>-vue:../"; do
     pid=$(sentry_api "<sentry-api>/api/0/projects/<sentry-org>/${pair%%:*}/" | jq -r .id)
     jq -n --arg p "$pid" --arg r "$repo" --arg i "$integration" --arg s "${pair#*:}" '{projectId: $p, repositoryId: $r, integrationId: $i, stackRoot: $s, sourceRoot: "", defaultBranch: "<branch>"}' \
       | sentry_api -X POST --data-binary @- "<sentry-api>/api/0/organizations/<sentry-org>/code-mappings/" | jq -c '{id, stackRoot, repo: .repoName}'
   done
   ```

4. Source code from GitHub on the Go frames. The binary carries no sources, so Sentry loads them from the repository:

   ```bash
   sentry_api -X PUT -d '{"scmSourceContextEnabled": true}' "<sentry-api>/api/0/projects/<sentry-org>/<projectname>-go/" | jq '{slug, scmSourceContextEnabled}'
   ```

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
   dsn() { sentry_api "<sentry-api>/api/0/projects/<sentry-org>/$1/keys/" | jq -r '[.[] | select(.isActive)][0].dsn.public'; }
   gh variable set APP_DOMAIN --repo <repo> --body '<domain>'
   gh variable set DEFAULT_LOCALE --repo <repo> --body '<locale>'
   gh variable set LOCALES --repo <repo> --body '<locales>'
   gh variable set DB_NAME --repo <repo> --body '<projectname>'
   gh variable set DB_USERNAME --repo <repo> --body '<projectname>'
   gh variable set SENTRY_DSN --repo <repo> --body "$(dsn <projectname>-go)"
   gh variable set SENTRY_FRONTEND_DSN --repo <repo> --body "$(dsn <projectname>-vue)"
   gh variable list --repo <repo> --json name --jq '.[].name'
   ```

   Mails stay off until `SMTP_ENABLED=true` with `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME` and `SMTP_SENDER_EMAIL`. The measurement tools turn on with their IDs.
3. **The secrets.** The compose mounts `DB_PASSWORD` into the app and the database and `POSTGRES_PASSWORD` into the database, and both are required. `SMTP_PASSWORD` and `OTEL_EXPORTER_OTLP_HEADERS` are optional. None of them is ever a variable.

   ```bash
   openssl rand -hex 24 | tr -d '\n' | gh secret set DB_PASSWORD --repo <repo>
   openssl rand -hex 24 | tr -d '\n' | gh secret set POSTGRES_PASSWORD --repo <repo>
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
     | ssh <db-server> 'docker exec -i $(docker ps -q -f name=<stack>_db | head -1) psql -v ON_ERROR_STOP=1 -U postgres -d postgres -f -' \
     && printf '%s' "$PW" | gh secret set DB_PASSWORD --repo <repo>
   ```

   `POSTGRES_PASSWORD` counts only on an empty volume too. A new value of the secret restarts `db` on the next deploy and leaves the password of the superuser as it was.

## 5. Backups

The label `backup: pgbackrest` under `deploy.labels` of `db` asks the deploy for point-in-time backups.
- The deploy sets `PGBACKREST_REPO1_PATH` to `/<stack>/db` and mounts the shared backup keys into `db` only. The keys must exist in the Swarm, or the deploy fails; a Swarm without them needs the label removed.
- Only a PostgreSQL image with pgBackRest uses them. The default `postgres:18.6-alpine` ignores them, and the database runs without backups.

To switch them on, set the variable `POSTGRES_IMAGE` to the PostgreSQL image with pgBackRest that `<infra-repo>` builds and run Deploy. That image starts pgBackRest from its own entrypoint and leaves the secrets with a `_FILE` variable to PostgreSQL.

## 6. Grafana

[grafana-dashboard.json](grafana-dashboard.json) has placeholders instead of the project name:
- `<projectname>` in the `uid` and in the tag in `tags`;
- `<stack>` in the default of the variable `project`, which picks the stack in Prometheus and Loki;
- `<domain>` in the `title`.

The metrics carry `service_namespace` from the OTLP export, and the row Logs reads Loki by `service_namespace` and `service_name="<stack>_app"`. The panels of CPU, container memory and the time since the last start read cAdvisor, which needs the same two labels on the containers. Series pushed over OTLP get no staleness marker, so the panels expect Prometheus with `--query.lookback-delta=2m`, which drops a stopped replica after 4 pushes, and its Grafana datasource with `timeInterval: 30s`, so `$__rate_interval` spans at least 4 pushes. The file in the repository stays as it is, and the import fills in the values. The `sed` patterns are written as `[<]…[>]`, so nobody mistakes them for placeholders of this guide.

```bash
dashboard=$(sed -e 's/[<]projectname[>]/<projectname>/g' -e 's/[<]stack[>]/<stack>/g' -e 's/[<]domain[>]/<domain>/g' docs/grafana-dashboard.json)
grep -c '[<>]' <<<"$dashboard"
jq '{dashboard: (. + {id: null}), overwrite: true, folderUid: ""}' <<<"$dashboard" \
  | curl -sS --fail-with-body -X POST -H @<(printf 'Authorization: Bearer %s' "$GRAFANA_TOKEN") -H 'Content-Type: application/json' \
      --data-binary @- <grafana>/api/dashboards/db | jq '{status, url}'
```

`grep -c` must print 0. A Grafana that loads its dashboards from the files of `<infra-repo>` takes the filled file there instead.

## 7. Verification

`<SHA>` is the deployed commit, in the `data-release` of the tag `<meta name="sentry">` on `https://<domain>/<language>/app`. Claude checks these itself:

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

- securityheaders.com and MDN Observatory show A+.
- Only with a build that uploads source maps:
  - The release in Sentry has its commits and a deploy to `production`:

    ```bash
    sentry_api "<sentry-api>/api/0/organizations/<sentry-org>/releases/<SHA>/" | jq '{version, commitCount, deployCount, projects: [.projects[].slug]}'
    ```

  - Sentry finds the `sentry-dbid-…` of the live `/build/assets/app-….js`, and the map never leaks out:

    ```bash
    js=$(curl -sS https://<domain>/<language>/app | grep -oE '/build/assets/app-[A-Za-z0-9_-]+\.js' | head -1)
    id=$(curl -sS "https://<domain>$js" | grep -oE 'sentry-dbid-[0-9a-f-]{36}' | head -1 | cut -c13-)
    sentry_api "<sentry-api>/api/0/projects/<sentry-org>/<projectname>-vue/artifact-lookup/?debug_id=$id" | jq 'length'
    curl -s -o /dev/null -w '%{http_code}\n' "https://<domain>$js.map"
    ```

    The first command must print at least 1, the second 404.

With the person's permission, because an issue is created and an email arrives:

- **A test error of the frontend.** On `https://<domain>/<language>/app`, run the code below in the browser console. A Claude with a browser does it itself, otherwise the person does. The error arises in `assets/shared/Sentry/startSentry.ts`, so it tests the source maps on our own code.

  ```js
  const probe = new Event('securitypolicyviolation');
  Object.defineProperty(probe, 'sourceFile', { get() { throw new Error('Sentry probe'); } });
  setTimeout(() => document.dispatchEvent(probe));
  ```

  Then check the event and close the issue:
  - The event in `<projectname>-vue` must exist. With source maps, it has the frame `../assets/shared/Sentry/startSentry.ts` with its source line, and `GET <sentry-api>/api/0/projects/<sentry-org>/<projectname>-vue/stacktrace-link/?file=../assets/shared/Sentry/startSentry.ts&platform=javascript&commitId=<SHA>` answers a `sourceUrl` into GitHub.
  - Then resolve the issue: `PUT <sentry-api>/api/0/organizations/<sentry-org>/issues/<issue id>/` with `{"status":"resolved"}`. The `<issue id>` is the `groupID` of the event.
- **Frontend and backend errors joined.** In production this would need a database outage, so only on request. Locally it can be tried like this:
  - a binary built as in the image (`go build -trimpath`);
  - the real DSNs and `APP_ENVIRONMENT=development`;
  - a temporary Postgres that you stop after the start.

  `fetch('/readyz')` from the frontend then answers 503, and the backend error has the same `trace_id` as the frontend error from the same page.
- **Go code in Sentry** loads from GitHub only in the UI; the API can't check it. Open a backend issue and check that its frames show lines of code.
