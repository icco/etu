import Error from "next/error";

export default function NotAuthorized() {
  return <Error statusCode={403} title="Forbidden" />;
}
