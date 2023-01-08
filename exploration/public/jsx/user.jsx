export class User {
  constructor(name) {
    this.username = username;
    this.oldCart = null;
    this.newCart = null;
    this.results = null;
    this.currentChallenge = null;
  }
  loadAllUserData(renderCallback) {
    var self = this;

    //var challengePromise = self.loadUserChallengeData();
    // var cartPromise = challengePromise.then(function() {
    //   self.loadUserResultData();
    // });
    var cartPromise = self.loadUserResultData();
    cartPromise.then(renderCallback, function (error) {
      console.error('Failed!', error);
    });
  }
  // passing in self because otherwise, the scope can be screwed up if
  //     this is called from Promise
  loadUserChallengeData() {
    var self = this;
    var promise = new Promise(function (resolve, reject) {
      var challengeReq = new XMLHttpRequest();
      challengeReq.onload = function () {
        //self.results = JSON.parse(challengeReq.responseText);
        resolve(self);
      };
      challengeReq.onerror = function () {
        reject(Error('It broke'));
      };
      challengeReq.open(
        'GET',
        '/userchallenge/' + self.username + '/findallchallenges',
      );
      challengeReq.send(null);
    });

    return promise;
  }
  // passing in self because otherwise, the scope can be screwed up if
  //     this is called from Promise
  loadUserResultData() {
    var self = this;
    var promise = new Promise(function (resolve, reject) {
      var resultsReq = new XMLHttpRequest();
      resultsReq.onload = function () {
        debugger;
        self.results = JSON.parse(resultsReq.responseText);
        resolve(self);
      };
      resultsReq.onerror = function () {
        reject(Error('It broke'));
      };
      resultsReq.open('GET', '/usercart/' + self.username + '/findallcarts');
      resultsReq.send(null);
    });

    return promise;
  }
  // DELETE:Replaced by loadUserResultData using Promise
  // getUserData: function(username, callback) {
  //   var self = this;
  //   var xhr = new XMLHttpRequest();
  //   xhr.onload = function() {
  //     self.results = JSON.parse(xhr.responseText);
  //     callback();
  //   };
  //   xhr.open('GET', '/usercart/' + this.username + '/findallcarts');
  //   xhr.send(null);
  // },
  updateCart(result) {
    if (this.oldCart == null) {
      this.oldCart = result;
      return;
    }
    var latestCart = this.oldCart;
    if (this.newCart != null) {
      latestCart = this.newCart;
    }
    var ivnames = variableModels.iVariables.map(function (iv) {
      return iv.name;
    });
    for (var i = 0; i < ivnames.length; i++) {
      if (result[ivnames[i]] != latestCart[ivnames[i]]) {
        this.oldCart = latestCart;
        this.newCart = result;
        return;
      }
    }
  }
  addResult(result, renderCallback) {
    var self = this;

    self.updateCart(result);
    debugger;
    var addCartPromise = new Promise(function (resolve, reject) {
      var xhr = new XMLHttpRequest();
      xhr.onload = function () {
        resolve(self);
      };
      xhr.error = function () {
        reject();
      };
      xhr.open('POST', '/usercart/' + self.username + '/addcartdata');
      xhr.setRequestHeader('Content-Type', 'application/json');
      xhr.send(JSON.stringify(result));
    });

    var loadUserCartPromise = addCartPromise.then(function () {
      debugger;
      return self.loadUserResultData();
    });

    loadUserCartPromise.then(renderCallback);
  }
  enterChallenge(renderCallback) {
    // if (!this.currentChallenge) {
    //   // if the user data are empty, receive it
    //   this.loadAllUserData(renderCallback);
    // } else {
    renderCallback();
    // }
  }
}

var UserResultData = React.createClass({
  render: function () {
    var user = this.props.user;
    var variableModels = this.props.variableModels;
    var ivnames = variableModels.iVariables.map(function (iv) {
      return iv.name;
    });
    var resultsDisplay = [];
    var allDisplay = [user.results.length];

    switch (this.props.mode) {
      /*      case 'all':
        var results = user.results;
        resultsDisplay = results.map(function(result) {
          var index = result[variableModels.dvResultCount];
          return <UserResult variableModels={variableModels} data={result} index={'#' + index}/>;
        });
        break;*/
      case 'notebook':
        var newDisplay = null;
        var oldDisplay = null;
        for (var j = 0; j < user.results.length; j++) {
          var result = user.results[j];
          var isNew = true;
          var isOld = true;
          for (var i = 0; i < ivnames.length; i++) {
            if (
              !user.newCart ||
              result[ivnames[i]] != user.newCart[ivnames[i]]
            ) {
              isNew = false;
            }
            if (
              !user.oldCart ||
              result[ivnames[i]] != user.oldCart[ivnames[i]]
            ) {
              isOld = false;
            }
          }
          var index = result[variableModels.dvResultCount];
          allDisplay[index - 1] = (
            <UserResult
              variableModels={variableModels}
              data={result}
              index={'#' + index}
            />
          );

          if (isNew) {
            newDisplay = (
              <UserResult
                variableModels={variableModels}
                data={result}
                index={'#' + index + ' (Newly Saved)'}
              />
            );
          } else if (isOld) {
            oldDisplay = (
              <UserResult
                variableModels={variableModels}
                data={result}
                index={'#' + index + ' (Last Saved)'}
              />
            );
          }
        }
        if (newDisplay) {
          resultsDisplay.push(newDisplay);
        }
        if (oldDisplay) {
          resultsDisplay.push(oldDisplay);
        }
        break;
    }

    var headers = variableModels.iVariables.map(function (iv) {
      //return <th><VariableImage name={iv.name}/>{iv.label}</th>;
      return <th>{iv.label}</th>;
    });

    return (
      <table className="result">
        <thead>
          <tr>
            <th></th>
            {headers}
            <th>{variableModels.dvLabel}</th>
          </tr>
        </thead>
        <tbody>
          {resultsDisplay}
          <tr>
            <td co={headers.length}>All Results</td>
          </tr>
          {allDisplay}
        </tbody>
      </table>
    );
  },
});

var UserResult = React.createClass({
  render: function () {
    var variableModels = this.props.variableModels;
    var data = this.props.data;
    var dvValues = data[variableModels.dvName].join(', ');
    var index = this.props.index;

    var variables = variableModels.iVariables.map(function (variable) {
      return <UserResultSelection iv={variable} value={data[variable.name]} />;
    });

    return (
      <tr>
        <td>Cart {index} :</td>
        {variables}
        <td>{dvValues}</td>
      </tr>
    );
  },
});

var UserResultSelection = React.createClass({
  getDisplayValue: function (value) {
    var options = this.props.iv.options;
    for (var i = 0; i < options.length; i++) {
      if (options[i].value == value) {
        return options[i].label;
      }
    }
    return null;
  },

  render: function () {
    var iv = this.props.iv;
    var ivValue = this.getDisplayValue(this.props.value);
    return <td>{ivValue}</td>;
  },
});
