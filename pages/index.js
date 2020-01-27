import Head from "next/head";
import { withAuth, withLoginRequired } from "use-auth0-hooks";

import { withApollo } from "../lib/apollo";
import { AuthOptions } from "../lib/auth";

import App from "../components/App";
import Header from "../components/Header";
import LogList from "../components/LogList";
import NotAuthorized from "../components/NotAuthorized";
import Submit from "../components/Submit";

const Page = ({ auth }) => {
  return (
    <App>
      <Head>
        <title>Etu Time Tracking</title>
      </Head>

      <Header noLogo />

      <Submit />
      <LogList />
    </App>
  );
};

export default withLoginRequired(withAuth(withApollo(Page), AuthOptions));
