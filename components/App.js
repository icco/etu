import React from "react";
import Router from "next/router";
import * as fathom from "../lib/fathom";

Router.onRouteChangeComplete = url => {
  fathom.pageview(url);
};

export default ({ children }) => <main>{children}</main>;
