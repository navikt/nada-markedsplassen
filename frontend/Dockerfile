FROM node:23.9.0-alpine AS node-with-deps

WORKDIR /usr/app

COPY package*.json .

RUN --mount=type=secret,id=NODE_AUTH_TOKEN sh -c \
    'npm config set //npm.pkg.github.com/:_authToken=$(cat /run/secrets/NODE_AUTH_TOKEN)'
RUN npm config set @navikt:registry=https://npm.pkg.github.com

COPY package.json package-lock.json ./

RUN npm install --quiet

FROM node-with-deps

COPY . .

ENV NODE_ENV=production
ENV NEXT_PUBLIC_ENV=production
ENV NEXT_PUBLIC_BACKEND=http://nada-backend/api
ENV NPM_CONFIG_CACHE=/tmp

RUN npm run build

CMD npm run start
