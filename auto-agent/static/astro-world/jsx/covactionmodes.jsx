import {React} from './deps.js';
import {
  MultiFactorsCausality,
  MultiFactorsCausalityLevels,
  MemoForm,
} from './action.js';

export class SelectTargetFactor extends React.Component {
  state = {enabled: false};

  isEnabled() {
    return this.state.enabled;
  }

  handleChange = event => {
    this.setState({enabled: true});
  };

  handleSubmit = event => {
    event.preventDefault(); // default might be to follow a link, instead, takes control over the event

    var user = this.props.user;
    var onComplete = this.props.onComplete;
    var e = document.getElementById('promptId');
    var promptId = e ? e.value : '';
    var e = document.getElementById('phaseId');
    var phaseId = e ? e.value : '';
    var f = document.getElementById('covactionForm');
    e = f.elements['covactioninput'];
    var value = e ? e.value : '';
    e.value = '';
    var text, id;

    var options = user.getPrompt().Options;
    for (var i = 0; i < options.length; i++) {
      if (options[i].ResponseId == value) {
        text = options[i].Text;
        id = value;
        break;
      }
    }

    var response = {};
    response.text = text;
    response.id = id;
    var jsonResponse = JSON.stringify(response);
    user.submitResponse(promptId, phaseId, jsonResponse, onComplete);
  };

  render() {
    var user = this.props.user;
    var prompt = user.getPrompt();

    var promptId = prompt.PromptId;
    var phaseId = user.getCurrentPhaseId();

    if (!prompt.Options) {
      console.error('Error: Select factor UI without options!');
      return <div></div>;
    }
    var options = prompt.Options.map(function (option, i) {
      return <FactorPromptOption option={option} key={i} />;
    });

    return (
      <form
        id="covactionForm"
        onSubmit={this.handleSubmit}
        onChange={this.handleChange}
      >
        <div className="hbox">
          <div className="frame">
            <table>
              <tbody>
                <tr>
                  <td className="question">Select the factor to investigate</td>
                </tr>
                <tr>
                  <td>{prompt.Text}</td>
                </tr>
                {options}
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
            key={'SelectTargetFactor'}
          >
            Enter
          </button>
        </p>
      </form>
    );
  }
}

export function FactorPromptOption(props) {
  var option = props.option;
  return (
    <tr>
      <td>
        <label>
          <input type="radio" name="covactioninput" value={option.ResponseId} />
          <br />
          {option.Text}
        </label>
      </td>
    </tr>
  );
}

export function PriorBeliefFactors(props) {
  var question = 'Select "Yes" for factor that you think makes a difference.';
  var formName = 'covactionForm';
  return (
    <MultiFactorsCausality
      formName={formName}
      question={question}
      user={props.user}
      onComplete={props.onComplete}
      app={props.app}
    />
  );
}

export function PriorBeliefLevels(props) {
  var question =
    'Choose the level of the factor that you think would be best for performance.';
  var formName = 'covactionForm';
  return (
    <MultiFactorsCausalityLevels
      formName={formName}
      question={question}
      user={props.user}
      onComplete={props.onComplete}
      app={props.app}
    />
  );
}

export class RecordSelection extends React.Component {
  state = {enabled: false};

  isEnabled() {
    return this.state.enabled;
  }

  // return an array of selected levels for each factor
  // f.FactorId : the id of a factor
  // f.SelectedLevelId: the id of the level selected for the factor
  getSelectedFactors(record) {
    var user = this.props.user;
    var prompt = user.getPrompt();
    var form = document.getElementById('covactionForm');

    var factorOrder = [];
    var tempfactors = Object.keys(user.getContentFactors()).map(function (
      fkey,
      i,
    ) {
      var factor = user.getContentFactors()[fkey];
      factorOrder[i] = factor.Order;
      var fid = form.elements[factor.FactorId + record];
      if (fid) {
        var f = {};
        f.FactorId = factor.FactorId;
        f.SelectedLevelId = fid ? fid.value : '';
        return f;
      }
    });

    var selectedFactors = [];
    for (var i = 0; i < tempfactors.length; i++) {
      selectedFactors[factorOrder[i]] = tempfactors[i];
    }
    return selectedFactors;
  }

