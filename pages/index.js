import Head from "next/head";

import App from "../components/App";
import Header from "../components/Header";
import Submit from "../components/Submit";
import LogList from "../components/LogList";
import { checkLoggedIn } from "../lib/auth";

const Index = props => {
  let content = (
    <h1 class="f3 f1-m f-headline-l mw7 center">Please sign-in to see logs.</h1>
  );

  if (props.loggedInUser) {
    content = (
      <>
        <Submit />
        <LogList />
      </>
    );
  }

  return (
    <App>
      <Head>
        <title>Etu Time Tracking</title>
      </Head>
      <Header loggedInUser={props.loggedInUser} noLogo />
      {content}
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
