FROM node:12-alpine
WORKDIR /opt
EXPOSE 8080

ENV NODE_ENV=production
ENV GRAPHQL_ORIGIN="https://graphql.natwelch.com/graphql"
ENV PORT=8080

RUN apk add --no-cache git

COPY . .

RUN yarn install --non-interactive --frozen-lockfile
RUN yarn run build
CMD ["yarn", "run", "start"]