  handleChange = event => {
    var doubleRecord = this.props.doubleRecord;
    var comparePrevious = this.props.comparePrevious;

    var selectedFactors;
    if (!comparePrevious) {
      selectedFactors = this.getSelectedFactors('1');
      for (var i = 0; i < selectedFactors.length; i++) {
        if (selectedFactors[i].SelectedLevelId == '') {
          return;
        }
      }
    }
    if (doubleRecord) {
      selectedFactors = this.getSelectedFactors('2');
      for (var i = 0; i < selectedFactors.length; i++) {
        if (selectedFactors[i].SelectedLevelId == '') {
          return;
        }
      }
    }
    this.setState({enabled: true});
  };

  handleSubmit = event => {
    event.preventDefault();

    var user = this.props.user;
    var prompt = user.getPrompt();
    var onComplete = this.props.onComplete;
    var doubleRecord = this.props.doubleRecord;
    var comparePrevious = this.props.comparePrevious;

    var e = document.getElementById('promptId');
    var promptId = e ? e.value : '';
    var e = document.getElementById('phaseId');
    var phaseId = e ? e.value : '';
    var f = document.getElementById('covactionForm');

    var response = {};

    // !doubleRecord && !comparePrevious:
    // Set response.RecordNoOne

    // doubleRecord && !comparePrevious:
    // Set response.RecordNoOne and response.RecordNoTwo

    // doubleRecord && comparePrevious:
    // Set response.UseDBRecordNoOne and response.RecordNoTwo

    // !doubleRecord && comparePrevious:
    // Not used

    var r1selectedFactors, r2selectedFactors;
    response.UseDBRecordNoOne = false;
    response.UseDBRecordNoTwo = false;

    if (!doubleRecord && !comparePrevious) {
      r1selectedFactors = this.getSelectedFactors('1');
    } else if (doubleRecord && !comparePrevious) {
      r1selectedFactors = this.getSelectedFactors('1');
      r2selectedFactors = this.getSelectedFactors('2');
    } else {
      response.UseDBRecordNoOne = true;
      r2selectedFactors = this.getSelectedFactors('2');
    }

    response.RecordNoOne = r1selectedFactors;
    response.RecordNoTwo = r2selectedFactors;

    var jsonResponse = JSON.stringify(response);
    user.submitResponse(promptId, phaseId, jsonResponse, onComplete);
  };

