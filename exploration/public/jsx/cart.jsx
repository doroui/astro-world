export function VariableImage(props) {
  return <img src={'images/' + props.name + '.png'} height="100" />;
}

export class IndependentVariable extends React.Component {
  render() {
    var iv = this.props.iv;
    var name = iv.name;
    var handleChange = this.handleChange;
    var options = iv.options.map(option => (
      <IndependentVariableOption name={name} ivOption={option} />
    ));

    return (
      <tr className="iv">
        <td>
          <VariableImage name={iv.name} />
        </td>
        <td>{iv.label}</td>
        <td>{options}</td>
      </tr>
    );
  }
}

// this.props.name is the name of the IndependentVariable this option
//     is associated with.
// ivOption.value is the value that gets saved when the option is selected.
export class IndependentVariableOption extends React.Component {
  render() {
    var ivOption = this.props.ivOption;
    return (
      <label>
        <input type="radio" name={this.props.name} value={ivOption.value} />
        {ivOption.label}
      </label>
    );
  }
}

// Request renders and submit the Variable-level-selection form where user select a level
//     for each IndependentVariable.
// Request has a list of properties defined as attributes in the Request components.
// (For example, see <Result variableModels={variableModels} data={state.newResult}/> in app.jsx)
// They are stored as properties in this.props
// - variableModels
// - data,
// - onComplete, call back function provided by callers of Request
export class Request extends React.Component {
  state = {};

  // @param {Event} e The event within the Variable-level-selection form,
  //     for now, they are only events from an IndependentVariableOption
  // The hashtable is to keep track of the IndependentVariable that has its option selected:
  //     e.target.name is the name of the IndependentVariable
  //     e.target.value is the value of the IndependentVariableOption
  handleChange = e => {
    var state = {};
    state[e.target.name] = e.target.value;
    this.setState(state);
  };

  handleSubmit = e => {
    e.preventDefault();
    this.post(this.state);
  };

  post(data) {
    if (!this.isEnabled()) return;

    var xhr = new XMLHttpRequest();
    xhr.onload = () => {
      if (this.props.onComplete) {
        this.props.onComplete(JSON.parse(xhr.responseText));
      }
    };
    xhr.open('POST', '/carts/gettrips');
    xhr.setRequestHeader('Content-Type', 'application/json');
    xhr.send(JSON.stringify(data));
  }

  // Check all IndependentVariable, only allow form submit if all variables
  //     have their options selected.
  // (See handleChange for keeping track of the variables with options selected)
  isEnabled() {
    var variables = this.props.variableModels.iVariables;
    for (var i = 0; i < variables.length; i++) {
      if (!(variables[i].name in this.state)) {
        return false;
      }
    }
    return true;
  }

  render() {
    var variables = this.props.variableModels.iVariables.map(variable => (
      <IndependentVariable iv={variable} />
    ));

    return (
      <form
        onSubmit={this.handleSubmit}
        onChange={this.handleChange}
        className="request"
      >
        <table>
          <tbody>{variables}</tbody>
        </table>
        <button type="submit" disabled={!this.isEnabled()}>
          See Results
        </button>
      </form>
    );
  }
}
// End --- Request

// Result renders the outcome screen based on the return results from the backend
//     given the levels of the IndependentVariables.
// It displays the outcome and the list of IndependentVariables and their levels
// dvValues is the outcome variable (aka dependent variable)
export class Result extends React.Component {
  render() {
    var variableModels = this.props.variableModels;
    var data = this.props.data;
    var dvValues = data[variableModels.dvName].join(', ');

    var variables = variableModels.iVariables.map(variable => (
      <ResultSelection iv={variable} value={data[variable.name]} />
    ));

    return (
      <table className="result">
        <tbody>
          <tr>
            <td></td>
            <td>{variableModels.dvLabel}:</td>
            <td>{dvValues}</td>
          </tr>
          {variables}
        </tbody>
      </table>
    );
  }
}

// ResultSelection renders one independent variable with its selected level
export class ResultSelection extends React.Component {
  getDisplayValue(value) {
    var options = this.props.iv.options;
    for (var i = 0; i < options.length; i++) {
      if (options[i].value == value) {
        return options[i].label;
      }
    }
    return null;
  }

  render() {
    var iv = this.props.iv;
    var ivValue = this.getDisplayValue(this.props.value);
    return (
      <tr>
        <td>
          <VariableImage name={iv.name} />
        </td>
        <td>{iv.label}:</td>
        <td>{ivValue}</td>
      </tr>
    );
  }
}
