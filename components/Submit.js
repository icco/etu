import React from "react";
import gql from "graphql-tag";
import { Query, Mutation } from "react-apollo";
import { withRouter } from "next/router";
import Link from "next/link";
import { ErrorMessage, Loading } from "@icco/react-common";

import { getToken } from "../lib/auth.js";

const baseUrl = process.env.GRAPHQL_ORIGIN.substring(
  0,
  process.env.GRAPHQL_ORIGIN.lastIndexOf("/")
);

const SaveLog = gql`
  mutation SaveLog($content: String!, $project: String!, $code: String!) {
    insertLog(
      input: { code: $code, description: $content, project: $project }
    ) {
      id
      datetime
    }
  }
`;

class Submit extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      content: "",
      project: "",
      code: "",
    };
  }

  handleBasicChange = event => {
    const target = event.target;
    const value = target.type === "checkbox" ? target.checked : target.value;
    let name = target.name;

    if (name == "") {
      name = target.id;
    }

    this.setState({
      [name]: value,
    });
  };

  handleEditorChange = value => {
    this.setState({
      content: value(),
    });
  };

  render() {
    return (
      <Mutation mutation={SaveLog}>
        {(insertLog, { loading, error, data }) => {
          if (loading) {
            return <Loading key={0} />;
          }

          if (error) {
            return <ErrorMessage message="Page not found." />;
          }

          return (
            <div>
              <form
                autoComplete="off"
                onSubmit={e => {
                  e.preventDefault();
                  insertLog({
                    variables: {
                      content: this.state.content,
                      project: this.state.project,
                      code: this.state.code,
                    },
                  });
                }}
                className="pa4 black-80"
              >
                <div className="measure mv2">
                  <label htmlFor="code" className="f6 b db mb2">
                    Code
                  </label>
                  <input
                    id="code"
                    className="input-reset db mb2"
                    type="text"
                    aria-describedby="code-desc"
                    onChange={this.handleBasicChange}
                    maxLength="3"
                    style={{
                      border: "none",
                      width: "4.5ch",
                      background:
                        "repeating-linear-gradient(90deg, dimgrey 0, dimgrey 1ch, transparent 0, transparent 1.5ch) 0 100%/100% 2px no-repeat",
                      font: "5ch monospace",
                      letterSpacing: ".5ch",
                    }}
                  />
                  <small id="code-desc" className="f6 black-60 db mb2">
                    <p>Category, Focus, Introversion</p>
                    <ol>
                      <li>Educating</li>
                      <li>Building</li>
                      <li>Living</li>
                    </ol>
                  </small>
                </div>

                <div className="measure mv2">
                  <label htmlFor="project" className="f6 b db mb2">
                    Project
                  </label>
                  <input
                    id="project"
                    className="input-reset ba b--black-20 pa2 mb2 db w-100"
                    type="text"
                    onChange={this.handleBasicChange}
                    aria-describedby="project-desc"
                  />
                  <small id="project-desc" className="f6 black-60 db mb2">
                    What project is this?
                  </small>
                </div>

                <div className="measure mv2">
                  <label htmlFor="content" className="f6 b db mb2">
                    Log Entry
                  </label>
                  <input
                    id="content"
                    className="input-reset ba b--black-20 pa2 mb2 db w-100"
                    type="text"
                    aria-describedby="content-desc"
                    onChange={this.handleBasicChange}
                  />
                  <small id="content-desc" className="f6 black-60 db mb2">
                    What's your update?
                  </small>
                </div>

                <input
                  className="b ph3 pv2 input-reset ba b--black bg-transparent grow pointer f6 dib"
                  type="submit"
                  value="Submit"
                />
              </form>
            </div>
          );
        }}
      </Mutation>
    );
  }
}

export default withRouter(Submit);
