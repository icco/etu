const getAuth0 = options => {
  const auth0 = require("auth0-js");

  const AUTH0_CLIENT_ID = "MwFD0COlI4F4AWvOZThe1psOIletecnL";
  const AUTH0_DOMAIN = "icco.auth0.com";

  return new auth0.WebAuth({
    clientID: AUTH0_CLIENT_ID,
    domain: AUTH0_DOMAIN,
  });
};

const getBaseUrl = () => `${window.location.protocol}//${window.location.host}`;

const getOptions = container => {
  return {
    audience: "https://natwelch.com",
    responseType: "token id_token",
    redirectUri: `${getBaseUrl()}/auth/signed-in`,
    scope: "openid email",
  };
};

export const authorize = () => getAuth0().authorize(getOptions());
export const logout = () => getAuth0().logout({ returnTo: getBaseUrl() });
export const parseHash = callback => getAuth0().parseHash(callback);
