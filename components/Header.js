import React from "react";
import Link from "next/link";
import { Logo, Loading } from "@icco/react-common";
import { useRouter } from "next/router";

import { useLoggedIn } from "../lib/auth";

export default function Header({ noLogo }) {
  const { pathname, query } = useRouter();
  const { loading, login, logout, loggedInUser, error } = useLoggedIn();

  if (error) {
    if (error.error != "consent_required") {
      throw error;
    }
  }

  const elements = {
    about: (
      <Link key="/about" href="/about" prefetch={false}>
        <a className="f6 link dib dim mr3 black mr4-ns">about</a>
      </Link>
    ),
    signin: (
      <a
        className="f6 link dib dim mr3 black mr4-ns pointer"
        onClick={() => login({ appState: { returnTo: { pathname, query } } })}
      >
        Sign In
      </a>
    ),
    signout: (
      <a
        className="f6 link dib dim mr3 black mr4-ns pointer"
        onClick={() => logout({ appState: { returnTo: { pathname, query } } })}
      >
        Sign Out
      </a>
    ),
    largelogo: (
      <header className="mv5 center mw6">
        <Link href="/">
          <a className="link dark-gray dim">
            <Logo size={200} className="center" />
            <h1 className="tc">Nat? Nat. Nat!</h1>
          </a>
        </Link>
      </header>
    ),
    smalllogo: (
      <Link href="/">
        <a className="link dark-gray dim">
          <Logo size={50} className="v-mid mh0-ns dib-ns center ph0 logo" />
        </a>
      </Link>
    ),
    adminlink: <></>,
  };

  let nav = <>{elements.signin}</>;

  if (loading) {
    nav = (
      <>
        <div className="dib h1">
          <Loading key={0} />
        </div>
      </>
    );
  }

  if (loggedInUser) {
    elements.adminlink = (
      <Link key="/admin" href="/admin">
        <a className="f6 link dib dim mr3 black mr4-ns">{loggedInUser.role}</a>
      </Link>
    );
    nav = (
      <>
        {elements.adminlink}
        {elements.signout}
      </>
    );
  }

  return (
    <>
      <nav className="flex justify-between ttc">
        <div className="flex items-center pa3">
          {noLogo ? elements.smalllogo : ""}
        </div>
        <div className="flex-grow pa3 flex items-center">
          {elements.about}
          {nav}
        </div>
      </nav>
      {noLogo ? "" : elements.largelogo}
    </>
  );
}
