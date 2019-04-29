import React from "react";
import gql from "graphql-tag";
import { Query, Mutation } from "react-apollo";
import { withRouter } from "next/router";
import Link from "next/link";

import { getToken } from "../lib/auth.js";
import ErrorMessage from "./ErrorMessage";
import Loading from "./Loading";

const baseUrl = process.env.GRAPHQL_ORIGIN.substring(
  0,
  process.env.GRAPHQL_ORIGIN.lastIndexOf("/")
);

const SaveLog = gql`
  mutation SaveLog(
    $content: String!
    $project: String!
    $code: String!
  ) {
    insertLog(input: {
      code: $code
      description: $content
      project: $project
    }) {
      id
      datetime
    }
  }
`;

class Submit extends React.Component {
  render() {
    return (
              <Mutation mutation={SaveLog}>
                {(insertLog, { data }) => (
      <div>
                  <form
                    onSubmit={e => {
                      e.preventDefault();
                      insertLog({
                        variables: {
    content: this.state.content,
    project : this.state.project,
    code: this.state.code ,
                        },
                      });
                    }}
className="pa4 black-80"
                  >
          <div className="measure mv2">
            <label for="name" className="f6 b db mb2">
              Code
            </label>
            <input
              id="name"
              className="input-reset ba b--black-20 pa2 mb2 db w-100"
              type="text"
              aria-describedby="name-desc"
            />
            <small id="name-desc" className="f6 black-60 db mb2">
              Helper text for the form control.
            </small>
          </div>
          <div className="measure mv2">
            <label for="name" className="f6 b db mb2">
              Log Entry
            </label>
            <input
              id="name"
              className="input-reset ba b--black-20 pa2 mb2 db w-100"
              type="text"
              aria-describedby="name-desc"
            />
            <small id="name-desc" className="f6 black-60 db mb2">
              Helper text for the form control.
            </small>
          </div>
          <div className="measure mv2">
            <label for="name" className="f6 b db mb2">
              Project
            </label>
            <input
              id="name"
              className="input-reset ba b--black-20 pa2 mb2 db w-100"
              type="text"
              aria-describedby="name-desc"
            />
            <small id="name-desc" className="f6 black-60 db mb2">
              Helper text for the form control.
            </small>
          </div>
          <input
            className="b ph3 pv2 input-reset ba b--black bg-transparent grow pointer f6 dib"
            type="submit"
            value="Submit"
          />
        </form>
      </div>
                )}
              </Mutation>
    );
  }
}

export default withRouter(Submit);
