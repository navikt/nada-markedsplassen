FROM node:25.6.0-alpine AS dependencies

WORKDIR /usr/app

COPY package.json package-lock.json ./

RUN --mount=type=secret,id=NODE_AUTH_TOKEN sh -c \
    'npm config set //npm.pkg.github.com/:_authToken=$(cat /run/secrets/NODE_AUTH_TOKEN)'
RUN npm config set @navikt:registry=https://npm.pkg.github.com

RUN npm ci --quiet

FROM node:25.6.0-alpine AS builder

WORKDIR /usr/app

COPY --from=dependencies /usr/app/node_modules ./node_modules
COPY . .

ENV NODE_ENV=production
ENV NEXT_PUBLIC_ENV=production
ENV NEXT_PUBLIC_BACKEND=http://nada-backend/api

RUN npm run build

FROM node:25.6.0-alpine AS runtime

WORKDIR /usr/app

ENV NODE_ENV=production
ENV NEXT_PUBLIC_ENV=production
ENV NEXT_PUBLIC_BACKEND=http://nada-backend/api
ENV NPM_CONFIG_CACHE=/tmp

COPY package.json package-lock.json ./

RUN --mount=type=secret,id=NODE_AUTH_TOKEN sh -c \
    'npm config set //npm.pkg.github.com/:_authToken=$(cat /run/secrets/NODE_AUTH_TOKEN)'
RUN npm config set @navikt:registry=https://npm.pkg.github.com

RUN npm ci --only=production --quiet

COPY --from=builder /usr/app/.next ./.next
COPY --from=builder /usr/app/public ./public

CMD ["npm", "run", "start"]