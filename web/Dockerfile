FROM node:18-alpine

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm i

COPY . .

ENV NODE_ENV=development
ENV PORT=3000

EXPOSE 3000

CMD ["npm", "run", "dev"]
