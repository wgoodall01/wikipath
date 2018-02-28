import React from 'react';
import './PathDisplay.css';

const PathDisplay = ({path}) => (
  <ol>
    {path.map((e, i) => (
      <li className="PathDisplay_item" key={i}>
        <a
          className="PathDisplay_link"
          href={`https://en.wikipedia.org/wiki/${e.replace(/\ /g, '_')}`}
          target="_blank"
          rel="noopener"
        >
          {e}
        </a>
      </li>
    ))}
  </ol>
);

export default PathDisplay;
