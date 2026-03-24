## Prerequisites:

Install [Node 24](https://nodejs.org/en) (LTS) or later.

Install pnpm
````
RUN npm install -g pnpm
````

Set your GitHub token (with `read:packages` scope) in your shell profile (e.g. `~/.zshrc`):
```
export NPM_AUTH_TOKEN=<your-github-token>
```

Then install dependencies:
```
pnpm install
```

Optionally add binaries for local node_modules in path:
```
export PATH=./node_modules/.bin:$PATH
```

## Development

Run the development server:

```bash
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

## Building the Docker image locally

The Docker image expects the app to be built outside of Docker. Run these steps from the `frontend/` directory first:

```bash
pnpm install --frozen-lockfile
NODE_ENV=production NEXT_PUBLIC_ENV=production NEXT_PUBLIC_BACKEND=http://nada-backend/api pnpm build
pnpm prune --prod
```

Then build and run the image from the repo root:

```bash
docker build -t nada-markedsplassen-frontend frontend/
docker run -p 3001:3000 nada-markedsplassen-frontend
```

Open [http://localhost:3001](http://localhost:3001). Port 3001 is used to avoid conflicts with local dev servers on 3000.

### Storybooks

For easier development of frontend components you can add stories to the `stories` folder and then run:
```bash
$ pnpm storybook
```

## Updating schema:

Fetch the latest version of the schema from the [backend](https://github.com/navikt/nada-backend/blob/main/spec-v1.0.yaml)
Then run

```
npx openapi-typescript ../nada-backend/spec-v1.0.yaml --output lib/schema/schema.ts
pnpm format-schema
```
