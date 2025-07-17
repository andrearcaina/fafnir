FROM node:18-alpine

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm i

CMD ["npm", "run", "dev"]