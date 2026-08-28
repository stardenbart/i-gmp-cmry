# syntax=docker/dockerfile:1
FROM node:20.19-alpine3.22 AS builder

WORKDIR /app

# Build argument – dipass dari docker-compose via build.args
ARG INTERNAL_BACKEND_URL=http://backend:8080
ARG MINIO_ENDPOINT=http://minio:9000
# NEXT_PUBLIC_* vars are inlined into the client bundle at build time by
# Next.js — setting them only in docker-compose's runtime `environment:`
# block (like the vars above) does nothing for this one; it MUST be a
# build arg, or the push-subscribe button will silently get `undefined`
# in the browser.
ARG NEXT_PUBLIC_VAPID_PUBLIC_KEY=
ENV INTERNAL_BACKEND_URL=$INTERNAL_BACKEND_URL
ENV MINIO_ENDPOINT=$MINIO_ENDPOINT
ENV NEXT_PUBLIC_VAPID_PUBLIC_KEY=$NEXT_PUBLIC_VAPID_PUBLIC_KEY
ENV NEXT_TELEMETRY_DISABLED=1

# Install dependencies
COPY package.json package-lock.json* ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --no-audit --no-fund

# Copy source code dan build
COPY . .
RUN npm run build
RUN rm -f .next/standalone/.env .next/standalone/.env.*

# ─────────────────────────────────────────────
# Runner stage
FROM node:20.19-alpine3.22 AS runner

WORKDIR /app

ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    HOSTNAME=0.0.0.0 \
    PORT=3000 \
    INTERNAL_BACKEND_URL=http://backend:8080 \
    MINIO_ENDPOINT=http://minio:9000

COPY --from=builder --chown=node:node /app/.next/standalone ./
COPY --from=builder --chown=node:node /app/.next/static ./.next/static
COPY --from=builder --chown=node:node /app/public ./public

USER node

EXPOSE 3000

CMD ["node", "server.js"]
