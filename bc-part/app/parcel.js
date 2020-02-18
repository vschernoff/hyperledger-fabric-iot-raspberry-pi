const path = require('path');
const Bundler = require('parcel-bundler'); // eslint-disable-line import/no-extraneous-dependencies

const clientPath = path.join(__dirname, './src/index.js');

const PROD = process.env === 'production';

const deafaultConfig = {
  outDir: './dist',
  logLevel: 3,
  sourceMaps: false,
  detailedReport: true,
  cache: !PROD
};

const clientOpts = {
  outDir: './dist/static',
  outFile: 'bundle.js',
  minify: PROD,
  target: 'browser'
};

(async () => {
  const client = new Bundler(clientPath, Object.assign(deafaultConfig, clientOpts));

  await client.bundle();
  console.info('client bundled');
  process.exit(0);
})();
