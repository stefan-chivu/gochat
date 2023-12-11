const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function (app) {
    app.use(
        '/api',
        createProxyMiddleware({
            target: 'http://12.12.12.10:8080',
            changeOrigin: true,
        })
    );
};