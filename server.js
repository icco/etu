const express = require("express");
const next = require("next");
const helmet = require("helmet");
const { SSLMiddleware } = require("@icco/react-common");

const dev = process.env.NODE_ENV !== "production";
const port = parseInt(process.env.PORT, 10) || 3000;
const app = next({ dev });
const handle = app.getRequestHandler();

app.prepare().then(() => {
  const server = express();
  server.use(helmet());

  server.set("trust proxy", true);
  server.use(SSLMiddleware());

  server.get("*", (req, res) => {
    return handle(req, res);
  });

  server.listen(port, err => {
    if (err) throw err;
    console.log(`> Ready on http://localhost:${port}`);
  });
});
