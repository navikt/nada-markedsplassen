# Dependencies
FROM node:25.8.1-alpine AS dependencies

WORKDIR /usr/app

RUN npm install -g pnpm

COPY package.json pnpm-lock.yaml ./

RUN --mount=type=secret,id=NODE_AUTH_TOKEN \
    printf "//npm.pkg.github.com/:_authToken=%s\n" \
    "$(cat /run/secrets/NODE_AUTH_TOKEN)" > .npmrc && \
    pnpm install --frozen-lockfile && \
    rm -f .npmrc

# Builder
FROM node:25.8.1-alpine AS builder

WORKDIR /usr/app
RUN npm install -g pnpm

COPY --from=dependencies /usr/app/node_modules ./node_modules
COPY . .

ENV NODE_ENV=production
ENV NEXT_PUBLIC_ENV=production
ENV NEXT_PUBLIC_BACKEND=http://nada-backend/api

RUN pnpm build


# Runtime
FROM node:25.8.1-alpine AS runtime

WORKDIR /usr/app
RUN npm install -g pnpm

ENV NODE_ENV=production
ENV NEXT_PUBLIC_ENV=production
ENV NEXT_PUBLIC_BACKEND=http://nada-backend/api

COPY package.json pnpm-lock.yaml ./

RUN --mount=type=secret,id=NODE_AUTH_TOKEN \
    printf "//npm.pkg.github.com/:_authToken=%s\n" \
    "$(cat /run/secrets/NODE_AUTH_TOKEN)" > .npmrc && \
    pnpm install --prod --frozen-lockfile && \
    rm -f .npmrc

COPY --from=builder /usr/app/.next ./.next
COPY --from=builder /usr/app/public ./public

CMD ["pnpm", "start"]