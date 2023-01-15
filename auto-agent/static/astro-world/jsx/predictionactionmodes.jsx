import {React} from './deps.js';

export function FactorsRequestForm(props) {
  var question =
    'Check the box for up to four factors that you would like to know about an applicant.';
  var formName = 'predictionactionForm';
  return (
    <MultipleFactorsSelect
      formName={formName}
      question={question}
      user={props.user}
      onComplete={props.onComplete}
      app={props.app}
    />
  );
}

export function ContributingFactorsForm(props) {
  var question =
    'Which of the four factors you have data on mattered to your prediction?';
  var formName = 'predictionactionForm';
  var factors = props.user.getState().DisplayFactors;
  return (
    <MultipleFactorsSelect
      formName={formName}
      factors={factors}
      question={question}
      user={props.user}
      onComplete={props.onComplete}
      app={props.app}
    />
  );
}

export class MultipleFactorsSelect extends React.Component {
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
    var tempfactorIdsMap;

    if (this.props.factors) {
      tempfactorIdsMap = this.props.factors.map(v => v.FactorId);
    } else {
      tempfactorIdsMap = Object.keys(user.getContentFactors());
    }

    var tempfactors = tempfactorIdsMap.map((fkey, i) => {
      var factor = user.getContentFactors()[fkey];
      factorOrder[i] = factor.Order;
      var fid = form.elements[factor.FactorId];
      if (fid) {
        var f = {};
        f.FactorId = factor.FactorId;
        f.IsBeliefCausal = fid.checked;
        return f;
      }
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

    var count = 0;
    for (var i = 0; i < response.BeliefFactors.length; i++) {
      if (response.BeliefFactors[i]) {
        if (response.BeliefFactors[i].IsBeliefCausal) {
          count++;
        }
      }
    }

    if (count > 4) {
      alert(
        'You have selected more than 4 factors. Please remove at least 1 and try again.',
      );
    } else {
      var jsonResponse = JSON.stringify(response);
      user.submitResponse(promptId, phaseId, jsonResponse, onComplete);
    }
  };

  render() {
    var user = this.props.user;
    var formName = this.props.formName;
    var question = this.props.question;

    var prompt = user.getPrompt();

    var promptId = prompt.PromptId;
    var phaseId = user.getCurrentPhaseId();

    var factorOrder = [];
    var tempfactorIdsMap;

    if (this.props.factors) {
      tempfactorIdsMap = this.props.factors.map(v => v.FactorId);
    } else {
      tempfactorIdsMap = Object.keys(user.getContentFactors());
    }

    var tempfactors = tempfactorIdsMap.map((fkey, i) => {
      var factor = user.getContentFactors()[fkey];
      var factorId = factor.FactorId;
      factorOrder[i] = factor.Order;

      return (
        <tr key={i}>
          <td>
            <label>
              <input type="checkbox" name={factorId} />
            </label>
          </td>
          <td className="factorNameFront">{factor.Text}</td>
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
            key={'MultipleFactorsSelect'}
          >
            Enter
          </button>
        </p>
      </form>
    );
  }
}

export class SelectTeam extends React.Component {
  state = {enabled: false, showRecord: false, record: null};

  isEnabled() {
    return this.state.enabled;
  }

  handleChange = () => {
    this.setState({enabled: true});
  };

  showRecord(record) {
    this.state.showRecord = true;
    this.state.record = record;
    this.setState(this.state);
  }

  hideRecord() {
    this.state.showRecord = false;
    this.state.record = null;
    this.setState(this.state);
  }

  getSelectedRecords() {
    var user = this.props.user;
    var formName = 'predictionactionForm';

    var form = document.getElementById(formName);

    var records = user.getState().AllPredictionRecords.map(v => {
      var r = form.elements[v.RecordName];
      if (r) {
        var f = {};
        f.RecordNo = v.RecordNo;
        f.IsSelected = r.checked;
        return f;
      }
    });

    return records;
  }

  handleSubmit = event => {
    event.preventDefault();
    var user = this.props.user;

    var onComplete = this.props.onComplete;

    var e = document.getElementById('promptId');
    var promptId = e ? e.value : '';
    e = document.getElementById('phaseId');
    var phaseId = e ? e.value : '';

    var response = {};
    response.Predictions = this.getSelectedRecords();

    var count = 0;
    for (var i = 0; i < response.Predictions.length; i++) {
      if (response.Predictions[i]) {
        if (response.Predictions[i].IsSelected) {
          count++;
        }
      }
    }

    if (count < 5 || count > 5) {
      alert('You should select exactly 5 applicants. Please try again.');
    } else {
      var jsonResponse = JSON.stringify(response);
      user.submitResponse(promptId, phaseId, jsonResponse, onComplete);
    }
  };

  render() {
    var user = this.props.user;
    var isSummary = this.props.isSummary;

    var prompt = user.getPrompt();
    var question =
      'Check the box for up to five applicants that you would like to be in your team.';
    var formName = 'predictionactionForm';

    var promptId = prompt.PromptId;
    var phaseId = user.getCurrentPhaseId();

    var applicants = user.getState().AllPredictionRecords.map((record, i) => {
      var recordOnClick = () => {
        this.showRecord(record);
      };
      if (isSummary) {
        var isSelectedRecord = record.IsSelected ? (
          <label>
            <input type="checkbox" checked />
          </label>
        ) : (
          <label>&nbsp;</label>
        );
        return (
          <tr key={i}>
            <td>
              <label>{isSelectedRecord}</label>
            </td>
            <td className="factorNameFront"># {record.RecordNo}</td>
            <td className="factorNameFront">
              <button type="button" onClick={recordOnClick}>
                {record.RecordName}
              </button>
            </td>
            <td className="factorNameFront">{record.PredictedPerformance}</td>
          </tr>
        );
      }
      return (
        <tr key={i}>
          <td>
            <label>
              <input type="checkbox" name={record.RecordName} />
            </label>
          </td>
          <td className="factorNameFront"># {record.RecordNo}</td>
          <td className="factorNameFront">
            <button type="button" onClick={recordOnClick}>
              {record.RecordName}
            </button>
          </td>
          <td className="factorNameFront">{record.PredictedPerformance}</td>
        </tr>
      );
    });

    var recordDetails = this.state.showRecord ? (
      <div className="no-border-frame">
        <div className="hbox">
          <PredictionRecord
            user={user}
            record={this.state.record}
            showPerformancePrediction
            key={this.state.record.RecordNo}
          />
        </div>
        <button autoFocus onClick={this.hideRecord}>
          Hide Record
        </button>
      </div>
    ) : null;
    var submitButton = isSummary ? null : (
      <button
        type="submit"
        disabled={!this.isEnabled()}
        key={'MultipleFactorsSelect'}
      >
        Enter
      </button>
    );

    return (
      <div>
        <form
          id={formName}
          onSubmit={this.handleSubmit}
          onChange={this.handleChange}
        >
          <div className="hbox">
            <div className="frame">
              <table className="prediction-team">
                <tbody>
                  <tr>
                    <td colSpan="4" className="question">
                      {question}
                    </td>
                  </tr>
                  <tr>
                    <td>&nbsp;</td>
                    <td className="factorNameFront">Record Number</td>
                    <td className="factorNameFront">Applicant&apos;s Name</td>
                    <td className="factorNameFront">
                      Performance You Predicted
                    </td>
                  </tr>
                  {applicants}
                </tbody>
              </table>
            </div>
          </div>
          <p>
            <input type="hidden" id="promptId" value={promptId} />
            <input type="hidden" id="phaseId" value={phaseId} />
            {submitButton}
          </p>
        </form>
        {recordDetails}
      </div>
    );
  }
}

export class PredictionRecord extends React.Component {
  state = {mode: 0};

  render() {
    var user = this.props.user;
    var record = this.props.record;
    var showPerformancePrediction = this.props.showPerformancePrediction;
    var predictionHistory = this.props.predictionHistory;

    record = record ? record : user.getState().TargetPrediction;

    var recordDetails = r => {
      var performancePrediction = rr =>
        showPerformancePrediction ? (
          <p className="predicted-performance-level">
            You predicted {rr.RecordName}&apos;s performance to be:
            <span className="grade">{rr.PredictedPerformance}</span>
          </p>
        ) : null;

      var factorOrder = [];
      var tempfactors = user.getState().DisplayFactors.map((v, i) => {
        var factor = v;
        factorOrder[i] = factor.Order;
        var fid = factor.FactorId;
        var selectedf = r.FactorLevels[fid];
        var SelectedLevelName = selectedf.SelectedLevel;

        var size = factor.Levels.length;
        if (size == 2) {
          factor.Levels[2] = factor.Levels[1];
          factor.Levels[1] = '_';
        }
        var levels = factor.Levels.map((level, j) => {
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
        <div className="hbox">
          <div className="frame" key={r.RecordNo}>
            <table className="record">
              <tbody>
                <tr>
                  <td colSpan="4" className="robot">
                    Applicant #{r.RecordNo} <b>{r.RecordName}</b>
                  </td>
                </tr>
                {factors}
              </tbody>
            </table>
            {performancePrediction(r)}
          </div>
        </div>
      ) : null;
    };

    var targetRecordDetails = recordDetails(record);
    var hasHistory = false;
    var allPreviousRecordsDetails = user
      .getState()
      .AllPredictionRecords.map(v => {
        if (v.RecordNo < record.RecordNo) {
          hasHistory = true;
          return recordDetails(v);
        }
        return null;
      });
    // Show previous records in reverse order
    var reversePreviousRecordsDetails = allPreviousRecordsDetails.map(
      (v, i) =>
        allPreviousRecordsDetails[allPreviousRecordsDetails.length - i - 1],
    );
    var title = hasHistory ? (
      <h3 className="recordHeading">Your Prediction History</h3>
    ) : null;

    if (predictionHistory) {
      return (
        <div>
          {title}
          {reversePreviousRecordsDetails}
        </div>
      );
    } else {
      return <div>{targetRecordDetails}</div>;
    }
  }
}
