import {React} from './deps.js';
import {
  DELAY_PROMPT_TIME_LONG,
  DELAY_PROMPT_TIME_SHORT,
  LONG_PROMPT_SIZE,
  MESSAGE_COUNT_LIMIT,
  REALLYSHORT_PROMPT_SIZE,
  RESPONSE_SYSTEM_GENERATED,
  UI_PROMPT_ENTER_TO_CONTINUE,
  UI_PROMPT_MC,
  UI_PROMPT_STRAIGHT_THROUGH,
  UI_PROMPT_TEXT,
  DELAY_PROMPT_TIME_REALLY_SHORT,
} from './user.js';

var MSG_ROBOT = 'robot';
var MSG_HUMAN = 'student';
var DisplayText = {};
DisplayText[MSG_ROBOT] = 'Researcher';

export class Dialog extends React.Component {
  constructor(props) {
    super(props);
    var state = {mode: 0, UIAction: '', showPhaseLinks: false};
    var user = this.props.user;
    var history = user.getHistory() ? user.getHistory() : {};
    state.isNewUser = history.length == 0;
    state.welcomeText = state.isNewUser
      ? 'Welcome to Astro-world!'
      : "Welcome back! Let's pick up where we left off.";
    this.state = state;
  }

  changeState = () => {
    var app = this.props.app;
    if (this.props.user.getAction().UIActionModeId != this.state.UIAction) {
      this.state.UIAction = this.props.user.getAction().UIActionModeId;
      app.changeState();
    } else {
      this.setState(this.state);
    }
  };
  showAction = () => {
    var app = this.props.app;
    app.showAction();
  };
  togglePhaseLinks(event) {
    if (event) {
      event.preventDefault();
    }

    this.state.showPhaseLinks = !this.state.showPhaseLinks;
    this.setState(this.state);
  }
  gotoPhase(phase) {
    this.state.showPhaseLinks = false;
    var phaseno = '';
    switch (phase) {
      case 'Cov':
        phaseno = 'I';
        break;
      case 'Chart':
        phaseno = 'II';
        break;
      case 'Prediction':
        phaseno = 'III';
        break;
    }

    this.state.welcomeText = 'Welcome to Phase ' + phaseno + '.';
    var self = this;
    var user = this.props.user;
    var onComplete = function () {
      self.setState(self.state);
    };
    var response = {};
    var jsonResponse = JSON.stringify(response);
    user.gotoPhase(phase, jsonResponse, onComplete);
  }
  render() {
    var state = this.state;
    var user = this.props.user;
    var app = this.props.app;
    var history = user.getHistory() ? user.getHistory() : {};
    var oldHistoryLength;
    var oldHistory, newHistory;
    if (user.getArchiveHistoryLength() <= MESSAGE_COUNT_LIMIT) {
      oldHistoryLength = user.getArchiveHistoryLength();
      oldHistory = history.slice(0, oldHistoryLength);
      newHistory = history.slice(oldHistoryLength);
    } else {
      oldHistoryLength =
        user.getArchiveHistoryLength() - history[0].MessageNo + 1;
      oldHistory = user.getHistory()
        ? user.getHistory().slice(0, oldHistoryLength)
        : {};
      newHistory = user.getHistory()
        ? user.getHistory().slice(oldHistoryLength)
        : {};
    }
    var messages = newHistory.map(function (message, i) {
      return (
        <div key={i}>
          <Message
            texts={message.Texts}
            mtype={message.Mtype}
            app={app}
            user={user}
          />
        </div>
      );
    });
    var prompt = user.getPrompt();
    var welcomeText = this.state.welcomeText;

    var self = this;
    var phaseILink = function () {
      self.gotoPhase('Cov');
    };
    var phaseIILink = function () {
      self.gotoPhase('Chart');
    };
    var phaseIIILink = function () {
      self.gotoPhase('Prediction');
    };

    var helpText = this.state.showPhaseLinks ? (
      <div className="help">
        <a onClick={phaseILink}>Phase I</a>{' '}
        <a onClick={phaseIILink}>Phase II</a>{' '}
        <a onClick={phaseIIILink}>Phase III</a>
      </div>
    ) : (
      <div className="help">
        <a onClick={this.togglePhaseLinks}>?</a>
      </div>
    );

    // var helpText = <div className="help"><a href={phaseILink}>Phase I</a>|<a href={phaseIILink}>Phase II</a>|<a href={phaseIIILink}>Phase III</a></div>;

    if (!prompt || Object.keys(prompt).length == 0) {
      return (
        <div>
          {helpText}
          <Title user={user} welcomeText={welcomeText} />
          <OldHistory user={user} oldHistory={oldHistory} />
          {messages}
        </div>
      );
    } else {
      return (
        <div>
          {helpText}
          <Title user={user} welcomeText={welcomeText} />
          <OldHistory user={user} oldHistory={oldHistory} />
          {messages}
          <Prompt
            user={user}
            prompt={prompt}
            onShowInput={this.showAction}
            onComplete={this.changeState}
            app={app}
            key={prompt.PromptId}
          />
        </div>
      );
    }
  }
}

