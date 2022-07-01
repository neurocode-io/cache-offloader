FROM node:16-bullseye-slim AS builder

ARG cache_offloader_version=0.1.5

RUN apt update && apt install -y git wget \
  && git clone https://github.com/vercel/commerce.git \
  && cd commerce \
  && npm install --save-prod \
  && NODE_ENV="production" npm run build \
  && wget https://github.com/neurocode-io/cache-offloader/releases/download/v${cache_offloader_version}/cache-offloader_${cache_offloader_version}_linux_amd64.tar.gz \
  && tar -xvf cache-offloader_${cache_offloader_version}_linux_amd64.tar.gz


FROM node:16-bullseye-slim

WORKDIR /app

ENV NODE_ENV="production"
ENV NEXT_TELEMETRY_DISABLED="1"

COPY --from=builder /commerce/site/public ./public
COPY --from=builder /commerce/site/.next ./.next
COPY --from=builder /commerce/node_modules ./node_modules
COPY --from=builder /commerce/site/package.json ./package.json
COPY --from=builder /commerce/cache-offloader ./cache-offloader
COPY entrypoint.sh .

ENV DOWNSTREAM_HOST=http://localhost:3000
ENV CACHE_SHOULD_HASH_QUERY="true"
ENV CACHE_STRATEGY=LFU
ENV CACHE_SIZE_MB=10
ENV CACHE_STALE_WHILE_REVALIDATE_SEC=60
ENV SERVER_PORT=8000
ENV SERVER_STORAGE=memory
ENV CACHE_IGNORE_ENDPOINTS="/probes/liveness"
ENV CACHE_IGNORE_ENDPOINTS="/probes/readiness"

RUN chown -R node:node /app

USER node
EXPOSE 8000

ENTRYPOINT [ "./entrypoint.sh" ]
