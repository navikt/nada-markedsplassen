# Dependencies
FROM node:25.6.0-alpine AS dependencies

WORKDIR /usr/app

RUN npm install -g pnpm

COPY package.json pnpm-lock.yaml ./

RUN --mount=type=secret,id=NODE_AUTH_TOKEN sh -c \
    'pnpm config set //npm.pkg.github.com/:_authToken=$(cat /run/secrets/NODE_AUTH_TOKEN)'

RUN pnpm config set @navikt:registry=https://npm.pkg.github.com

RUN pnpm install --frozen-lockfile


# Builder
FROM node:25.6.0-alpine AS builder

WORKDIR /usr/app
RUN npm install -g pnpm

COPY --from=dependencies /usr/app/node_modules ./node_modules
COPY . .

ENV NODE_ENV=production
ENV NEXT_PUBLIC_ENV=production
ENV NEXT_PUBLIC_BACKEND=http://nada-backend/api

RUN pnpm build


# Runtime
FROM node:25.6.0-alpine AS runtime

WORKDIR /usr/app
RUN npm install -g pnpm

ENV NODE_ENV=production
ENV NEXT_PUBLIC_ENV=production
ENV NEXT_PUBLIC_BACKEND=http://nada-backend/api

COPY package.json pnpm-lock.yaml ./

RUN --mount=type=secret,id=NODE_AUTH_TOKEN sh -c \
    'pnpm config set //npm.pkg.github.com/:_authToken=$(cat /run/secrets/NODE_AUTH_TOKEN)'

RUN pnpm config set @navikt:registry=https://npm.pkg.github.com

RUN pnpm install --prod --frozen-lockfile

COPY --from=builder /usr/app/.next ./.next
COPY --from=builder /usr/app/public ./public

CMD ["pnpm", "start"]