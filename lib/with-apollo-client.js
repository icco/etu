import React from "react";
import Head from "next/head";
import { getDataFromTree } from "react-apollo";

import { initApollo } from "./init-apollo";
import { logger } from "./logger";

export default App => {
  return class Apollo extends React.Component {
    static displayName = `withApollo(${
      App.displayName ? App.displayName : "App"
    })`;

    static async getInitialProps(ctx) {
      const {
        Component,
        router,
        ctx: { res },
      } = ctx;

      // TODO: Possibly change this to pass the token, instead of in the apollo
      // link.
      const apollo = initApollo({}, {});

      ctx.ctx.apolloClient = apollo;

      let appProps = {};
      if (App.getInitialProps) {
        appProps = await App.getInitialProps(ctx);
      }

      if (res && res.finished) {
        // When redirecting, the response is finished.
        // No point in continuing to render
        return {};
      }

      if (!process.browser) {
        try {
          // Run all GraphQL queries
          await getDataFromTree(
            <App
              {...appProps}
              Component={Component}
              router={router}
              apolloClient={apollo}
            />
          );
        } catch (error) {
          // Prevent Apollo Client GraphQL errors from crashing SSR.
          // Handle them in components via the data.error prop:
          // https://www.apollographql.com/docs/react/api/react-apollo.html#graphql-query-data-error
          logger.error(error, "Uncaught error while running `getDataFromTree`");
        }

        // getDataFromTree does not call componentWillUnmount
        // head side effect therefore need to be cleared manually
        Head.rewind();
      }

      // Extract query data from the Apollo's store
      const apolloState = apollo.cache.extract();

      return {
        ...appProps,
        apolloState,
      };
    }

    constructor(props) {
      super(props);
      // `getDataFromTree` renders the component first, the client is passed off as a property.
      // After that rendering is done using Next's normal rendering pipeline
      this.apolloClient = initApollo(props.apolloState, {});
    }

    render() {
      return <App {...this.props} apolloClient={this.apolloClient} />;
    }
  };
};