  render() {
    var state = this.state;
    var user = this.props.user;
    var app = this.props.app;
    var doubleRecord = this.props.doubleRecord;
    var comparePrevious = this.props.comparePrevious;
    var prompt = user.getPrompt();
    var factors = user.getContentFactors();

    var promptId = prompt.PromptId;
    var phaseId = user.getCurrentPhaseId();
    var recordOneFactors,
      recordTwoFactors = null;
    var factorOrder = [];

    if (!comparePrevious) {
      var tempfactors = Object.keys(factors).map(function (fkey, i) {
        var factor = factors[fkey];
        factorOrder[i] = factor.Order;
        return <FactorSelection factor={factor} key={i} record="1" />;
      });
      recordOneFactors = [];
      for (var i = 0; i < tempfactors.length; i++) {
        recordOneFactors[factorOrder[i]] = tempfactors[i];
      }
    }
    var tempfactors = Object.keys(factors).map(function (fkey, i) {
      var factor = factors[fkey];
      factorOrder[i] = factor.Order;
      return <FactorSelection factor={factor} key={i} record="2" />;
    });
    recordTwoFactors = [];
    for (var i = 0; i < tempfactors.length; i++) {
      recordTwoFactors[factorOrder[i]] = tempfactors[i];
    }

    if (!doubleRecord) {
      return (
        <form
          id="covactionForm"
          onSubmit={this.handleSubmit}
          onChange={this.handleChange}
        >
          <div className="hbox">
            <div className="frame">
              <table className="record">
                <tbody>
                  <tr>
                    <td colSpan="4" className="recordTitle">
                      First Record
                    </td>
                  </tr>
                  {recordOneFactors}
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
              key={'RecordSelection'}
            >
              Enter
            </button>
          </p>
        </form>
      );
    } else if (!comparePrevious) {
      return (
        <form
          id="covactionForm"
          onSubmit={this.handleSubmit}
          onChange={this.handleChange}
        >
          <div className="hbox">
            <div className="frame">
              <table className="record">
                <tbody>
                  <tr>
                    <td colSpan="4" className="recordTitle">
                      First Record
                    </td>
                  </tr>
                  {recordOneFactors}
                </tbody>
              </table>
            </div>
            <div className="frame">
              <table className="record">
                <tbody>
                  <tr>
                    <td colSpan="4" className="recordTitle">
                      Second Record
                    </td>
                  </tr>
                  {recordTwoFactors}
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
              key={'RecordSelection'}
            >
              Enter
            </button>
          </p>
        </form>
      );
    } else {
      return (
        <div className="hbox">
          <div className="frame">
            <form
              id="covactionForm"
              onSubmit={this.handleSubmit}
              onChange={this.handleChange}
            >
              <table className="record">
                <tbody>
                  <tr>
                    <td colSpan="4" className="recordTitle">
                      Second Record
                    </td>
                  </tr>
                  {recordTwoFactors}
                </tbody>
              </table>
              <p>
                <input type="hidden" id="promptId" value={promptId} />
                <input type="hidden" id="phaseId" value={phaseId} />
                <button
                  type="submit"
                  disabled={!this.isEnabled()}
                  key={'RecordSelection'}
                >
                  Enter
                </button>
              </p>
            </form>
          </div>
          <RecordPerformance user={user} app={app} recordOneOnly />
        </div>
      );
    }
  }
}

function FactorSelection(props) {
  var factor = props.factor;
  var record = props.record;

  var size = factor.Levels.length;
  if (size == 2) {
    factor.Levels[2] = factor.Levels[1];
    factor.Levels[1] = '_';
  }

  var levels = factor.Levels.map(function (level, i) {
    if (level == '_') {
      return <td key={i}>&nbsp;</td>;
    }
    return (
      <FactorLevelSelection
        factor={factor}
        level={level}
        key={i}
        record={record}
        size={size}
      />
    );
  });

  return (
    <tr>
      <td colSpan="3" className="factorLevelRow">
        <label className="factorNameRow">{factor.Text}</label>
        <table style={{width: '100%'}}>
          <tbody>
            <tr>{levels}</tr>
          </tbody>
        </table>
      </td>
    </tr>
  );
}

function FactorLevelSelection(props) {
  var factor = props.factor;
  var record = props.record;
  var level = props.level;
  var imgPath = '/img/' + level.ImgPath;
  var factorId = factor.FactorId + record;

  return (
    <td style={{width: '33%'}}>
      <label>
        <input type="radio" name={factorId} value={level.FactorLevelId} />
        <img src={imgPath} />
        <br />
        {level.Text}
      </label>
    </td>
  );
}

export class RecordPerformance extends React.Component {
  state = {mode: 0};

