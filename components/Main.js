import { withRouter } from "next/router";
import React from "react";

import LogList from "../components/LogList";
import Submit from "../components/Submit";

class Main extends React.Component {
  render() {
    let content = (
      <>
        <h1 className="f-headline-ns f-subheadline mw7 center pa4">Please sign-in to see logs.</h1>
      </>
    );

    if (this.props.loggedInUser) {
      content = (
        <>
          <Submit />
          <LogList />
        </>
      );
    }

    return content;
  }
}

export default withRouter(Main);
