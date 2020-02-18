module.exports = state => `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>HLF IOT</title>
    <link rel="shortcut icon" href="/favicon.b08504fd.ico">
    <link rel="stylesheet" href="/index.css" />
</head>
<body>
<div id="root"></div>
<script>
  window.__STATE__ = ${JSON.stringify(state)}
</script>
<script src="/index.js"></script>
</body>
</html>
`;
