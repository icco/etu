import React from "react";

import { logout } from "../../lib/auth0";

class SignOut extends React.Component {
  componentDidMount() {
    logout();
  }
  render() {
    return null;
  }
}

export default SignOut;
