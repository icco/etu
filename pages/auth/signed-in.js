import React from "react";
import Router from "next/router";

import { setToken } from "../../lib/auth";
import { parseHash } from "../../lib/auth0";

export default class SignedIn extends React.Component {
  componentDidMount() {
    parseHash((err, result) => {
      if (err) {
        return;
      }

      setToken(result.idToken, result.accessToken);
      Router.push("/");
    });
  }

  render() {
    return null;
  }
}
