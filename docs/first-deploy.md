# First deploy

This guide deploys a new project from this skeleton to your own servers: Cloudflare, Sentry, Dokploy with PostgreSQL, Prometheus, Loki and Grafana. Prometheus, Loki with Alloy and Grafana already run on the servers; the guide only adds the project to them. A person can follow it by hand, or Claude reads it and does most steps itself through the APIs and SSH. Skills with a map of the infrastructure or a password manager speed the work up, but the guide doesn't depend on them.

## For Claude

1. Ask which skills and tools the person has: a map of the infrastructure, a password manager with a CLI, signed-in CLIs. Fill in as many inputs of the table below as they give.
2. Show the whole table of inputs to the person for approval. Ask for whatever is missing.
3. Check the tools and access:
   - `curl`, `jq`, `openssl`, `awk`, `shasum` and `ssh` are installed;
   - a local clone of `<repo>`, from whose root the guide runs;
   - SSH to `<app-server>` and to `<monitoring-server>`;
   - the Dokploy API answers `project.all` and `compose.one` for `<alloy-compose>`, and the Sentry API answers `GET <sentry-api>/api/0/`.
4. List the manual steps of "What the person does", with their links. Items 1 to 5 must be done before step 1, item 6 comes in step 3, item 7 at the end.
5. Go through steps 1 to 7. The last step is the verification.
6. Describe every write to production (Dokploy, SSH to the servers, Sentry) beforehand and have it confirmed.
7. Never print tokens and never ask for them in the chat. Read them from exported environment variables (`SENTRY_TOKEN`, `DOKPLOY_TOKEN`, `GRAFANA_TOKEN`) or from the password manager.
   - Pass them into a header through `-H @<(printf …)`.
   - Pass secrets to `jq` and `awk` through `--rawfile` and `<(printf …)`, not through `--arg`. Then none of them shows in the arguments of a process.
8. Every Bash call is a new shell, so define the helper functions again in each command. Write the database password and the metrics account to Dokploy in the same command that creates them. Later, read them back from `application.one`.
9. The answers of the Dokploy and Sentry APIs contain secrets (the whole env of the app, keys). Print only chosen fields of them through `jq`.

## Inputs

| Placeholder | What it is | Example |
| --- | --- | --- |
| `<projectname>` | name of the project, which names the database, the Sentry projects and the Prometheus job | `shop` |
| `<domain>` | domain of the site | `shop.cz` |
| `<repo>` | GitHub repository `owner/name` | `firma/shop` |
| `<branch>` | branch the deploy comes from | `main` |
| `<locale>` | default language of the site (`DEFAULT_LOCALE`); its language is `<language>` in the SPA addresses | `cs_CZ` |
| `<locales>` | allowed languages (`LOCALES`), `<locale>` included | `cs_CZ,en_US` |
| `<dokploy>` | address of the Dokploy panel | `https://dokploy.firma.cz` |
| `<dokploy-project>` | project and environment in Dokploy the app belongs to | `shop` / `production` |
| `<app-server>` | SSH alias of the server where the app and the database will run, and its server in Dokploy | `app1` |
| `<monitoring-server>` | SSH alias of the server with Prometheus | `mon1` |
| `<prometheus.yml>` | path to the Prometheus config on that server | `/opt/prometheus/prometheus.yml` |
| `<prometheus-container>` | name of the Prometheus container | `prometheus` |
| `<prometheus-config>` | path to the config inside the container | `/etc/prometheus/prometheus.yml` |
| `<grafana>` | address of Grafana | `https://grafana.firma.cz` |
| `<loki-datasource>` | uid of the Loki datasource in Grafana | `loki` |
| `<alloy-compose>` | `composeId` of the compose with Alloy on `<app-server>`. Alloy sends the logs of the services on its map to Loki | the ID from the address of the compose in the Dokploy panel |
| `<sentry-org>`, `<sentry-team>` | slug of the organization and of the team in Sentry | `firma`, `firma` |
| `<sentry-api>` | the Sentry API, `https://de.sentry.io` for organizations in the EU | `https://sentry.io` |

The Sentry links below hold for sentry.io. A self-hosted Sentry has the same pages at its own address.

## What the person does

These steps can't go through an API, or the token would pass through the chat. Give Claude the tokens as environment variables exported before it starts, or through a password manager with a CLI.

