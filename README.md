<h1>
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/img/logo-dark.svg">
    <img src="assets/img/logo.svg" alt="Gin GoKick" width="294" height="64">
  </picture>
</h1>

Lightweight full-stack Golang boilerplate. Build apps in Go the light, clean and fast way: Gin, Vue 3 and strict rules under which people and AI write equally good code.

Clone the repo and let AI do the work. [AGENTS.md](AGENTS.md) holds the rules that people and AI follow alike.

📖 **[Read the documentation](docs/README.md)** to see what you get on day one: sign-in with sessions, languages, emails, data grids, a strict CSP, Sentry, metrics and cookie consent, already wired together and guarded by the tools.

**Live demo**

- 🇺🇸 [gin-gokick.strategio.dev/en](https://gin-gokick.strategio.dev/en)
- 🇨🇿 [gin-gokick.strategio.dev](https://gin-gokick.strategio.dev)

**Quick start**

You need Go, Node 24 with Yarn, and Docker with OrbStack for the database.

1. **Clone the repository**

   ```bash
   git clone https://github.com/jzaplet/gin-gokick.git && cd gin-gokick
   ```

2. **Create the `.env`**

   ```bash
   cp .env.example .env
   ```

3. **Start PostgreSQL**

   ```bash
   docker compose up -d db
   ```

4. **Install the packages** (`corepack enable` once per machine gives Yarn 4)

   ```bash
   corepack enable && yarn install
   ```

5. **Start Vite** and keep it running

   ```bash
   yarn dev
   ```

6. **Start the app** in a second terminal and open [localhost:8020](http://localhost:8020)

   ```bash
   go run ./app
   ```

[Quick start](docs/README.md#quick-start) in the documentation says how to run it without OrbStack and how to install the git hooks.
