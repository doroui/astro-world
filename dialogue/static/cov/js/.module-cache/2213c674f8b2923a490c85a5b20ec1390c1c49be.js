/** @jsx React.DOM */

// npm install -g react-tools
// jsx -w -x jsx public/js public/js

var MSG_ROBOT = 'robot';
var MSG_HUMAN = 'student';
var DisplayText = {};
DisplayText[MSG_ROBOT] = 'Researcher';

// Render the title of the chat window
var Title = React.createClass({displayName: 'Title',
  render: function() {
    var user = this.props.user;

    return  React.DOM.div(null, "Researcher:",React.DOM.br(null),
              " Hello ", user.getUsername(),React.DOM.br(null),
              " Welcome to the Challenge");
  }
});


// Render each message
var Message = React.createClass({displayName: 'Message',
  render: function() {
    var message = this.props.message;
    var human = this.props.human;

    if (message.type == MSG_ROBOT) {
      return React.DOM.div(null, DisplayText[MSG_ROBOT], " : ", message.text);
    } else if (message.type == MSG_HUMAN) {
      return React.DOM.div(null, human, " : ", message.text);
    }
    console.error("Unknown sender!", error);
    return React.DOM.div(null, this.props.message.text);
  }
});

// Renter input window
var Input = React.createClass({displayName: 'Input',

  getInitialState: function() {
    this.setState({enabled:false});
    return {};
  },

  isEnabled: function() {
    return this.state.enabled;
  },

  handleChange: function(event) {
    this.setState({enabled:true});
  },

  handleSubmit: function(event) {
    event.preventDefault();

    var user = this.props.user;
    var onComplete = this.props.onComplete;
    //TODO need to send workflowstate id also
    user.submitTextResponse(document.getElementById("workflowStateID").value, document.getElementById("input").value, onComplete)
  },

  render: function() {
    var user = this.props.user;
    var workflowStateID = user.getWorkflowStateID();

    return  React.DOM.form( {onSubmit:this.handleSubmit, onChange:this.handleChange,
            className:"request"}, 
              React.DOM.textarea( {id:"input"}),
              React.DOM.input( {type:"hidden", name:"workflowStateID", value:"{workflowStateID}"}),
              React.DOM.button( {type:"submit", disabled:!this.isEnabled()}, "Enter")
            );
  }
});