1. **Dokploy:** an API key from Profile → API into `DOKPLOY_TOKEN`. In Settings → Git, connect a GitHub account that sees `<repo>`.
2. **Sentry, personal token** for the configuration, on `https://<sentry-org>.sentry.io/settings/account/api/auth-tokens/`, into `SENTRY_TOKEN`. Permissions:
   - Project: Admin, to create projects;
   - Release: Admin;
   - Alerts: Write;
   - Organization: Read and Integrations, for the code mappings;
   - Issue & Event: Write, to resolve the test error.

   What the token can really do shows in `GET <sentry-api>/api/0/`, in the field `auth.scopes`.
3. **Sentry, GitHub:** `https://<sentry-org>.sentry.io/settings/integrations/github/` → Add Installation for the account that owns `<repo>`.
4. **Grafana:** a service account token allowed to import dashboards (Administration → Service accounts) into `GRAFANA_TOKEN`.
5. **Cloudflare:** step 1 is done in the panel.
6. **Sentry, organization token** for the build: create it when Claude asks in step 3. Create it on `https://<sentry-org>.sentry.io/settings/auth-tokens/` → Create New Token. Its fixed permissions cover only source maps and releases. Put it straight into the app in Dokploy, Environment → Build-time Secrets, as `SENTRY_AUTH_TOKEN=<token>`. Store it nowhere else.
7. **hstspreload.org:** submit the domain by hand once no subdomain runs on HTTP only.

## Helper functions

The steps below rely on these functions (bash or zsh):

```bash
sentry_api() { curl -sS --fail-with-body -H @<(printf 'Authorization: Bearer %s' "$SENTRY_TOKEN") -H 'Content-Type: application/json' "$@"; }
dokploy_api() { curl -sS --fail-with-body -H @<(printf 'x-api-key: %s' "$DOKPLOY_TOKEN") -H 'Content-Type: application/json' "$@"; }
```

## 1. Cloudflare

In the Cloudflare panel for the zone `<domain>`:

1. DNS: A records `@` and `www` to the IP of `<app-server>`, with the proxy on.
2. SSL/TLS: mode Full.
3. Rules → Redirect Rules → Single Redirect:
   - the condition `http.host eq "www.<domain>"`;
   - the dynamic target `concat("https://<domain>", http.request.uri.path)`;
   - 301 and Preserve query string.

## 2. Sentry

Sentry must be ready before the first deploy. The image build uploads source maps to a project and a repository that must already exist, and fails otherwise.

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

3. Code mappings, so frames lead to the release's commit on GitHub. The repository must show in Sentry once GitHub is installed; otherwise stop and go back to manual item 3.
   - **Go:** the binary is built with `-trimpath`, and its frames have the path `<Go module>/app/…`. The stack root is therefore the module from `go.mod` with a slash, and the source root is empty.
   - **Vue:** after the source maps, the frames have the path `../assets/…`, so the stack root is `../`. Check it after the first error in step 7.

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

Step 3 loads the DSNs of both projects straight into the env of the app.

## 3. Dokploy: the database and the app

Everything can be done in the UI or through the API. `<dokploy>/swagger` shows the exact shapes of the bodies. The `serverId` of `<app-server>` is in the UI under Remote Servers.

1. **Project and environment:** find `<dokploy-project>` in `project.all`, or create it through `project.create`. `environment.byProjectId` gives the ID of the environment.
2. **PostgreSQL 18** on `<app-server>`: `postgres.create` with the `appName` `<projectname>-db`, then `postgres.deploy`. Dokploy adds six random characters to the `appName` (`shop-db-x7k2qp`). The whole `appName` from the answer is also the internal host of the database. The user is `postgres` with a generated password, and External Port stays empty. Set a daily backup to S3 in the UI, on the tab Backups.
3. **The app** on `<app-server>`: `application.create` with the `appName` `<projectname>-app`; take the `applicationId` from the answer. The whole `appName` (`shop-app-x7k2qp`) is the name of the Swarm service, by which Alloy recognizes the app's logs in step 5. Then:
   - `saveGithubProvider`: `<repo>`, branch `<branch>`, Build Path `/`. `github.githubProviders` gives the `githubId`.
   - `saveBuildType`: Dockerfile `docker/production/Dockerfile`, Docker Context Path `.`. An empty context isn't enough: Dokploy would take the folder of the Dockerfile, and the build wouldn't find `go.mod`.
   - `domain.create`: `<domain>`, port 8020, HTTPS with Let's Encrypt.
   - Leave Swarm Settings empty. Dokploy deploys with `start-first` and `rollback` and takes the healthcheck on `/readyz` from the image.

   Don't deploy yet.
