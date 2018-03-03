import React from 'react';
import PropTypes from 'prop-types';
import './PathInput.css';

const submitHandler = onSubmit => e => {
  e.preventDefault();
  onSubmit();
};

const randomize = (setVal, name) => async () => {
  const resp = await fetch('/api/random');
  if (resp.ok) {
    const rj = await resp.json();
    setVal(name, rj.title);
  } else {
    setVal(name, '[not found]');
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

PathField.propTypes = {
  setVal: PropTypes.func.isRequired,
  name: PropTypes.string.isRequired,
  label: PropTypes.string.isRequired,
  val: PropTypes.string.isRequired
};

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

PathInput.propTypes = {
  setVal: PropTypes.func.isRequired,
  onSubmit: PropTypes.func.isRequired,
  from: PropTypes.string.isRequired,
  to: PropTypes.string.isRequired
};

export default PathInput;
