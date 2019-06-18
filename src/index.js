import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';

// Use aXe a11y testing in development.
var axe = require('react-axe');

if (process.env.NODE_ENV !== 'production') {
  axe(React, ReactDOM, 1000);
}

ReactDOM.render(<App />, document.getElementById('root'));
