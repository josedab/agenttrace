# Dependencies stage
FROM node:20.19.4-alpine3.22@sha256:df02558528d3d3d0d621f112e232611aecfee7cbc654f6b375765f72bb262799 AS deps
WORKDIR /app

# Install dependencies
COPY package.json package-lock.json* ./
RUN npm ci

# Builder stage
FROM node:20.19.4-alpine3.22@sha256:df02558528d3d3d0d621f112e232611aecfee7cbc654f6b375765f72bb262799 AS builder
WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Set environment variables for build
ARG NEXT_PUBLIC_API_URL=""
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production

# Build the application
RUN npm run build
RUN mkdir -p /app/.next/cache

# Runner stage
FROM cgr.dev/chainguard/node@sha256:c002402b355201714dfbeb3b1a7ca99152cccced4b334496db2442a3a9bf4883 AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Copy built assets
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nonroot:nonroot /app/.next/standalone ./
COPY --from=builder --chown=nonroot:nonroot /app/.next/static ./.next/static
COPY --from=builder --chown=nonroot:nonroot /app/.next/cache ./.next/cache

USER 0
RUN chown -R 65532:65532 /app/.next
USER 65532

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["server.js"]
