## Prerequisites:

Install [Node18](https://nodejs.org/en) or later

Install dependencies:
```
npm install
```

Optionally add binaries for local node_modules in path:
```
export PATH=./node_modules/.bin:$PATH
```

## Development:
First, run the development server:

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

## Updating schema:

Fetch the latest version of the schema from the [backend](https://github.com/navikt/nada-backend/blob/main/spec-v1.0.yaml)
Then run

```
npx openapi-typescript ../nada-backend/spec-v1.0.yaml --output lib/schema/schema.ts
npm run format-schema
```
