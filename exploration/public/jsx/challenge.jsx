// function ChallengeHandler(challenge) {
//   this.username = username;
//   this.oldCart = null;
//   this.newCart = null;
//   this.results = null;
//   this.currentChallengeID = null;
// }

// User.prototype = {

//   loadAllUserData: function(renderCallback) {
//     var cartPromise = this.loadUserResultData(this.username);

//     var challengePromise = cartPromise.then((username) =>{
//                                               return this.loadUserChallengeData(username);
//                                             });
//     challengePromise.then(renderCallback, (error) =>{
//                                             console.error("Failed!", error);
//                                           });
//   },

// };

export class Challenge extends React.Component {
  // getInitialState() {
  //   this.setState({ enabled: false });
  //   debugger;
  //   return {};
  // }

  state = { enabled: false };

  handleChange = event => {
    // var state = {};
    // state[event.target.id] = event.target.value;
    this.setState({ enabled: true });
  };

  handleSubmit = event => {
    event.preventDefault();

    this.post(this.state);
  };

  post(data) {
    /*    if (!this.isEnabled())
      return;

    var xhr = new XMLHttpRequest();
    xhr.onload = () => {
      if (this.props.onComplete) {
        this.props.onComplete(JSON.parse(xhr.responseText));
      }
    };
    xhr.open('POST', '/carts/gettrips');
    xhr.setRequestHeader('Content-Type', 'application/json');
    xhr.send(JSON.stringify(data));
*/
  }

  isEnabled() {
    return this.state.enabled;
  }

  render() {
    var user = this.props.user;
    var variableModels = this.props.variableModels;
    var ivnames = variableModels.iVariables.map(iv => iv.name);

    var variables = this.props.variableModels.iVariables.map(variable => (
      <IndependentVariable iv={variable} />
    ));

    switch (this.state.mode) {
      default:
        return (
          <form
            onSubmit={this.handleSubmit}
            onChange={this.handleChange}
            className="request"
          >
            <table>
              <tbody>
                <tr>
                  <td>
                    What did you find out about whether the Handle Length makes
                    a difference?
                  </td>
                  <td>
                    <textarea id="handlelength"></textarea>
                  </td>
                </tr>
                <tr>
                  <td>What results show you are right?</td>
                  <td>
                    <textarea id="results"></textarea>
                  </td>
                </tr>
              </tbody>
            </table>
            <button type="submit" disabled={!this.isEnabled()}>
              Enter
            </button>
          </form>
        );
    }
  }
}
