// https://github.com/zeit/next-plugins/tree/master/packages/next-css
const withCSS = require("@zeit/next-css");
const port = process.env.PORT || 8080;
module.exports = withCSS({
  env: {
    GRAPHQL_ORIGIN: process.env.GRAPHQL_ORIGIN,
    AUTH0_CLIENT_ID: "MwFD0COlI4F4AWvOZThe1psOIletecnL",
    AUTH0_DOMAIN: "icco.auth0.com",
    DOMAIN: process.env.DOMAIN || `http://localhost:${port}`,
    PORT: port,
  },
});
