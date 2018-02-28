import React from 'react';
import './PathDisplay.css';

const wikiRoot = process.env.REACT_APP_WIKI_ROOT || 'https://en.wikipedia.org/wiki';

const PathDisplay = ({path}) => (
  <ol>
    {path.map((e, i) => (
      <li className="PathDisplay_item" key={i}>
        <a
          className="PathDisplay_link"
          href={wikiRoot + '/' + e.replace(/\ /g, '_')}
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
