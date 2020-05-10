import ErrorPage from "next/error";

function Error({ statusCode }) {
  return <ErrorPage statusCode={statusCode} />;
}

Error.getServerSideProps = ({ res, err }) => {
  const statusCode = res ? res.statusCode : err ? err.statusCode : 404;
  return { statusCode };
};

export default Error;
