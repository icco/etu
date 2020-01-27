import Head from "next/head";
import { withAuth, withLoginRequired } from "use-auth0-hooks";

import { withApollo } from "../lib/apollo";
import { AuthOptions } from "../lib/auth";

import App from "../components/App";
import Main from "../components/Main";
import Header from "../components/Header";
import NotAuthorized from "../components/NotAuthorized";

const Page = ({ auth }) => {
  return (
    <App>
      <Head>
        <title>Etu Time Tracking</title>
      </Head>

      <Header noLogo />

      <Main />
    </App>
  );
};

export default withLoginRequired(withAuth(withApollo(Page), AuthOptions));