4. **The database role and the env of the app** in one command. The env has the same structure as `.env.example`, comments included, so the two compare line by line.
   - The values of the example are never taken over, because they are for development (`GIN_MODE=debug`, `METRICS_USER=admin`…).
   - A variable the command doesn't set stays empty, and empty means the default: `release`, JSON logs, level `info`, environment `production`, release = the commit SHA, and mails off (`SMTP_ENABLED`).
   - The variables only for compose and tests (`POSTGRES_*`, `APP_DOMAIN`, `DB_TEST_*`) and the Sentry build variables stay empty in the env of the app. The organization token never belongs in the env.

   Run the block only once. A second run would write a new password into the env, while the role in the database would keep the old one. Item 6 changes the password.

   ```bash
   app=$(dokploy_api "<dokploy>/api/application.one?applicationId=<applicationId>")
   PW=$(openssl rand -hex 24)
   { printf "\\set password '%s'\n" "$PW"; cat docker/postgres/app-role.sql; } \
     | ssh <app-server> 'docker exec -i $(docker ps -q -f name=<appName of the database> | head -1) psql -v ON_ERROR_STOP=1 -U postgres -d postgres -v role=<projectname> -v db=<projectname> -f -' || exit 1
   dsn() { sentry_api "<sentry-api>/api/0/projects/<sentry-org>/$1/keys/" | jq -r '[.[] | select(.isActive)][0].dsn.public'; }
   values=$(printf '%s\n' \
     "DB_HOST=<appName of the database>" \
     "DB_NAME=<projectname>" \
     "DB_USERNAME=<projectname>" \
     "DB_PASSWORD=$PW" \
     "DB_PARAMS=?sslmode=disable&pool_max_conns=10" \
     "DEFAULT_LOCALE=<locale>" \
     "LOCALES=<locales>" \
     "TRUSTED_PROXIES=127.0.0.1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16" \
     "METRICS_USER=$(openssl rand -hex 16)" \
     "METRICS_PASSWORD=$(openssl rand -hex 32)" \
     "SENTRY_DSN=$(dsn <projectname>-go)" \
     "SENTRY_FRONTEND_DSN=$(dsn <projectname>-vue)")
   env=$(awk 'NR == FNR { i = index($0, "="); v[substr($0, 1, i - 1)] = substr($0, i + 1); next }
     /^[A-Za-z_][A-Za-z0-9_]*=/ { k = substr($0, 1, index($0, "=") - 1); print k "=" v[k]; next } { print }' <(printf '%s\n' "$values") .env.example)
   jq -n --rawfile env <(printf '%s' "$env") --slurpfile a <(printf '%s' "$app") \
     '{applicationId: $a[0].applicationId, env: $env, buildArgs: "SENTRY_ORG=<sentry-org>\nSENTRY_PROJECT=<projectname>-vue\nSENTRY_REPOSITORY=<repo>", buildSecrets: ($a[0].buildSecrets // ""), createEnvFile: $a[0].createEnvFile}' \
     | dokploy_api -X POST --data-binary @- "<dokploy>/api/application.saveEnvironment"
   dokploy_api "<dokploy>/api/application.one?applicationId=<applicationId>" \
     | jq '{filled: ([.env | split("\n")[] | select(test("^(DB_HOST|DB_NAME|DB_USERNAME|DB_PASSWORD|DB_PARAMS|DEFAULT_LOCALE|METRICS_USER|METRICS_PASSWORD|SENTRY_DSN|SENTRY_FRONTEND_DSN)=.+"))] | length), buildArgs}'
   ```

   The check must show `filled: 10` and three build args. `application.saveEnvironment` always wants all five fields (`applicationId`, `env`, `buildArgs`, `buildSecrets`, `createEnvFile`). For a later change, load the current values and send them back, or they are deleted. After every change, compare the fingerprint (`shasum`) of the lines that shouldn't have changed.
5. **The build secret and the deploy:** now the person puts in the organization token (manual item 6). Without it, the build uploads nothing to Sentry. With it, any error of the upload stops the build, and Dokploy keeps the previous version running. Then run `application.deploy` and wait until `https://<domain>/readyz` answers 200. The log of the deploy must show `migration applied` and `listening`.
6. **Optional: change the database password.** The password of the role is readable only in `DB_PASSWORD` in Dokploy; Postgres keeps its hash. `app-role.sql` only creates the role, so a change is an `ALTER ROLE`. In the same command, replace `DB_PASSWORD` the way item 4 does, then run Redeploy. Open connections keep working, and new ones can't connect with the old password.

   ```bash
   PW=$(openssl rand -hex 24)
   printf "ALTER ROLE \"<projectname>\" PASSWORD '%s';\n" "$PW" \
     | ssh <app-server> 'docker exec -i $(docker ps -q -f name=<appName of the database> | head -1) psql -v ON_ERROR_STOP=1 -U postgres -d postgres -f -'
   ```

