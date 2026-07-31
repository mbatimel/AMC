FROM node:22-alpine AS deps

WORKDIR /src

COPY admin-front/package.json admin-front/yarn.lock ./
RUN yarn install --frozen-lockfile

FROM node:22-alpine AS builder

WORKDIR /src

COPY --from=deps /src/node_modules ./node_modules
COPY admin-front/ ./

RUN yarn build

FROM node:22-alpine AS runner

ENV NODE_ENV=production
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /src/public ./public
COPY --from=builder --chown=app:app /src/.next/standalone ./
COPY --from=builder --chown=app:app /src/.next/static ./.next/static

# Каталог для PORTAL_DATA_FILE. Named volume наследует владельца этой
# директории при первом монтировании — иначе процесс под юзером app
# не смог бы писать состояние портальных модулей.
RUN mkdir -p /data && chown app:app /data

USER app
EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME=0.0.0.0

ENTRYPOINT ["node", "server.js"]
