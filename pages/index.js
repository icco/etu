import Head from "next/head";

import App from "../components/App";
import Header from "../components/Header";
import Main from "../components/Main";
import { checkLoggedIn } from "../lib/auth";
import { initApollo } from "../lib/init-apollo";

export default class extends React.Component {
  async componentDidMount() {
    const { loggedInUser } = await checkLoggedIn(initApollo());
    this.setState({ loggedInUser });
  }

  render() {
    if (!this.state || !this.state.loggedInUser) {
      this.state = { loggedInUser: null };
    }

    return (
      <App>
        <Head>
          <title>Etu Time Tracking</title>
        </Head>
        <Header loggedInUser={this.state.loggedInUser} noLogo />
        <Main loggedInUser={this.state.loggedInUser} />
      </App>
    );
  }
}
