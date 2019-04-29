import React from "react";

import { authorize } from "../../lib/auth0";

class SignIn extends React.Component {
  componentDidMount() {
    authorize();
  }
  render() {
    return null;
  }
}

export default SignIn;
