const path = require('path');
const webpack = require('webpack');
const packageJson = require('./package.json');

module.exports = {
    entry: './src/index.tsx',
    output: {
        path: path.resolve(__dirname, 'dist'),
        filename: 'main.js',
        library: ['com.mattermost.plugin-approver2'],
        libraryTarget: 'window'
    },
    resolve: {
        extensions: ['.ts', '.tsx', '.js', '.jsx'],
        modules: ['node_modules']
    },
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: 'ts-loader',
                exclude: /node_modules/
            }
        ]
    },
    externals: {
        react: 'React',
        'react-dom': 'ReactDOM',
        redux: 'Redux',
        'react-redux': 'ReactRedux',
        'mattermost-redux': 'MattermostRedux'
    },
    plugins: [
        new webpack.DefinePlugin({
            PLUGIN_VERSION: JSON.stringify(packageJson.version)
        })
    ],
    devtool: 'source-map',
    mode: process.env.NODE_ENV === 'production' ? 'production' : 'development'
};
