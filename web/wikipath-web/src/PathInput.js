import React from 'react';
import './PathInput.css';

const submitHandler = onSubmit => e => {
  e.preventDefault();
  onSubmit();
};

const randomize = (setVal, name) => async e => {
  const resp = await fetch('/api/random');
  if (resp.ok) {
    const rj = await resp.json();
    setVal(name, rj.title);
  } else {
    alert("Couldn't find random article.");
  }
};

const PathInput = ({setVal, from, to, onSubmit}) => (
  <form onSubmit={submitHandler(onSubmit)}>
    <label className="PathInput_label">
      <span>From:</span>
      <input
        className="PathInput_input"
        type="text"
        onChange={e => setVal('from', e.target.value)}
        value={from}
      />
      <button className="PathInput_random" type="button" onClick={randomize(setVal, 'from')}>
        ?
      </button>
    </label>
    <label className="PathInput_label">
      <span>To:</span>
      <input
        className="PathInput_input"
        type="text"
        onChange={e => setVal('to', e.target.value)}
        value={to}
      />
      <button className="PathInput_random" type="button" onClick={randomize(setVal, 'to')}>
        ?
      </button>
    </label>
    <button className="PathInput_btn" type="submit">
      Go!
    </button>
  </form>
);

export default PathInput;
