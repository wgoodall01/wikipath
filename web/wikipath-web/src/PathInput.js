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

const PathField = ({label, name, setVal, val}) => (
  <div className="PathInput_field">
    <div>
      <label htmlFor={`PathField_input_${name}`}>{label}</label>
      <button
        tabIndex="-1"
        className="PathInput_random"
        type="button"
        onClick={randomize(setVal, name)}
      >
        random
      </button>
    </div>
    <input
      id={`PathField_input_${name}`}
      className="PathInput_input"
      type="text"
      onChange={e => setVal(name, e.target.value)}
      value={val}
    />
  </div>
);

const PathInput = ({setVal, from, to, onSubmit}) => (
  <form onSubmit={submitHandler(onSubmit)}>
    <div className="PathInput_fieldset">
      <PathField label="From" name="from" setVal={setVal} val={from} />
      <PathField label="To" name="to" setVal={setVal} val={to} />
    </div>
    <button className="PathInput_btn" type="submit">
      Go!
    </button>
  </form>
);

export default PathInput;
