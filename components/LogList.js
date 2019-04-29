import { graphql } from "react-apollo";
import gql from "graphql-tag";
import Link from "next/link";

import Loading from "./Loading";
import ErrorMessage from "./ErrorMessage";

const PER_PAGE = 20;

function LogList({ data: { loading, error, logs } }) {
  if (error) return <ErrorMessage message="Error loading User's Logs." />;
  if (loading) {
    return <Loading key={0} />;
  }
  return (
    <section className="mw8 center">
      <ul className="list pl0" key="ul">
        {logs.map(log => (
          <li className="mb5 ml4 mr3" key={"log-list-" + log.id}>
            <div className="f6 db pb1 gray">
              <span className="db dbi-ns mr3">{log.code}</span>
              <Link
                as={`/wiki/${log.project}`}
                href={`/wiki?id=${log.project}`}
              >
                <a className="db dbi-ns mr3">{log.project}</a>
              </Link>
              <span className="db dbi-ns mr3">{log.datetime}</span>
              <Link as={`/log/${log.id}`} href={`/log?id=${log.id}`}>
                <a className="db dbi-ns mr3">{log.id}</a>
              </Link>
            </div>
            <div>{log.description}</div>
          </li>
        ))}
      </ul>
    </section>
  );
}

export const userLogs = gql`
  query {
    logs {
      id
      code
      datetime
      description
      project
    }
  }
`;

export default graphql(userLogs, {
  options: {
    variables: {
      offset: 0,
      perpage: PER_PAGE,
    },
  },
})(LogList);
