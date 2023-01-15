import {React} from './deps.js';
import {PHASE_COV, PHASE_CHART, PHASE_PREDICTION} from './user.js';
import {CovAction} from './covaction.js';
import {ChartAction} from './chartaction.js';
import {PredictionAction} from './predictionaction.js';

export class Action extends React.Component {
  state = {mode: 0};

  changeState = () => {
    // this.setState({mode: 0});
    var app = this.props.app;
    app.changeState();
  };

  render() {
    var user = this.props.user;
    var app = this.props.app;
    switch (user.getCurrentPhaseId()) {
      case PHASE_COV:
        return (
          <div>
            <CovAction user={user} onComplete={this.changeState} app={app} />
          </div>
        );
      case PHASE_CHART:
        return (
          <div>
            <ChartAction user={user} onComplete={this.changeState} app={app} />
          </div>
        );
      case PHASE_PREDICTION:
        return (
          <div>
            <PredictionAction
              user={user}
              onComplete={this.changeState}
              app={app}
            />
          </div>
        );
    }
  }
}

export class MultiFactorsCausality extends React.Component {
  state = {enabled: false};

  isEnabled() {
    return this.state.enabled;
  }

  // return an array of selected levels for each factor
  // f.FactorId : the id of a factor
  // f.SelectedLevelId: the id of the level selected for the factor
  getSelectedFactors() {
    var user = this.props.user;
    var formName = this.props.formName;

    var form = document.getElementById(formName);

    var factorOrder = [];
    var tempfactors = Object.keys(user.getContentFactors()).map((fkey, i) => {
      var factor = user.getContentFactors()[fkey];
      factorOrder[i] = factor.Order;
      var fid = form.elements[factor.FactorId];
      var f = {};
      f.FactorId = factor.FactorId;
      f.IsBeliefCausal = fid.value == 'true' ? true : false;
      return f;
    });

    var selectedFactors = [];
    for (var i = 0; i < tempfactors.length; i++) {
      selectedFactors[factorOrder[i]] = tempfactors[i];
    }
    return selectedFactors;
  }

  handleChange = () => {
    this.setState({enabled: true});
  };

  handleSubmit = event => {
    event.preventDefault();

    var user = this.props.user;

    var onComplete = this.props.onComplete;

    var e = document.getElementById('promptId');
    var promptId = e ? e.value : '';
    e = document.getElementById('phaseId');
    var phaseId = e ? e.value : '';

    var response = {};
    response.BeliefFactors = this.getSelectedFactors();

    var jsonResponse = JSON.stringify(response);
    user.submitResponse(promptId, phaseId, jsonResponse, onComplete);
  };

  render() {
    var user = this.props.user;
    var formName = this.props.formName;
    var question = this.props.question;

    var prompt = user.getPrompt();

    var promptId = prompt.PromptId;
    var phaseId = user.getCurrentPhaseId();

    var factorOrder = [];
    var tempfactors = Object.keys(user.getContentFactors()).map((fkey, i) => {
      var factor = user.getContentFactors()[fkey];
      var factorId = factor.FactorId;
      factorOrder[i] = factor.Order;

      return (
        <tr key={i}>
          <td className="factorNameFront">{factor.Text}</td>
          <td>
            <label>
              <input type="radio" name={factorId} value={true} />
              <br />
              Yes
            </label>
          </td>
          <td>
            <label>
              <input type="radio" name={factorId} value={false} />
              <br />
              No
            </label>
          </td>
        </tr>
      );
    });

    var factors = [];
    for (var i = 0; i < tempfactors.length; i++) {
      factors[factorOrder[i]] = tempfactors[i];
    }

    return (
      <form
        id={formName}
        onSubmit={this.handleSubmit}
        onChange={this.handleChange}
      >
        <div className="hbox">
          <div className="frame">
            <table>
              <tbody>
                <tr>
                  <td colSpan="4" className="question">
                    {question}
                  </td>
                </tr>
                {factors}
              </tbody>
            </table>
          </div>
        </div>
        <p>
          <input type="hidden" id="promptId" value={promptId} />
          <input type="hidden" id="phaseId" value={phaseId} />
          <button
            type="submit"
            disabled={!this.isEnabled()}
            key={'MultiFactorsCausality'}
          >
            Enter
          </button>
        </p>
      </form>
    );
  }
}

