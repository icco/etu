import Head from "next/head";
import { withAuth, withLoginRequired } from "use-auth0-hooks";

import { withApollo } from "../lib/apollo";
import { useLoggedIn } from "../lib/auth";

import App from "../components/App";
import Header from "../components/Header";
import NotAuthorized from "../components/NotAuthorized";
import Submit from "../components/Submit";
import LogList from "../components/LogList";

const Page = ({ auth }) => {
  const { loggedInUser } = useLoggedIn();
  console.log(loggedInUser)
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
