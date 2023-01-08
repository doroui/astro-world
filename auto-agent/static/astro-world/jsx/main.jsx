import {ReactDOM} from './deps.js';

// npm install -g react-tools
// jsx -w -x jsx public/js public/js

import {App} from './app.js';
import {User} from './user.js';

function getQueryStringValue(key) {
  return unescape(
    window.location.search.replace(
      new RegExp(
        '^(?:.*[&\\?]' + encodeURIComponent(key) + '(?:\\=([^&]*))?)?.*$',
        'i',
      ),
      '$1',
    ),
  );
}

// Begin side-effect
var username = getQueryStringValue('user');

function main() {
  if (!username) {
    window.location = 'index.html';
    return;
  } else {
    var user = new User(username);

    // React.renderComponent(<App variableModels={variableModels} user={user}/>,
    //                   document.body);

    user.loadAllUserData(function () {
      ReactDOM.render(<App user={user} />, document.getElementById('main'));
    });
  }
}

main();
