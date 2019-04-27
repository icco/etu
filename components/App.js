import React from "react";
import Router from "next/router";
import * as gtag from "../lib/gtag";
import * as fathom from "../lib/fathom";

Router.onRouteChangeComplete = url => {
  gtag.pageview(url);
  fathom.pageview(url);
};

export default ({ children }) => <main>{children}</main>;