## 4. Prometheus

The job takes the account for `/metrics` straight from the env of the app and reaches the server through the stdin of SSH, so it's never printed. The block assumes that `scrape_configs` comes last in `<prometheus.yml>` and that its jobs are indented by two spaces. Otherwise, add the job by hand. When the check of the config fails, the block restores the backup.

```bash
envs=$(dokploy_api "<dokploy>/api/application.one?applicationId=<applicationId>" | jq -r .env)
user=$(sed -n 's/^METRICS_USER=//p' <<<"$envs"); pass=$(sed -n 's/^METRICS_PASSWORD=//p' <<<"$envs")
ssh <monitoring-server> 'cp -p <prometheus.yml> <prometheus.yml>.bak-first-deploy'
printf "  - job_name: '<projectname>'\n    scheme: https\n    metrics_path: /metrics\n    basic_auth:\n      username: '%s'\n      password: '%s'\n    static_configs:\n      - targets: ['<domain>']\n" "$user" "$pass" \
  | ssh <monitoring-server> 'cat >> <prometheus.yml>'
ssh <monitoring-server> 'docker exec <prometheus-container> promtool check config <prometheus-config> && docker kill -s HUP <prometheus-container> || { cp -p <prometheus.yml>.bak-first-deploy <prometheus.yml>; exit 1; }'
```

The job is appended with `>>`, not `sed -i`. When the file is mounted into the container as a single file, the container wouldn't see the new file `sed -i` writes.

Within a few tens of seconds, the job must be `up`:

```bash
ssh <monitoring-server> 'docker exec <prometheus-container> wget -qO- http://localhost:9090/api/v1/targets' \
  | jq -r '.data.activeTargets[] | select(.labels.job == "<projectname>") | .health'
```

## 5. Logs

Alloy on `<app-server>` reads the logs of the containers and sends only the services on its map to Loki.
- The map is the file `config.alloy` in the compose `<alloy-compose>`, in Dokploy under Advanced → Mounts.
- Each rule of the map gives the name of a Swarm service the label `project/service`. Alloy turns it into the labels `project`, `service` and `service_name`, which the dashboard reads.
- The app doesn't change; it keeps logging JSON to stdout.

The block adds a rule for the app to the map and deploys the compose again. The compose `<alloy-compose>` has its own command in Dokploy with `--force-recreate`. Without it, Dokploy doesn't restart the container, and Alloy doesn't load the new map.

```bash
app=$(dokploy_api "<dokploy>/api/application.one?applicationId=<applicationId>" | jq -r .appName)
mount=$(dokploy_api "<dokploy>/api/compose.one?composeId=<alloy-compose>" | jq '.mounts[] | select(.filePath == "config.alloy") | {mountId, content}')
config=$(jq -r .content <<<"$mount" | awk -v app="$app" '{ print } /targets = discovery.docker.containers.targets/ {
  printf "\n  rule {\n    source_labels = [\"__meta_docker_container_label_com_docker_swarm_service_name\"]\n    regex         = \"%s\"\n    target_label  = \"__service\"\n    replacement   = \"<projectname>/app\"\n  }\n", app }')
[ "$(grep -c "regex         = \"$app\"" <<<"$config")" = 1 ] || exit 1
jq -n --rawfile c <(printf '%s' "$config") --slurpfile m <(printf '%s' "$mount") \
  '{mountId: $m[0].mountId, type: "file", content: $c, filePath: "config.alloy", mountPath: "/etc/alloy/config.alloy", serviceType: "compose", composeId: "<alloy-compose>"}' \
  | dokploy_api -X POST --data-binary @- "<dokploy>/api/mounts.update" | jq '{mountId}'
dokploy_api -X POST --data '{"composeId":"<alloy-compose>"}' "<dokploy>/api/compose.deploy" | jq '{success}'
```

The block stops before writing unless the map holds the rule for the app exactly once.
- After a second run, the rule would be there twice.
- When it isn't there at all, the map has no line `targets = discovery.docker.containers.targets`. Then add the rule by hand in Dokploy.

## 6. Grafana

[grafana-dashboard.json](grafana-dashboard.json) has placeholders instead of the project name:
- `<projectname>` in the `uid`, in the tag in `tags`, and in the defaults of the variables `job` and `project`;
- `<domain>` in the `title`.

