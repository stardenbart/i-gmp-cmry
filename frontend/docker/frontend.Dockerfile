# Gunakan image node alpine untuk build
FROM node:20-alpine AS builder

WORKDIR /app

# Build argument – dipass dari docker-compose via build.args
ARG INTERNAL_BACKEND_URL=https://mills-bare-harris-deemed.trycloudflare.com
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

# Install dependencies
COPY package.json package-lock.json* ./
RUN npm install

# Copy source code dan build
COPY . .
RUN rm -rf .next
RUN npm run build

# ─────────────────────────────────────────────
# Runner stage
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV production
ENV INTERNAL_BACKEND_URL=https://mills-bare-harris-deemed.trycloudflare.com
ENV MINIO_ENDPOINT=http://minio:9000

# Copy dari builder
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public

EXPOSE 3000

CMD ["npm", "start"]
