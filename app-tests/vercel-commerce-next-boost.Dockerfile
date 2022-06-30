FROM node:16-bullseye-slim AS builder

RUN apt update && apt install -y git \
  && git clone https://github.com/vercel/commerce.git \
  && cd commerce \
  && npm install --save-prod \
  && npm install @next-boost/next-boost @next-boost/hybrid-disk-cache \
  && NODE_ENV="production" npm run build \
  && sed -i 's/next start/next-boost/g' site/package.json



FROM node:16-bullseye-slim

WORKDIR /app

ENV NODE_ENV="production"
ENV NEXT_TELEMETRY_DISABLED="1"

COPY --from=builder /commerce/site/public ./public
COPY --from=builder /commerce/site/.next ./.next
COPY --from=builder /commerce/node_modules ./node_modules
COPY --from=builder /commerce/site/package.json ./package.json

RUN chown -R node:node /app

USER node

EXPOSE 3000


ENTRYPOINT [ "npm", "start" ]