export class MultiFactorsCausalityLevelSelection extends React.Component {
  render() {
    var factor = this.props.factor;
    var level = this.props.level;
    var imgPath = '/img/' + level.ImgPath;
    var factorId = factor.FactorId;

    return (
      <td>
        <label>
          <input type="radio" name={factorId} value={level.FactorLevelId} />
          <img src={imgPath} />
          <br />
          {level.Text}
        </label>
      </td>
    );
  }
}

export class MultiFactorsCausalityLevels extends React.Component {
  state = {enabled: false};

  isEnabled() {
    return this.state.enabled;
  }

  // return an array of selected levels for each factor
  // f.FactorId : the id of a factor
  // f.SelectedLevelId: the id of the level selected for the factor
  getSelectedFactors() {
    var user = this.props.user;
    var formName = this.props.formName;
    var form = document.getElementById(formName);

    var factorOrder = [];
    var tempfactors = Object.keys(user.getContentFactors()).map((fkey, i) => {
      var factor = user.getContentFactors()[fkey];
      factorOrder[i] = factor.Order;
      var fid = form.elements[factor.FactorId];
      var f = {};
      f.FactorId = factor.FactorId;
      f.BestLevelId = fid ? fid.value : '';
      f.IsBeliefCausal = factor.IsBeliefCausal;
      return f;
    });

    var selectedFactors = [];
    for (var i = 0; i < tempfactors.length; i++) {
      selectedFactors[factorOrder[i]] = tempfactors[i];
    }
    return selectedFactors;
  }

  handleChange = () => {
    this.setState({enabled: true});
  };

  handleSubmit = event => {
    event.preventDefault();

    var user = this.props.user;
    var onComplete = this.props.onComplete;

    var e = document.getElementById('promptId');
    var promptId = e ? e.value : '';
    e = document.getElementById('phaseId');
    var phaseId = e ? e.value : '';

    var response = {};
    response.BeliefFactors = this.getSelectedFactors();

    var jsonResponse = JSON.stringify(response);
    user.submitResponse(promptId, phaseId, jsonResponse, onComplete);
  };

  render() {
    var user = this.props.user;
    var formName = this.props.formName;
    var question = this.props.question;

    var prompt = user.getPrompt();
    var promptId = prompt.PromptId;
    var phaseId = user.getCurrentPhaseId();

    var factorOrder = [];
    var tempfactors = Object.keys(user.getContentFactors()).map((fkey, i) => {
      var factor = user.getContentFactors()[fkey];
      if (factor.IsBeliefCausal) {
        factorOrder[i] = factor.Order;

        var levels = factor.Levels.map((level, j) => (
          <MultiFactorsCausalityLevelSelection
            factor={factor}
            level={level}
            key={j}
          />
        ));

        return (
          <tr key={i}>
            <td className="factorNameFront">{factor.Text}</td>
            {levels}
          </tr>
        );
      } else {
        return null;
      }
    });
    var factors = [];
    for (var i = 0; i < tempfactors.length; i++) {
      factors[factorOrder[i]] = tempfactors[i];
    }

    return (
      <form
        id={formName}
        onSubmit={this.handleSubmit}
        onChange={this.handleChange}
      >
        <div className="hbox">
          <div className="frame">
            <table>
              <tbody>
                <tr>
                  <td colSpan="4" className="question">
                    {question}
                  </td>
                </tr>
                {factors}
              </tbody>
            </table>
          </div>
        </div>
        <p>
          <input type="hidden" id="promptId" value={promptId} />
          <input type="hidden" id="phaseId" value={phaseId} />
          <button
            type="submit"
            disabled={!this.isEnabled()}
            key={'MultiFactorsCausalityLevels'}
          >
            Enter
          </button>
        </p>
      </form>
    );
  }
}

