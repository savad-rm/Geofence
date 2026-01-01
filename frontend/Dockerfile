FROM node:18-alpine AS builder

WORKDIR /app

COPY package.json ./

RUN npm install

COPY . .

ENV REACT_APP_API_URL=http://localhost:8080

RUN npm run build

FROM node:18-alpine

RUN npm install -g serve

WORKDIR /app

COPY --from=builder /app/build .

EXPOSE 3000

CMD ["serve", "-s", ".", "-l", "3000"]