The row Logs takes its logs from Loki by the label `project`. The file in the repository stays as it is, and the import fills in the values. The `sed` patterns are written as `[<]…[>]`, so nobody mistakes them for placeholders of this guide.

```bash
dashboard=$(sed -e 's/[<]projectname[>]/<projectname>/g' -e 's/[<]domain[>]/<domain>/g' docs/grafana-dashboard.json)
grep -c '[<>]' <<<"$dashboard"
jq '{dashboard: (. + {id: null}), overwrite: true, folderUid: ""}' <<<"$dashboard" \
  | curl -sS --fail-with-body -X POST -H @<(printf 'Authorization: Bearer %s' "$GRAFANA_TOKEN") -H 'Content-Type: application/json' \
      --data-binary @- <grafana>/api/dashboards/db | jq '{status, url}'
```

`grep -c` must print 0.

## 7. Verification

`<SHA>` is the deployed commit, in the `data-release` of the tag `<meta name="sentry">` on `https://<domain>/<language>/app`. Claude checks these itself:

- `https://<domain>/` answers 200 with `cf-cache-status: DYNAMIC`, `/readyz` 200, `/metrics` without the account 401, and `https://www.<domain>/` 301 to the apex.
- `<meta name="sentry">` on `https://<domain>/<language>/app` has `data-environment="production"` and a `data-release` equal to the SHA of the last commit of `<branch>`.
- The release in Sentry has its commits and a deploy to `production`:

  ```bash
  sentry_api "<sentry-api>/api/0/organizations/<sentry-org>/releases/<SHA>/" | jq '{version, commitCount, deployCount, projects: [.projects[].slug]}'
  ```

- Source maps: Sentry finds the `sentry-dbid-…` of the live `/build/assets/app-….js`, and the map never leaks out:

  ```bash
  js=$(curl -sS https://<domain>/<language>/app | grep -oE '/build/assets/app-[A-Za-z0-9_-]+\.js' | head -1)
  id=$(curl -sS "https://<domain>$js" | grep -oE 'sentry-dbid-[0-9a-f-]{36}' | head -1 | cut -c13-)
  sentry_api "<sentry-api>/api/0/projects/<sentry-org>/<projectname>-vue/artifact-lookup/?debug_id=$id" | jq 'length'
  curl -s -o /dev/null -w '%{http_code}\n' "https://<domain>$js.map"
  ```

  The first command must print at least 1, the second 404.
- Loki has the logs of the app. The healthcheck calls `/readyz` every 30 s, so the number must be above 0:

  ```bash
  curl -sS --fail-with-body -G -H @<(printf 'Authorization: Bearer %s' "$GRAFANA_TOKEN") \
    "<grafana>/api/datasources/proxy/uid/<loki-datasource>/loki/api/v1/query" \
    --data-urlencode 'query=sum(count_over_time({project="<projectname>"}[5m]))' | jq -r '.data.result[0].value[1]'
  ```

- securityheaders.com and MDN Observatory show A+.

With the person's permission, because an issue is created and an email arrives:

- **A test error of the frontend.** On `https://<domain>/<language>/app`, run the code below in the browser console. A Claude with a browser does it itself, otherwise the person does. The error arises in `assets/shared/Sentry/startSentry.ts`, so it tests the source maps on our own code.

  ```js
  const probe = new Event('securitypolicyviolation');
  Object.defineProperty(probe, 'sourceFile', { get() { throw new Error('Sentry probe'); } });
  setTimeout(() => document.dispatchEvent(probe));
  ```

  Then check the event and close the issue:
  - The event in `<projectname>-vue` must have the frame `../assets/shared/Sentry/startSentry.ts` with its source line.
  - `GET <sentry-api>/api/0/projects/<sentry-org>/<projectname>-vue/stacktrace-link/?file=../assets/shared/Sentry/startSentry.ts&platform=javascript&commitId=<SHA>` must answer a `sourceUrl` into GitHub.
  - Then resolve the issue: `PUT <sentry-api>/api/0/organizations/<sentry-org>/issues/<issue id>/` with `{"status":"resolved"}`. The `<issue id>` is the `groupID` of the event.
- **Frontend and backend errors joined.** In production this would need a database outage, so only on request. Locally it can be tried like this:
  - a binary built as in the image (`go build -trimpath`);
  - the real DSNs and `SENTRY_ENVIRONMENT=development`;
  - a temporary Postgres that you stop after the start.

  `fetch('/readyz')` from the frontend then answers 503, and the backend error has the same `trace_id` as the frontend error from the same page.
- **Go code in Sentry** loads from GitHub only in the UI; the API can't check it. Open a backend issue and check that its frames show lines of code.
