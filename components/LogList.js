import { graphql } from "react-apollo";
import gql from "graphql-tag";
import Link from "next/link";

import Loading from "./Loading";
import Log from "./Log";
import ErrorMessage from "./ErrorMessage";

const PER_PAGE = 20;

function LogList({ data: { loading, error, logs } }) {
  if (error) return <ErrorMessage message="Error loading User's Logs." />;
  if (loading) {
    return <Loading key={0} />;
  }
  return (
    <section className="mw8">
      <ul className="list pl0" key="ul">
        {logs.map(log => (
          <Log key={log.id} id={log.id} data={{ log }} />
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
