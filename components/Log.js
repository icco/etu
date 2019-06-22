import { graphql } from "react-apollo";
import gql from "graphql-tag";
import Link from "next/link";

import Loading from "./Loading";
import ErrorMessage from "./ErrorMessage";

function Log({ data: { loading, error, log } }) {
  if (error) return <ErrorMessage message="Error loading log entry." />;
  if (loading) {
    return <Loading key={0} />;
  }

  return (
    <li className="mb5 ml4 mr3" key={"log-" + log.id}>
      <div className="f6 db pb1 gray">
        <span className="db dbi-ns mr3">{log.code}</span>
        <Link as={`/wiki/${log.project}`} href={`/wiki?id=${log.project}`}>
          <a className="db dbi-ns mr3">{log.project}</a>
        </Link>
        <span className="db dbi-ns mr3">{log.datetime}</span>
        <Link as={`/log/${log.id}`} href={`/log?id=${log.id}`}>
          <a className="db dbi-ns mr3">{log.id}</a>
        </Link>
      </div>
      <div>{log.description}</div>
    </li>
  );
}

export const userLog = gql`
  query getLog($id: ID!) {
    log(id: $id) {
      id
      code
      datetime
      description
      project
    }
  }
`;

export default graphql(userLog, {
  options: props => ({
    variables: {
      id: props.id,
    },
  }),
})(Log);
