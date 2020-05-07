import React from "react"
import Head from "next/head";
import { withAuth, withLoginRequired } from "use-auth0-hooks";
import { ErrorMessage, Loading } from "@icco/react-common";

import { withApollo } from "../lib/apollo";
import { useLoggedIn } from "../lib/auth";

import App from "../components/App";
import Header from "../components/Header";
import NotAuthorized from "../components/NotAuthorized";
import Submit from "../components/Submit";
import LogList from "../components/LogList";

const Page = ({ auth }) => {
  const { loggedInUser, loading, error } = useLoggedIn();

  if (error) {
    return <ErrorMessage message="Error loading User's Logs." />;
  }

  if (loading) {
    return <Loading key={0} />;
  }

  if (!loggedInUser) {
    return <NotAuthorized />;
  }

  return (
    <App>
      <Head>
        <title>Etu Time Tracking</title>
      </Head>

      <Header noLogo />

      <Submit loggedInUser={loggedInUser} />
      <LogList loggedInUser={loggedInUser} />
    </App>
  );
};

export default withLoginRequired(withAuth(withApollo(Page)));
