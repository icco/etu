import Head from "next/head";

import App from "../components/App";
import Header from "../components/Header";
import Submit from "../components/submit";
import { checkLoggedIn } from "../lib/auth";

const Index = props => (
  <App>
    <Head>
      <title>Etu Time Tracking</title>
    </Head>
    <Header loggedInUser={props.loggedInUser} noLogo />
    <Submit loggedInUser={props.loggedInUser} />
  </App>
);

Index.getInitialProps = async ctx => {
  const { loggedInUser } = await checkLoggedIn(ctx.apolloClient);

  return {
    loggedInUser,
  };
};

export default Index;