var App = React.createClass({displayName: 'App',
  getInitialState: function() {
    return {mode: 0};
  },

  // continueFrom: function(e) {
  //   switch (this.state.mode) {
  //     case 1:
  //       this.setState({mode: 2});
  //       break;
  //     case 3:
  //       this.setState({mode: 2});
  //       break;
  //     case 4:
  //       this.setState({mode: 2});
  //       break;
  //     case 5:
  //       this.setState({mode: 2});
  //       break;
  //   }
  // },

  // showAllResultsForChallenge: function(e) {
  //   switch (this.state.mode) {
  //     case 5:
  //       this.setState({mode: 6});
  //       break;
  //   }
  // },

  // hideAllResultsForChallenge: function(e) {
  //   switch (this.state.mode) {
  //     case 6:
  //       this.setState({mode: 5});
  //       break;
  //   }
  // },

  // showChallenge: function(e) {
  //   var self = this;
  //   var renderCallback = function() {
  //     switch (self.state.mode) {
  //       case 1:
  //       case 3:
  //       case 4:
  //         self.setState({mode: 5});
  //         break;
  //     }
  //   };
  //   this.props.user.enterChallenge(renderCallback);
  // },

  // saveResult: function(e) {
  //   var self = this;
  //   this.props.user.addResult(this.state.newResult, function() {
  //     switch (self.state.mode) {
  //       case 1:
  //         self.setState({mode: 4});
  //         break;
  //       case 3:
  //         self.setState({mode: 4});
  //         break;
  //     }
  //   });
  // },

  changeState: function() {
    var state = this.state;
    // switch (state.mode) {
    //   case 0:
    //     this.setState({mode: 1, newResult: data});
    //     break;
    //   case 2:
    //     this.setState({mode: 3, oldResult: state.newResult, newResult: data});
    //     break;
    // }
  },

  render: function() {
    var state = this.state;
    var user = this.props.user;
    var messages = user.getHistory().map(
        function(message) {
          return Message( {message:message, human:user.getUsername()});
        });

    switch (this.state.mode) {
      case 0: //Show User Name
        return React.DOM.div( {className:"app"}, 
          Title( {user:user}),
          messages,
          Input( {user:user, onComplete:this.changeState})
        );
      // case 0: //Show Initial Request Form
      //   return <div className="app single-column">
      //     <Request variableModels={variableModels} onComplete={this.handleComplete}
      //         style={{width: '100%'}}/>
      //   </div>;
    
      // case 1: //Show One Case Results
      //   return <div className="app single-column">
      //     <Result variableModels={variableModels} data={state.newResult}/>
      //     <button onClick={this.continueFrom}>Go to Next Case</button>
      //     <button onClick={this.saveResult}>Save Result to Notebook</button>
      //     <button onClick={this.showChallenge}>Show Challenge</button>
      //   </div>;
    
      // case 2: //Show Request Form With Last Result
      //   return <table className="app"><tbody>
      //     <tr>
      //       <td>New Case:</td>
      //       <td>Last Case:</td>
      //     </tr>
      //     <tr>
      //       <td><Request variableModels={variableModels} onComplete={this.handleComplete}/></td>
      //       <td><Result variableModels={variableModels} data={state.newResult}/></td>
      //     </tr>
      //   </tbody></table>;
    
      // case 3: //Show Two Cases Results
      //   return <table className="app"><tbody>
      //     <tr>
      //       <td>New Case:</td>
      //       <td>Last Case:</td>
      //     </tr>
      //     <tr>
      //       <td><Result variableModels={variableModels} data={state.newResult}/></td>
      //       <td><Result variableModels={variableModels} data={state.oldResult}/></td>
      //     </tr>
      //     <tr>
      //       <td colSpan="2" style={{textAlign: 'center'}}>
      //         <button onClick={this.continueFrom}>Go to Next Case</button>
      //         <button onClick={this.saveResult}>Save Result to Notebook</button>
      //         <button onClick={this.showChallenge}>Show Challenge</button>
      //       </td>
      //     </tr>
      //   </tbody></table>;

      // case 4: //Show Notebook
      //   return <div className="app single-column">
      //     <UserResultData variableModels={this.props.variableModels} user={this.props.user} mode={'notebook'}/>
      //     <button onClick={this.continueFrom}>Go to Next Case</button>
      //     <button onClick={this.showChallenge}>Show Challenge</button>
      //   </div>;

      // case 5: //Show Challenge
      //   return <div className="app single-column">
      //     <table className="app"><tbody>
      //       <tr>
      //         <td><div className="app single-column">
      //           <Challenge variableModels={this.props.variableModels} user={this.props.user}/>
      //           <button onClick={this.showAllResultsForChallenge}>Show Notebook</button>
      //         </div></td>
      //       </tr>
      //   </tbody></table>
      //   </div>;

      // case 6: //Show Challenge with Notebook
      //   return <div className="app single-column">
      //     <table className="app"><tbody>
      //       <tr>
      //         <td><div className="app single-column">
      //           <Challenge variableModels={this.props.variableModels} user={this.props.user}/>
      //           <button onClick={this.hideAllResultsForChallenge}>Hide Notebook</button>
      //         </div></td>
      //         <td><div className="app single-column">
      //           <UserResultData variableModels={this.props.variableModels} user={this.props.user} mode={'notebook'}/>
      //         </div></td>
      //       </tr>
      //   </tbody></table>
      //   </div>;

    }

    throw new Error('Unexpected mode');
  }
});
