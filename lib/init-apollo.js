"use strict";

const fetch = require("isomorphic-unfetch");
const { ApolloClient } = require("apollo-client");
const { InMemoryCache } = require("apollo-cache-inmemory");
const { createHttpLink } = require("apollo-link-http");
const { setContext } = require("apollo-link-context");

const { getToken } = require("./auth.js");

let apolloClient = null;
const GRAPHQL_ORIGIN = process.env.GRAPHQL_ORIGIN;

// Polyfill fetch() on the server (used by apollo-client)
if (!process.browser) {
  global.fetch = fetch;
}

let exp = {};
exp["create"] = function(initialState, options) {
  const link = createHttpLink({ uri: GRAPHQL_ORIGIN });

  const authLink = setContext((_, { headers }) => {
    // get the authentication token from local storage if it exists
    const token = getToken();

    // return the headers to the context so httpLink can read them
    return {
      headers: {
        ...headers,
        authorization: token ? `Bearer ${token}` : "",
      },
    };
  });

  return new ApolloClient({
    connectToDevTools: process.browser,
    ssrMode: !process.browser, // Disables forceFetch on the server (so queries are only run once)
    link: authLink.concat(link),
    cache: new InMemoryCache().restore(initialState || {}),
  });
};

exp["initApollo"] = function(initialState, options) {
  // Make sure to create a new client for every server-side request so that data
  // isn't shared between connections (which would be bad)
  if (!process.browser) {
    return exp.create(initialState, options);
  }

  // Reuse client on the client-side
  if (!apolloClient) {
    apolloClient = exp.create(initialState, options);
  }

  return apolloClient;
};

module.exports = exp;
