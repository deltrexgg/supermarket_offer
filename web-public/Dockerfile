FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci || npm install

FROM node:22-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ARG NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
# The public address of this site: used in canonical links, robots.txt and the sitemap
ARG SITE_URL=http://localhost:3001
ENV SITE_URL=$SITE_URL
RUN npm run build

FROM node:22-alpine AS run
WORKDIR /app
ARG SITE_URL=http://localhost:3001
ENV NODE_ENV=production PORT=3000 HOSTNAME=0.0.0.0 SITE_URL=$SITE_URL
RUN addgroup -S app && adduser -S app -G app
COPY --from=build --chown=app:app /app/.next/standalone ./
COPY --from=build --chown=app:app /app/.next/static ./.next/static
COPY --from=build --chown=app:app /app/public ./public
USER app
EXPOSE 3000
CMD ["node", "server.js"]
