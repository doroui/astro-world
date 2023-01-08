import {React} from './deps.js';

import {
  SelectTargetFactor,
  PriorBeliefFactors,
  PriorBeliefLevels,
  RecordPerformance,
  RecordSelection,
} from './covactionmodes.js';

export class CovAction extends React.Component {
  state = {mode: 0};

  render() {
    var state = this.state;
    var user = this.props.user;
    var app = this.props.app;
    var prompt = user.getPrompt();
    var action = user.getAction();
    var onComplete = this.props.onComplete;

    var investigatingFactorHeading;
    if (user.getState().TargetFactor) {
      investigatingFactorHeading =
        user.getState().TargetFactor.FactorId != '' ? (
          <h3 className="recordHeading">
            Investigating Factor:{' '}
            <b>{user.getState().TargetFactor.FactorName}</b>
          </h3>
        ) : null;
    }

    if (action) {
      switch (action.UIActionModeId) {
        case 'NEW_TARGET_FACTOR':
          return (
            <SelectTargetFactor user={user} onComplete={onComplete} app={app} />
          );
        case 'PRIOR_BELIEF_FACTORS':
          return (
            <PriorBeliefFactors user={user} onComplete={onComplete} app={app} />
          );
        case 'PRIOR_BELIEF_LEVELS':
          return (
            <PriorBeliefLevels user={user} onComplete={onComplete} app={app} />
          );
        case 'RECORD_PERFORMANCE_TEST':
          return (
            <div>
              <RecordPerformance user={user} app={app} showPerformanceLevels />
            </div>
          );
        case 'RECORD_SELECT_ONE_TEST':
          return (
            <div>
              <RecordSelection user={user} onComplete={onComplete} app={app} />
            </div>
          );
        case 'RECORD_SELECT_ONE':
          return (
            <div>
              {investigatingFactorHeading}
              <RecordSelection user={user} onComplete={onComplete} app={app} />
            </div>
          );
        case 'RECORD_SELECT_TWO':
          return (
            <div>
              {investigatingFactorHeading}
              <RecordSelection
                user={user}
                onComplete={onComplete}
                app={app}
                doubleRecord
              />
            </div>
          );
        case 'RECORD_NO_PERFORMANCE':
          return (
            <div>
              {investigatingFactorHeading}
              <RecordPerformance
                user={user}
                onComplete={onComplete}
                app={app}
                hidePerformance
              />
            </div>
          );
        case 'RECORD_SELECT_ONE_AND_SHOW_PERFORMANCE':
          return (
            <div>
              {investigatingFactorHeading}
              <RecordSelection
                user={user}
                onComplete={onComplete}
                app={app}
                doubleRecord
                comparePrevious
              />
            </div>
          );
        case 'RECORD_ONE_PERFORMANCE':
          return (
            <div>
              {investigatingFactorHeading}
              <RecordPerformance user={user} app={app} recordOneOnly />
            </div>
          );
        case 'RECORD_PERFORMANCE':
          return (
            <div>
              {investigatingFactorHeading}
              <RecordPerformance user={user} app={app} />
            </div>
          );
        case 'MEMO_FORM':
          return <CovMemoForm user={user} onComplete={onComplete} app={app} />;
        case 'MEMO':
          return <Memo user={user} app={app} />;
        default:
          return <div></div>;
      }
    }
    return <div></div>;
  }
}
