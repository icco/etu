import React from "react";
import gql from "graphql-tag";
import { Query, Mutation } from "react-apollo";
import { withRouter } from "next/router";
import Link from "next/link";

import { getToken } from "../lib/auth.js";
import ErrorMessage from "./ErrorMessage";
import Loading from "./Loading";

const baseUrl = process.env.GRAPHQL_ORIGIN.substring(
  0,
  process.env.GRAPHQL_ORIGIN.lastIndexOf("/")
);

class Submit extends React.Component {
  render() {
    return <></>
  }
}

export default withRouter(Submit);