  render() {
    var state = this.state;
    var user = this.props.user;
    var app = this.props.app;
    var recordOneOnly = this.props.recordOneOnly;
    var recordTwoOnly = this.props.recordTwoOnly;
    var hidePerformance = this.props.hidePerformance;

    var prompt = user.getPrompt();
    var promptId = prompt.PromptId;
    var phaseId = user.getCurrentPhaseId();
    var record1 =
      user.getState().RecordNoOne && user.getState().RecordNoOne.RecordNo
        ? user.getState().RecordNoOne
        : null;
    var record2 =
      user.getState().RecordNoTwo && user.getState().RecordNoTwo.RecordNo
        ? user.getState().RecordNoTwo
        : null;

    var showPerformanceLevels = this.props.showPerformanceLevels ? (
      <div>
        <table className="record">
          <tbody>
            <tr>
              <td colSpan="4">Performance Levels:</td>
            </tr>
            <tr>
              <td>&nbsp;</td>
              <td>A</td>
              <td>-</td>
              <td>very well</td>
            </tr>
            <tr>
              <td>&nbsp;</td>
              <td>B</td>
              <td>-</td>
              <td>well</td>
            </tr>
            <tr>
              <td>&nbsp;</td>
              <td>C</td>
              <td>-</td>
              <td>so so</td>
            </tr>
            <tr>
              <td>&nbsp;</td>
              <td>D</td>
              <td>-</td>
              <td>poorly</td>
            </tr>
            <tr>
              <td>&nbsp;</td>
              <td>E</td>
              <td>-</td>
              <td>very poorly</td>
            </tr>
          </tbody>
        </table>
      </div>
    ) : null;

    var performance = function (r) {
      return !hidePerformance ? (
        <p className="performance-level">
          Performance Level:
          <span className="grade">{r.Performance}</span>
        </p>
      ) : null;
    };

    var recordDetails = function (r) {
      var factorOrder = [];
      var tempfactors = Object.keys(user.getContentFactors()).map(function (
        fkey,
        i,
      ) {
        var factor = user.getContentFactors()[fkey];
        factorOrder[i] = factor.Order;
        var fid = factor.FactorId;
        var selectedf = r.FactorLevels[fid];
        var SelectedLevelName = selectedf.SelectedLevel;

        var size = factor.Levels.length;
        if (size == 2) {
          factor.Levels[2] = factor.Levels[1];
          factor.Levels[1] = '_';
        }
        var levels = factor.Levels.map(function (level, j) {
          if (level.ImgPath) {
            var imgPath = '/img/' + level.ImgPath;
            if (level.Text == SelectedLevelName) {
              return (
                <td key={j}>
                  <label>
                    <img src={imgPath} />
                    <br />
                    {level.Text}
                  </label>
                </td>
              );
            }
          }
          return (
            <td key={j}>
              <label className="dimmed">
                <img src={imgPath} />
                <br />
                {level.Text}
              </label>
            </td>
          );
        });

        return (
          <tr key={i}>
            <td className="factorNameFront">{factor.Text}</td>
            {levels}
          </tr>
        );
      });
      var factors = [];
      for (var i = 0; i < tempfactors.length; i++) {
        factors[factorOrder[i]] = tempfactors[i];
      }

      return r ? (
        <div className="frame" key={r.RecordNo}>
          <table className="record">
            <tbody>
              <tr>
                <td colSpan="4" className="robot">
                  Record #{r.RecordNo} <b>{r.RecordName}</b>
                </td>
              </tr>
              {factors}
            </tbody>
          </table>
          {performance(r)}
        </div>
      ) : null;
    };

    var record1Details, record2Details;
    if (record1 && !recordTwoOnly) {
      record1Details = recordDetails(record1);
    }
    if (record2 && !recordOneOnly) {
      record2Details = recordDetails(record2);
    }
    return (
      <div className="hbox">
        {record1Details}
        {record2Details}
        {showPerformanceLevels}
      </div>
    );
  }
}

export function CovMemoForm(props) {
  var targetFactorName;
  if (props.user.getState().TargetFactor) {
    targetFactorName = props.user.getState().TargetFactor.FactorName;
  }
  var investigatingFactorHeading;
  if (targetFactorName) {
    investigatingFactorHeading = (
      <h3>
        Investigating Factor: <b>{targetFactorName}</b>
      </h3>
    );
  }

  return (
    <div>
      <MemoForm
        user={props.user}
        onComplete={props.onComplete}
        app={props.app}
      />
      <div>
        {investigatingFactorHeading}
        <RecordPerformance user={props.user} app={props.app} />
      </div>
    </div>
  );
}
