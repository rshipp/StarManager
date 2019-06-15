'use strict';

class StarManager extends React.Component {
  render() {
    return (
      <h1>StarManager</h1>
    );
  }
}

let domContainer = document.querySelector('#StarManager');
ReactDOM.render(<StarManager />, domContainer);
