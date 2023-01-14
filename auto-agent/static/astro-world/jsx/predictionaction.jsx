import {React} from './deps.js';
import {Chart} from './chartactionmodes.js';
import {
  FactorsRequestForm,
  PredictionRecord,
  ContributingFactorsForm,
  SelectTeam,
} from './predictionactionmodes.js';
import {ChartButtons} from './chartactionmodes.js';

export function PredictionAction(props) {
  var user = props.user;
  var app = props.app;
  var prompt = user.getPrompt();
  var action = user.getAction();
  var onComplete = props.onComplete;

  var recordsToShow = [
    {grade: 1, filter: 'fitness:average', no: 64},
    {grade: 4, filter: 'fitness:average', no: 47},
  ];

  if (action) {
    switch (action.UIActionModeId) {
      case 'TARGET_FACTOR_RECORDS':
        return (
          <Chart
            user={user}
            showTargetFactorRecords
            app={app}
            key={'TARGET_FACTOR_RECORDS'}
          />
        );
      case 'FACTORS_REQUEST_FORM':
        return (
          <div>
            <FactorsRequestForm user={user} onComplete={onComplete} app={app} />
            <ChartButtons user={props.user} app={props.app} />
          </div>
        );
      case 'PREDICTION_RECORD':
        return (
          <div>
            <PredictionRecord user={user} onComplete={onComplete} app={app} />
            <ChartButtons user={user} app={app} />
            <PredictionRecord
              user={user}
              onComplete={onComplete}
              app={app}
              predictionHistory
              showPerformancePrediction
            />
          </div>
        );
      case 'PREDICTION_RECORD_SHOW_PREDICTION':
        return (
          <PredictionRecord
            user={user}
            onComplete={onComplete}
            app={app}
            showPerformancePrediction
          />
        );
      case 'CONTRIBUTING_FACTORS_FORM':
        return (
          <div>
            <ContributingFactorsForm
              user={user}
              onComplete={onComplete}
              app={app}
            />
            <ChartButtons user={props.user} app={props.app} />
            <PredictionRecord
              user={props.user}
              onComplete={props.onComplete}
              app={props.app}
            />
          </div>
        );
      case 'SELECT_TEAM_SUMMARY':
        return (
          <SelectTeam user={user} onComplete={onComplete} app={app} isSummary />
        );
      case 'SELECT_TEAM':
        return <SelectTeam user={user} onComplete={onComplete} app={app} />;
      default:
        return <div></div>;
    }
  }
  return <div></div>;
}