// Render the title of the chat window
class OldHistory extends React.Component {
  state = {showMessages: false};

  changeState = () => {
    this.state.showMessages = !this.state.showMessages;
    this.setState(this.state); // This call triggers re-rendering
  };

  render() {
    var state = this.state;
    var user = this.props.user;
    var oldHistory = this.props.oldHistory ? this.props.oldHistory : {};
    var messages = oldHistory.map(function (message, i) {
      return (
        <div key={i}>
          <Message texts={message.Texts} mtype={message.Mtype} user={user} />
        </div>
      );
    });
    if (messages.length > 0) {
      if (state.showMessages) {
        return (
          <div>
            {messages}
            <button type="submit" onClick={this.changeState}>
              Click to Hide old chat history
            </button>
          </div>
        );
      } else {
        return (
          <div>
            <button type="submit" onClick={this.changeState}>
              Click to show old chat history
            </button>
          </div>
        );
      }
    }
    return <div></div>;
  }
}

function Title(props) {
  var user = props.user;
  var human = user.getScreenname()
    ? props.user.getScreenname()
    : props.user.getUsername();
  var welcomeText = props.welcomeText;
  return (
    <div className="researcher">
      <div className="name">{DisplayText[MSG_ROBOT]}</div>
      <div className="message">
        Hello {human}.<br />
        {welcomeText}
        <br />
      </div>
    </div>
  );
}

// Render each message
class MessageText extends React.Component {
  componentDidMount() {
    // var e = ReactDOM.findDOMNode(this);
    // e.scrollIntoView();
  }

  componentDidUpdate(prevProps, prevState) {
    // var e = ReactDOM.findDOMNode(this);
    // e.scrollIntoView();
  }

  render() {
    var message = this.props.message;
    var human = this.props.user.getScreenname()
      ? this.props.user.getScreenname()
      : this.props.user.getUsername();

    if (message.Text) {
      if (message.Mtype == MSG_ROBOT) {
        return (
          <div className="researcher">
            <div className="name">{DisplayText[MSG_ROBOT]}</div>
            <div className="message">{message.Text}</div>
          </div>
        );
      } else if (message.Mtype == MSG_HUMAN) {
        return (
          <div className="user">
            <div className="name">{human}</div>
            <div className="message">{message.Text}</div>
          </div>
        );
      }
      console.error('Unknown sender!', error);
      return (
        <div className="researcher">
          <div className="message">{this.props.message.Text}</div>
        </div>
      );
    }
    return <div></div>;
  }
}

// Render each message
class Message extends React.Component {
  state = {count: 1, complete: false};

  refreshAfterDelay() {
    var texts = this.props.texts;
    if (this.props.delay && !this.state.complete) {
      if (this.state.count < texts.length) {
        this.triggerDelay();
      }
      if (this.state.count == 2 || this.state.count == texts.length) {
        this.state.complete = true;
        this.setState(this.state);
        // This should only be necessary if delay is turned on
        // otherwise, everything would have been rendered.
        if (this.props.onComplete) {
          this.props.onComplete();
        }
      }
    }
    // // TODO This should not be needed because everything should have been rendered.
    // } else {
    //   this.state.complete = true;
    //   if (this.props.onComplete) {
    //       this.props.onComplete();
    //   }
    // }
  }

  componentDidMount() {
    this.refreshAfterDelay();
  }

  componentDidUpdate(prevProps, prevState) {
    this.refreshAfterDelay();
  }

