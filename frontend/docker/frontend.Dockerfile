# Gunakan image node alpine untuk build
FROM node:20-alpine AS builder

WORKDIR /app

# Build argument – dipass dari docker-compose via build.args
ARG INTERNAL_BACKEND_URL=http://backend:8080
ENV INTERNAL_BACKEND_URL=$INTERNAL_BACKEND_URL

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
ENV INTERNAL_BACKEND_URL=http://backend:8080

# Copy dari builder
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public

EXPOSE 3000

CMD ["npm", "start"]
