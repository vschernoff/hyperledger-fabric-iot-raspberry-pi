const path = require('path');
const express = require('express');
const jwt = require('jsonwebtoken');

const {
  PORT = 3000,
  JWT_SECRET = 'hlfiot-jwt-secret',
  ORG = 'a'
} = process.env;

const MOCK_DATA = require('./__mock-data__');

const app = express();
const router = express.Router();

router.post('/login', async (_, res) =>
  res.json({
    jwt: jwt.sign({ user: 'test-hlfiot-user' }, JWT_SECRET)
  })
);

router.get('*', (_, __, next) => {
  setTimeout(next, 750);
});

router.get('/mock-api', async ({ query }, res) =>
  res.json(MOCK_DATA[query.type])
);

app.use(express.static(path.join(__dirname, '../../dist/client')));

const html = require('./html');

const renderer = async (req, res) => {
  const data = { org: ORG };
  return res.send(html(data));
};

router.use('*', renderer);
app.use(router);

app.listen(PORT, () => {
  console.info(`listening on port: ${PORT}`);
});
