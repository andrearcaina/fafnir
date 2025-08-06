ARG NODE_VERSION=18

# Development stage
FROM node:${NODE_VERSION}-alpine AS development

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json* ./
RUN npm i

EXPOSE 5000

CMD ["npm", "run", "dev"]

# TODO: add build stage and prod stage