  triggerDelay() {
    if (this.state.count >= this.props.texts.length) {
      return;
    }
    var d = DELAY_PROMPT_TIME_SHORT;
    if (this.props.texts[this.state.count - 1].length > LONG_PROMPT_SIZE) {
      d = DELAY_PROMPT_TIME_LONG;
    } else if (
      this.props.texts[this.state.count - 1].length < REALLYSHORT_PROMPT_SIZE
    ) {
      d = DELAY_PROMPT_TIME_REALLY_SHORT;
    }
    this.state.interval = window.setInterval(this.unTriggerDelay, d);
  }

  unTriggerDelay = () => {
    window.clearInterval(this.state.interval);
    this.state.count++;
    this.setState(this.state);
  };

  render() {
    var texts = this.props.texts;
    var delay = this.props.delay;
    var mtype = this.props.mtype;
    var user = this.props.user;
    var lastCount;

    if (!delay) {
      lastCount = texts.length;
    } else {
      lastCount = this.state.count;
    }

    var messages = texts.slice(0, lastCount).map(function (text, i) {
      var message = {};
      message.Mtype = mtype;
      message.Text = text;
      return (
        <div key={i}>
          <MessageText message={message} user={user} />
        </div>
      );
    });

    return <div>{messages}</div>;
  }
}

class Prompt extends React.Component {
  state = {completePrompt: false};

  handleChange = event => {
    this.setState({});
  };

  showInput = () => {
    this.state.completePrompt = true;
    this.setState(this.state);
    if (this.props.onShowInput) {
      this.props.onShowInput();
    }
  };

  render() {
    var app = this.props.app;
    var prompt = this.props.prompt;
    var texts = prompt.Texts;
    var user = this.props.user;
    var onComplete = this.props.onComplete;

    var promptId = prompt.PromptId;
    var phaseId = user.CurrentPhaseId;

    var human = this.props.user.getScreenname()
      ? this.props.user.getScreenname()
      : this.props.user.getUsername();

    if (this.state.completePrompt) {
      switch (prompt.PromptType) {
        case UI_PROMPT_ENTER_TO_CONTINUE:
        case UI_PROMPT_TEXT:
        case UI_PROMPT_MC:
        case UI_PROMPT_STRAIGHT_THROUGH:
          return (
            <div key={promptId + user.getHistory().length}>
              <Message texts={texts} mtype={MSG_ROBOT} app={app} user={user} />
              <div className="user">
                <div className="name">{human}</div>
                <Input
                  user={user}
                  prompt={prompt}
                  onComplete={onComplete}
                  app={app}
                />
              </div>
            </div>
          );
        default:
          return (
            <div key={promptId + user.getHistory().length}>
              <Message texts={texts} mtype={MSG_ROBOT} app={app} user={user} />
            </div>
          );
      }
    }
    return (
      <div key={promptId + user.getHistory().length}>
        <Message
          texts={texts}
          delay={true}
          mtype={MSG_ROBOT}
          app={app}
          user={user}
          onComplete={this.showInput}
        />
      </div>
    );
  }
}

function PromptOption(props) {
  var option = props.option;
  return (
    <label>
      <input type="radio" name="dialoginput" value={option.ResponseId} />
      {option.Text}&nbsp;&nbsp;&nbsp;
    </label>
  );
}

class Input extends React.Component {
  state = {enabled: false, passthrough: true};

  componentDidMount() {
    if (this.triggerSubmit()) {
      this.handleSubmit();
    }
  }

  componentDidUpdate(prevProps, prevState) {
    if (this.triggerSubmit()) {
      this.handleSubmit();
    }
  }

  isEnabled() {
    return this.state.enabled;
  }

  handleChange = event => {
    var user = this.props.user;
    var f = document.getElementById('dialogForm');
    var e = f.elements['dialoginput'];
    var value = e ? e.value : '';
    switch (user.CurrentUIPrompt.PromptType) {
      case UI_PROMPT_TEXT:
        if (value.trim() != '') {
          this.setState({enabled: true});
        }
        break;
      default:
        this.setState({enabled: true});
    }
  };

  handleEnter = event => {
    if (this.state.enabled) {
      if (!event.shiftKey) {
        if (event.which == 13) {
          // "Enter" key was pressed.
          this.handleSubmit(event);
        }
      }
    }
  };

