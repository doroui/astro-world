import {React} from './deps.js';
import {Dialog} from './dialog.js';
import {UIACTION_INACTIVE, PHASE_CHART, PHASE_PREDICTION} from './user.js';
import {Action} from './action.js';
export class App extends React.Component {
  state = {mode: 0, actionReady: false};

  showAction = () => {
    var self = this;
    var user = this.props.user;
    // In cases when the dialog is ongoing and no UI action is needed
    // No need to re-render the action frame. This allows the last
    // action UI to be present
    var action = user.getAction();
    if (action) {
      if (
        !this.state.actionReady &&
        action.UIActionModeId != UIACTION_INACTIVE
      ) {
        switch (user.getCurrentPhaseId()) {
          case PHASE_CHART:
          case PHASE_PREDICTION:
            if (!user.AllPerformanceRecords) {
              var performanceRecordsPromise = user.loadAllPerformanceRecords();
              performanceRecordsPromise.then(
                function () {
                  self.state.actionReady = true;
                  self.setState(self.state);
                },
                function (error) {
                  console.error('Failed to all performance records!', error);
                },
              );
              return;
            }
        }
        this.state.actionReady = true;
      }
    }
    this.setState(this.state);
  };

  refreshDialog() {
    this.setState(this.state);
  }

  changeState = () => {
    this.setState({mode: 0, actionReady: false});
  };

  render() {
    var state = this.state;
    var user = this.props.user;
    var actionReady = this.state.actionReady;

    if (!actionReady) {
      return (
        <div className="content">
          <div className="dialog">
            <Dialog user={user} app={this} />
          </div>
          <div className="action"></div>
        </div>
      );
    } else {
      return (
        <div className="content">
          <div className="dialog">
            <Dialog user={user} app={this} />
          </div>
          <div className="action">
            <Action user={user} app={this} />
          </div>
        </div>
      );
    }
  }
}
