import React from 'react';
import './App.css';

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      status: 'READY', // one of "READY" "PROCESSING"
      from: '',
      to: '',
      duration: null,
      path: [],
      err: null
    };
  }

  async onSubmit() {
    const {from, to} = this.state;
    const url = `/api/query?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`;
    const resp = await fetch(url);
    const rj = await resp.json();
    this.setState({status: 'PROCESSING'});
    if (resp.ok) {
      // No error
      this.setState({
        status: 'READY',
        from: rj.from,
        to: rj.to,
        path: rj.path,
        duration: rj.duration,
        err: null
      });
    } else {
      // There's an error.
      this.setState({status: 'READY', path: null, duration: null, err: rj});
    }
  }

  onChange(e) {
    this.setState({[e.target.name]: e.target.value});
  }

  render() {
    const {to, from, path, duration, status, err} = this.state;

    let body;
    if (status === 'PROCESSING') {
      body = <h2>Processing...</h2>;
    } else {
      body = (
        <React.Fragment>
          <input
            className="App_input"
            type="text"
            name="from"
            onChange={this.onChange.bind(this)}
            value={from}
            placeholder="From"
          />
          <input
            className="App_input"
            type="text"
            name="to"
            onChange={this.onChange.bind(this)}
            value={to}
            placeholder="To"
          />
          <button onClick={this.onSubmit.bind(this)}>Go!</button>
          {err && (
            <div className="App_error">
              <h2>Error</h2>
              <p>{err.message}</p>
            </div>
          )}
          {path && <ol className="App_path">{path.map(e => <li key={e}>{e}</li>)}</ol>}
          {duration && <div className="App_duration">Done in {duration} seconds.</div>}
        </React.Fragment>
      );
    }

    return (
      <div className="App">
        <h1>Wikipath</h1>
        {body}
      </div>
    );
  }
}

export default App;
