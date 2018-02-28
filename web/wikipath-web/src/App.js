import React from 'react';
import PathInput from './PathInput.js';
import PathDisplay from './PathDisplay.js';
import './App.css';

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      status: 'READY', // one of "READY" "PROCESSING"
      from: '',
      to: '',
      result: null,
      //duration: null,
      //touched: null,
      //path: [],
      err: null
    };
  }

  async onSubmit() {
    const {from, to} = this.state;
    const url = `/api/query?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`;
    this.setState({status: 'PROCESSING'});
    const resp = await fetch(url);
    const rj = await resp.json();
    if (resp.ok) {
      // No error
      this.setState({
        status: 'READY',
        from: rj.from,
        to: rj.to,
        result: {
          path: rj.path,
          duration: rj.duration,
          touched: rj.touched
        },
        err: null
      });
    } else {
      // There's an error.
      this.setState({status: 'READY', result: null, err: rj});
    }
  }

  setVal(name, val) {
    this.setState({[name]: val});
  }

  render() {
    const {to, from, result, status, err} = this.state;
    const {path, duration, touched} = result || {};

    return (
      <div className="App">
        <h1>Wikipath</h1>
        <PathInput
          from={from}
          to={to}
          setVal={this.setVal.bind(this)}
          onSubmit={this.onSubmit.bind(this)}
        />
        {status == 'PROCESSING' && <h2>Processing...</h2>}
        {err && (
          <div className="App_error">
            <h2>Error</h2>
            <p>{err.message}</p>
          </div>
        )}
        {path && <PathDisplay path={path} />}
        {duration &&
          touched && (
            <div className="App_status">
              Searched {touched} articles in {duration} seconds
            </div>
          )}
      </div>
    );
  }
}

export default App;
