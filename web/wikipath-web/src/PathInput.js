import React from 'react';
import './PathInput.css';

const submitHandler = onSubmit => e => {
  e.preventDefault();
  onSubmit();
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
    </label>
    <label className="PathInput_label">
      <span>To:</span>
      <input
        className="PathInput_input"
        type="text"
        onChange={e => setVal('to', e.target.value)}
        value={to}
      />
    </label>
    <button className="PathInput_btn" type="submit">
      Go!
    </button>
  </form>
);

export default PathInput;