export class MemoForm extends React.Component {
  state = {enabled: false};

  isEnabled() {
    return this.state.enabled;
  }

  handleChange = () => {
    var form = document.getElementById('actionForm');
    var memo = form.elements['memo'];
    var evidence = form.elements['evidence'];
    if (memo.value && evidence.value) {
      this.setState({enabled: true});
    }
    return;
  };

  handleEnter = event => {
    if (!event.shiftKey) {
      if (event.which == 13) {
        // "Enter" key was pressed.
        this.handleSubmit(event);
      }
    }
  };

  handleSubmit = event => {
    if (event) {
      event.preventDefault();
    }

    var user = this.props.user;
    var targetFactorName, targetFactorId;
    if (user.getState().TargetFactor) {
      targetFactorName = user.getState().TargetFactor.FactorName;
      targetFactorId = user.getState().TargetFactor.FactorId;
    }
    var onComplete = this.props.onComplete;
    var e = document.getElementById('promptId');
    var promptId = e ? e.value : '';
    e = document.getElementById('phaseId');
    var phaseId = e ? e.value : '';
    var form = document.getElementById('actionForm');
    var ask = form.elements['ask'];
    var memo = form.elements['memo'];
    var evidence = form.elements['evidence'];

    var response = {};
    response.Ask = ask ? ask.value : '';
    response.Memo = memo ? memo.value : '';
    response.Evidence = evidence ? evidence.value : '';
    response.Id = targetFactorId;
    response.FactorName = targetFactorName;

    var jsonResponse = JSON.stringify(response);
    user.submitResponse(promptId, phaseId, jsonResponse, onComplete);
  };

  render() {
    var user = this.props.user;
    var prompt = user.getPrompt();

    var promptId = prompt.PromptId;
    var phaseId = user.getCurrentPhaseId();

    var targetFactorName;
    if (user.getState().TargetFactor) {
      targetFactorName = user.getState().TargetFactor.FactorName;
    }

    return (
      <div className="mbox">
        <h3>Memo to the foundation</h3>
        <form
          autoComplete="off"
          id="actionForm"
          onSubmit={this.handleSubmit}
          onChange={this.handleChange}
        >
          <p>
            We recommend that you &nbsp;
            <input
              type="text"
              name="ask"
              size="20"
              autoFocus
              className="con"
              placeholder="Enter ask/do not ask"
            />{' '}
            &nbsp; applicants about <u>{targetFactorName}</u> because &nbsp;
            <input
              type="text"
              name="memo"
              size="50"
              autoFocus
              className="con"
              placeholder="Enter if it does/does not make a difference."
            />
            <br />
            <br />
            Our evidence for claiming this is:
            <br />
            <textarea
              name="evidence"
              className="evid"
              onKeyDown={this.handleEnter}
              placeholder="Enter your answer here"
            ></textarea>
            <br />
          </p>
          <p>
            <input type="hidden" id="promptId" value={promptId} />
            <input type="hidden" id="phaseId" value={phaseId} />
            <button type="submit" disabled={!this.isEnabled()} key={'MemoForm'}>
              Enter
            </button>
          </p>
        </form>
      </div>
    );
  }
}

export function Memo(props) {
  var user = props.user;
  var targetFactorName;
  if (user.getState().TargetFactor) {
    targetFactorName = user.getState().TargetFactor.FactorName;
  }

  var ask, memo, evidence;

  if (user.getState().LastMemo) {
    ask = user.getState().LastMemo.Ask;
    memo = user.getState().LastMemo.Memo;
    evidence = user.getState().LastMemo.Evidence;
  }

  return (
    <div className="mbox">
      <h3>Memo to the foundation</h3>
      <p>
        We recommend that you <u>{ask}</u> applicants about{' '}
        <u>{targetFactorName}</u> because <u>{memo}</u>
        <br />
        <br />
        Our evidence for claiming this is:
        <br />
        <u>{evidence}</u>
        <br />
      </p>
    </div>
  );
}
