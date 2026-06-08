# syntax=docker/dockerfile:1

# ---- Build stage (Angular -> static fajlovi) ----
FROM node:20-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# ---- Runtime stage (nginx servira SPA) ----
FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
# Angular 17 "application" builder izlaz je u dist/<projekat>/browser
COPY --from=build /app/dist/shophub-web/browser /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK CMD wget -qO- http://localhost/ >/dev/null 2>&1 || exit 1
