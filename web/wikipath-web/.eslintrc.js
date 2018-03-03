module.exports = {
  extends: ['airbnb', 'prettier'],
  env: {browser: true},
  rules: {
    'react/jsx-filename-extension': false,
    'import/extensions': false,
    'jsx-a11y/label-has-for': {
      required: {some: ['nesting', 'id']}
    }
  }
};