  triggerSubmit() {
    if (this.props.prompt.PromptType == UI_PROMPT_STRAIGHT_THROUGH) {
      if (this.state.passthrough) {
        this.state.passthrough = false;
        return true;
      }
    }
    return false;
  }

  handleSubmit = event => {
    if (event) {
      event.preventDefault();
    }

    var user = this.props.user;
    var onComplete = this.props.onComplete;
    var e = document.getElementById('promptId');
    var promptId = e ? e.value : '';
    var e = document.getElementById('phaseId');
    var phaseId = e ? e.value : '';
    var f = document.getElementById('dialogForm');
    e = f.elements['dialoginput'];
    var value = e ? e.value : '';
    e.value = '';
    var text, id;
    var options = user.CurrentUIPrompt.Options;

    switch (user.CurrentUIPrompt.PromptType) {
      case UI_PROMPT_MC:
        for (var i = 0; i < options.length; i++) {
          if (options[i].ResponseId == value) {
            text = options[i].Text;
            id = value;
            break;
          }
        }
        break;
      case UI_PROMPT_TEXT:
        text = value;
        id = value;
        break;
      case UI_PROMPT_ENTER_TO_CONTINUE:
      case UI_PROMPT_STRAIGHT_THROUGH:
        text = RESPONSE_SYSTEM_GENERATED;
        id = RESPONSE_SYSTEM_GENERATED;
        break;
    }

    var response = {};
    response.text = text;
    response.id = id;
    var jsonResponse = JSON.stringify(response);
    user.submitResponse(promptId, phaseId, jsonResponse, onComplete);
  };

  render() {
    var app = this.props.app;
    var prompt = this.props.prompt;
    var texts = prompt.Texts;
    var user = this.props.user;
    var onComplete = this.props.onComplete;

    var promptId = prompt.PromptId;
    var phaseId = user.CurrentPhaseId;

    var human = this.props.user.getScreenname()
      ? this.props.user.getScreenname()
      : this.props.user.getUsername();

    switch (prompt.PromptType) {
      case UI_PROMPT_ENTER_TO_CONTINUE:
        return (
          <div className="form">
            <form
              id="dialogForm"
              onSubmit={this.handleSubmit}
              onChange={this.handleChange}
              className="request"
            >
              <input type="hidden" id="dialoginput" disabled />
              <input type="hidden" id="promptId" value={promptId} />
              <input type="hidden" id="phaseId" value={phaseId} />
              <button type="submit" autoFocus>
                Enter
              </button>
            </form>
          </div>
        );
      case UI_PROMPT_TEXT:
        return (
          <div className="form">
            <form
              id="dialogForm"
              onSubmit={this.handleSubmit}
              onChange={this.handleChange}
              className="request"
            >
              <textarea
                autoFocus
                name="dialoginput"
                onKeyDown={this.handleEnter}
              ></textarea>
              <br />
              <input type="hidden" id="promptId" value={promptId} />
              <input type="hidden" id="phaseId" value={phaseId} />
              <button type="submit" disabled={!this.isEnabled()}>
                Enter
              </button>
            </form>
          </div>
        );
      case UI_PROMPT_MC:
        if (!prompt.Options) {
          console.error('Error: MC Prompt without options!');
          return <div></div>;
        }
        var options = prompt.Options.map(function (option, i) {
          return <PromptOption option={option} key={i} />;
        });

        return (
          <div className="form">
            <form
              id="dialogForm"
              onSubmit={this.handleSubmit}
              onChange={this.handleChange}
              className="request"
            >
              {options}
              <br />
              <input type="hidden" id="promptId" value={promptId} />
              <input type="hidden" id="phaseId" value={phaseId} />
              <button autoFocus type="submit" disabled={!this.isEnabled()}>
                Enter
              </button>
            </form>
          </div>
        );
      case UI_PROMPT_STRAIGHT_THROUGH:
        return (
          <div className="form">
            <form
              id="dialogForm"
              onSubmit={this.handleSubmit}
              className="request"
            >
              <input type="text" name="dialoginput" disabled />
              <br />
              <input type="hidden" id="promptId" value={promptId} />
              <input type="hidden" id="phaseId" value={phaseId} />
              <button
                type="submit"
                id="submitButton"
                disabled={!this.isEnabled()}
              >
                Enter
              </button>
            </form>
          </div>
        );
      default:
        return <div></div>;
    }
  }
}
