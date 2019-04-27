import React from "react";
import ReactLoading from "react-loading";

const Loading = () => (
  <ReactLoading
    key="loading"
    type={"bars"}
    height={64}
    width={64}
    className="center dark-grey w3"
    style={{ fill: "#333" }}
  />
);

export default Loading;
