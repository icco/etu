import Head from "next/head";

import App from "../components/App";
import Header from "../components/Header";
import Submit from "../components/Submit";
import LogList from "../components/LogList";
import { checkLoggedIn } from "../lib/auth";

const Index = props => {
  let content = <div className="pa4">Please sign in to see logs.</div>;

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
