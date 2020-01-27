import fetch from "isomorphic-unfetch";
import { ApolloClient } from "apollo-client";
import { ApolloLink } from "apollo-link";
import { HttpLink } from "apollo-link-http";
import { InMemoryCache } from "apollo-cache-inmemory";
import { RetryLink } from "apollo-link-retry";
import { createPersistedQueryLink } from "apollo-link-persisted-queries";
import { onError } from "apollo-link-error";
import { setContext } from "apollo-link-context";

import { logger } from "./logger";
import { getToken } from "./auth";

export const GRAPHQL_ORIGIN =
  process.env.GRAPHQL_ORIGIN || "https://graphql.natwelch.com/graphql";

/**
 * Creates and configures the ApolloClient
 * @param  {Object} [initialState={}]
 */
export function createApolloClient(initialState = {}) {
  const httpLink = new HttpLink({ uri: GRAPHQL_ORIGIN });

  const retryLink = new RetryLink();

  const aptLink = createPersistedQueryLink({
    useGETForHashedQueries: false,
  });

  const errorLink = onError(
    ({ operation, response, graphQLErrors, networkError }) => {
      let data = {
        operation,
        response,
      };

      if (graphQLErrors) {
        graphQLErrors.forEach(err => {
          data["err"] = err;
          logger.error(data, `[GraphQL error]: ${err.message}`);
        });
      }

      if (networkError) {
        data["err"] = networkError;
        logger.error(data, `[Network error]: ${networkError}`);
      }
    }
  );

  const authLink = setContext((_, { headers }) => {
    // get the authentication token from local storage if it exists
    const token = getToken();

    // return the headers to the context so httpLink can read them
    return {
      headers: {
        ...headers,
        authorization: token && token.length > 5 ? `Bearer ${token}` : "",
      },
    };
  });

  const link = ApolloLink.from([
    errorLink,
    retryLink,
    authLink,
    aptLink,
    httpLink,
  ]);

  // Check out https://github.com/zeit/next.js/pull/4611 if you want to use the AWSAppSyncClient
  return new ApolloClient({
    ssrMode: typeof window === "undefined", // Disables forceFetch on the server (so queries are only run once)
    link,
    cache: new InMemoryCache().restore(initialState),
  });
}
