import React from 'react';
import PropTypes from 'prop-types';
import './PathDisplay.css';

const wikiRoot = process.env.REACT_APP_WIKI_ROOT || 'https://en.wikipedia.org/wiki';

const PathDisplay = ({path}) => (
  <ol className="PathDisplay">
    {path.map(e => (
      <li className="PathDisplay_item" key={e}>
        <a
          className="PathDisplay_link"
          href={`${wikiRoot}/${e.replace(/ /g, '_')}`}
          target="_blank"
          rel="noopener"
        >
          {e}
        </a>
      </li>
    ))}
  </ol>
);

PathDisplay.propTypes = {
  path: PropTypes.arrayOf(PropTypes.string).isRequired
};

export default PathDisplay;
