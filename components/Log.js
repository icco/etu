import { graphql } from "react-apollo";
import gql from "graphql-tag";
import Link from "next/link";

import Loading from "./Loading";
import ErrorMessage from "./ErrorMessage";

function Color(code) {
  let pieces = code.split("");
  if (pieces.length != 3) {
    return colors["black"];
  }

  let category = "black";
  switch (pieces[0]) {
    case "1":
      category = "orange";
      break;
    case "2":
      category = "gray";
      break;
    case "3":
      category = "teal";
      break;
  }

  let focus = Number.parseInt(pieces[1], 10);
  let introversion = Number.parseInt(pieces[2], 10);
  let idx = Math.floor((focus + introversion) / 2);
  console.log(pieces, category, idx, colors[category]);

  return colors[category][idx];
}

function Log({ data: { loading, error, log } }) {
  if (error) return <ErrorMessage message="Error loading log entry." />;
  if (loading) {
    return <Loading key={0} />;
  }

  return (
    <li className="mb5 ml4 mr3" key={"log-" + log.id}>
      <div className="f6 db pb1 gray">
        <div className="mb1">
          <svg className="v-mid h2 w1">
            <rect
              x="0"
              y="0"
              width="100%"
              height="100%"
              rx="2"
              ry="2"
              fill={Color(log.code)}
            ></rect>
          </svg>
          <span className="mh3 f3 v-mid">{log.code}</span>
        </div>
        <Link as={`/wiki/${log.project}`} href={`/wiki?id=${log.project}`}>
          <a className="db ml4">{log.project}</a>
        </Link>
        <span className="db ml4">
          <Link as={`/log/${log.id}`} href={`/log?id=${log.id}`}>
            <a className="mr3">{log.datetime}</a>
          </Link>
        </span>
      </div>
      <div className="db ml4">{log.description}</div>
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

// https://palx.jxnblk.com/1a616c.json
const colors = {
  base: "#1a616c",
  black: "#344143",
  gray: [
    "#f8f9f9",
    "#eceded",
    "#dee1e1",
    "#d0d3d3",
    "#c0c4c5",
    "#aeb4b4",
    "#99a1a2",
    "#818a8b",
    "#636e6f",
    "#344143",
  ],
  cyan: [
    "#e7eeef",
    "#cddcdf",
    "#afc8cb",
    "#8cafb5",
    "#5f9098",
    "#1a616c",
    "#175760",
    "#144b54",
    "#103d44",
    "#0b2a2f",
  ],
  blue: [
    "#e9ecf1",
    "#d1d7e1",
    "#b6bfd0",
    "#94a2ba",
    "#697c9e",
    "#1a386c",
    "#173160",
    "#132a52",
    "#0f2241",
    "#0a162a",
  ],
  indigo: [
    "#ecebf2",
    "#d6d4e3",
    "#bebad3",
    "#a09bbf",
    "#7771a3",
    "#251a6c",
    "#20175f",
    "#1b1351",
    "#150f3f",
    "#0d0927",
  ],
  violet: [
    "#efeaf1",
    "#ddd3e3",
    "#c8b8d2",
    "#af98bd",
    "#8e6da1",
    "#4e1a6c",
    "#451760",
    "#3b1352",
    "#2f0f41",
    "#1e0a2a",
  ],
  fuschia: [
    "#f1eaf0",
    "#e2d2e0",
    "#d1b8ce",
    "#bc97b7",
    "#a06c99",
    "#6c1a61",
    "#601756",
    "#53144a",
    "#43103c",
    "#2d0a28",
  ],
  pink: [
    "#f1eaed",
    "#e2d3d9",
    "#d1b8c1",
    "#bd98a5",
    "#a16d80",
    "#6c1a38",
    "#601732",
    "#53142b",
    "#421022",
    "#2c0a17",
  ],
  red: [
    "#f1eae9",
    "#e2d4d2",
    "#d0bab7",
    "#bb9b96",
    "#9f716a",
    "#6c251a",
    "#602117",
    "#531c14",
    "#421610",
    "#2c0f0a",
  ],
  orange: [
    "#efece7",
    "#dfd8cd",
    "#cbc1af",
    "#b5a68c",
    "#98835f",
    "#6c4e1a",
    "#604517",
    "#533c14",
    "#443110",
    "#2e210b",
  ],
  yellow: [
    "#edeee6",
    "#daddca",
    "#c5c9ab",
    "#abb186",
    "#8c9459",
    "#616c1a",
    "#576117",
    "#4b5414",
    "#3e4510",
    "#2b300b",
  ],
  lime: [
    "#eaefe6",
    "#d2decc",
    "#b8caad",
    "#98b389",
    "#71965c",
    "#386c1a",
    "#326117",
    "#2b5414",
    "#234410",
    "#182f0b",
  ],
  green: [
    "#e7efe8",
    "#cddfcf",
    "#afcbb3",
    "#8bb591",
    "#5e9866",
    "#1a6c25",
    "#176121",
    "#14541c",
    "#104417",
    "#0b2f10",
  ],
  teal: [
    "#e7efec",
    "#ccded8",
    "#aecbc0",
    "#8ab4a5",
    "#5d9782",
    "#1a6c4e",
    "#176146",
    "#14543c",
    "#104431",
    "#0b2f22",
  ],
};
