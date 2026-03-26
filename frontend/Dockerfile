FROM europe-north1-docker.pkg.dev/cgr-nav/pull-through/nav.no/node:24-slim

WORKDIR /usr/app

ENV NODE_ENV=production

COPY --chown=nonroot:nonroot node_modules ./node_modules
COPY --chown=nonroot:nonroot .next ./.next
COPY --chown=nonroot:nonroot public ./public

CMD ["node_modules/next/dist/bin/next", "start"]
