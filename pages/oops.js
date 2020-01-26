import React from "react";
import { useRouter } from "next/router";
import ErrorPage from "next/error";

export default function Oops() {
  const router = useRouter();
  const { message } = router.query;

  return <ErrorPage statusCode={500} title={message || "Unknown Error"} />;
}
