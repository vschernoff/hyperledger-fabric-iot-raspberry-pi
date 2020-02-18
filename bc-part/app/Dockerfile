FROM node:12.0.0-alpine

RUN mkdir /app
WORKDIR /app

COPY package.json yarn.lock /app/
RUN yarn

COPY . /app/

RUN yarn build:client

EXPOSE 3000

CMD [ "yarn", "start" ]
