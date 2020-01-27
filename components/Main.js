import { ErrorMessage, Loading } from "@icco/react-common";

import { useLoggedIn } from "../lib/auth";

import Submit from "./Submit";
import LogList from "./LogList";

export default function Main() {
  const { loggedInUser, loading, error } = useLoggedIn();

  if (error) {
    return <ErrorMessage message="Error loading User's Logs." />;
  }

  if (loading) {
    return <Loading key={0} />;
  }

  if (!loggedInUser) {
    return <ErrorMessage message="User not logged in." />;
  }

  return (
    <>
      <Submit loggedInUser={loggedInUser} />
      <LogList loggedInUser={loggedInUser} />
    </>
  );
}
