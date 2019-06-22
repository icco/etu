import Head from "next/head";

import App from "../components/App";
import Header from "../components/Header";
import LogList from "../components/LogList";
import Main from "../components/Main";
import Submit from "../components/Submit";
import { checkLoggedIn } from "../lib/auth";

const Index = props => {
  return (
    <App>
      <Head>
        <title>Etu Time Tracking</title>
      </Head>
      <Header loggedInUser={props.loggedInUser} noLogo />
      <Main loggedInUser={props.loggedInUser} />
    </App>
  );
};

Index.getInitialProps = async ctx => {
  const { loggedInUser } = await checkLoggedIn(ctx.apolloClient);

  return {
    loggedInUser,
  };
};

export default Index;
