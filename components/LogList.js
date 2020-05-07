import gql from "graphql-tag";
import { ErrorMessage, Loading } from "@icco/react-common";
import { useQuery } from "@apollo/react-hooks";

import Log from "./Log";

const PER_PAGE = 20;

export const userLogs = gql`
  query logs {
    logs {
      id
    }
  }
`;

export default function LogList({ loggedInUser }) {
  const { loading, error, data } = useQuery(userLogs, {
    variables: {
      offset: 0,
      perpage: PER_PAGE,
    },
    fetchPolicy: "no-cache",
  });

  if (error) {
    return <ErrorMessage message="Error loading User's Logs." />;
  }

  if (loading) {
    return <Loading key={0} />;
  }

  if (!loggedInUser) {
    return <ErrorMessage message="User not logged in." />;
  }

  const { logs } = data;

  return (
    <section className="mw8">
      <ul className="list pl0" key="ul">
        {logs.map((log) => (
          <Log key={log.id} id={log.id} />
        ))}
      </ul>
    </section>
  );
}
