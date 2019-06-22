import { withRouter } from "next/router";
import React from "react";

class Main extends React.Component {
  render() {
    let content = (
      <>
        <h1 class="f-headline mw7 center">Please sign-in to see logs.</h1>
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
