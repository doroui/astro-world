import {React} from './deps.js';
import {
  MultiFactorsCausality,
  MultiFactorsCausalityLevels,
  MemoForm,
} from './action.js';

const rectSize = 8;
const toolBoxSizeHeight = 200;
const toolBoxSizeWidth = 250;
const spacingFactor = 2.2;
const rowHeight = 100;
const elementsPerRow = 5;
const columnWidth = 225;
const columnLabelHeight = 20;
const paddingLeft = 50;
const paddingBottom = 50;
const paddingTop = 20;
const paddingRight = 20;
const rowColors = ['transparent', 'rgba(0,0,0,0.05)'];
const noFilterTitle = 'All';

// Key for a list of performance records
// This must match what are passed through user.getAllPerformanceRecords()
// Example format of key
// fitness:average
const noFilterKey = 'all'; // Key for all records

// Example format of data
// const data = [{
//   label: 'Average',
//   rcount: [5, 7, 3, 28, 1]
// }, {
//   label: 'Good',
//   rcount: [15, 2, 13, 8, 11]
// }];

export function Chart(props) {
  var user = props.user;
  var app = props.app;

  var recordsToShow = props.recordsToShow;

  var xTitle = '';
  var yLabels,
    xLabels = [];
  var performanceLabels = [
    {
      grade: 'A',
      label: 'A (very well)',
    },
    {
      grade: 'B',
      label: 'B (well)',
    },
    {
      grade: 'C',
      label: 'C (so so)',
    },
    {
      grade: 'D',
      label: 'D (poorly)',
    },
    {
      grade: 'E',
      label: 'E (very poorly)',
    },
  ];
  var data = [];
  var pRecords = user.getAllPerformanceRecords();

  var colFilters = [];
  if (props.filterRecords && props.filterRecords.length > 0) {
    colFilters = props.filterRecords;
    xTitle = props.filterFactorName;
    xLabels = props.filterLevelsLabels;
  } else if (props.showTargetFactorRecords) {
    var factors = user.getContentFactors();
    var targetFactorId;
    if (user.getState().TargetFactor) {
      targetFactorId = user.getState().TargetFactor.FactorId;
      var fkey = Object.keys(factors);
      for (var i = 0; i < fkey.length; i++) {
        if (factors[fkey[i]].FactorId == targetFactorId) {
          for (
            var colIndex = 0;
            colIndex < factors[fkey[i]].Levels.length;
            colIndex++
          ) {
            colFilters[colIndex] =
              targetFactorId +
              ':' +
              factors[fkey[i]].Levels[colIndex].FactorLevelId;
            xLabels[colIndex] = factors[fkey[i]].Levels[colIndex].Text;
          }
          xTitle = user.getState().TargetFactor.FactorName;
          break;
        }
      }
    }
  } else {
    colFilters = ['all'];
    xTitle = noFilterTitle;
  }

  for (var colIndex = 0; colIndex < colFilters.length; colIndex++) {
    data[colIndex] = {label: '', rcount: []};
    data[colIndex].label = xLabels[colIndex];
    for (var gindex = 0; gindex < pRecords.length; gindex++) {
      if (pRecords[gindex].Records[colFilters[colIndex]]) {
        data[colIndex].rcount[gindex] =
          pRecords[gindex].Records[colFilters[colIndex]].length;
      } else {
        data[colIndex].rcount[gindex] = 0;
      }
      // pRecords[gindex].Grade should be the same as performanceLabels[gindex].grade
    }
  }
  return (
    <div>
      <Graph
        singleColumn={props.singleColumn}
        user={user}
        app={app}
        colFilters={colFilters}
        data={data}
        allowToolboxToggle={props.allowToolbox}
        yTitle="Performance"
        xTitle={xTitle}
        yLabels={performanceLabels}
        recordsToShow={recordsToShow}
      />
    </div>
  );
}

function Diamond(props) {
  var h = rectSize / 2;

  var toggleToolbox = function () {
    props.onDiamondClick(props.col, props.grade, props.rIndex, true);
  };

  if (props.allowToolboxToggle) {
    return (
      <rect
        onClick={toggleToolbox}
        width={rectSize}
        height={rectSize}
        transform={`translate(${props.x},${props.y}) rotate(45) translate(-${h},-${h})`}
        style={{stroke: 'green', fill: 'green'}}
      />
    );
  }
  return (
    <rect
      width={rectSize}
      height={rectSize}
      transform={`translate(${props.x},${props.y}) rotate(45) translate(-${h},-${h})`}
      style={{stroke: 'green', fill: 'green'}}
    />
  );
}

function Diamonds(props) {
  var size = rectSize * spacingFactor;
  var diamonds = [];
  var ePerRow = elementsPerRow;
  if (props.singleColumn) {
    ePerRow = elementsPerRow * 2;
  }
  for (let i = 0; i < props.count; i++) {
    let y = Math.floor(i / ePerRow);
    let x = i % ePerRow;
    diamonds.push(
      <Diamond
        x={x * size}
        y={y * size}
        allowToolboxToggle={props.allowToolboxToggle}
        onDiamondClick={props.onDiamondClick}
        col={props.col}
        grade={props.grade}
        rIndex={i}
        key={i}
      />,
    );
  }
  return (
    <g transform={`translate(${size / 2}, -${size / 2}) scale(1, -1)`}>
      {diamonds}
    </g>
  );
}

function Column(props) {
  var totalHeight = props.rcount.length * rowHeight;
  var colWidth = columnWidth;
  var ePerRow = elementsPerRow;
  if (props.singleColumn) {
    colWidth = columnWidth * 2;
    ePerRow = elementsPerRow * 2;
  }
  const x = (colWidth - ePerRow * rectSize * spacingFactor) / 2;
  return (
    <g transform={`translate(${x},${0})`}>
      {props.rcount.map((count, i) => (
        <g transform={`translate(0,${(i + 1) * rowHeight})`} key={i}>
          <Diamonds
            singleColumn={props.singleColumn}
            allowToolboxToggle={props.allowToolboxToggle}
            onDiamondClick={props.onDiamondClick}
            count={props.rcount[i]}
            col={props.col}
            grade={i}
          />
        </g>
      ))}
    </g>
  );
}

function Toolbox(props) {
  var rcount = props.data[props.toolboxCol].rcount;
  var record = props.record;
  var totalHeight = rcount.length * rowHeight;
  var colWidth = columnWidth;
  var ePerRow = elementsPerRow;
  if (props.singleColumn) {
    colWidth = columnWidth * 2;
    ePerRow = elementsPerRow * 2;
  }
  var colX = (colWidth - ePerRow * rectSize * spacingFactor) / 2;

  var size = rectSize * spacingFactor;
  var x = (props.toolboxIndex % ePerRow) * size;
  var y = Math.floor(props.toolboxIndex / ePerRow) * size;
  var h = rectSize / 2;

  var factors = [];
  Object.keys(record.FactorLevels).map(function (l, j) {
    var i =
      Object.keys(record.FactorLevels).length -
      1 -
      record.FactorLevels[l].Order;
    factors[record.FactorLevels[l].Order] = (
      <text
        x={-toolBoxSizeWidth / 3}
        y={-columnLabelHeight * (i + 1)}
        textAnchor="start"
        className="axis-label"
        key={i}
      >
        {record.FactorLevels[l].FactorName}:&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
        {record.FactorLevels[l].SelectedLevel}
      </text>
    );
  });

  // Not used
  // var toggleToolbox = function(){
  //     props.onDiamondClick(props.toolboxCol, props.toolboxGrade, props.toolboxIndex, false)
  //   }

  return (
    <g transform={`translate(${props.toolboxCol * colWidth}, 0)`}>
      <g transform={`translate(${colX},0)`}>
        <g transform={`translate(0,${(props.toolboxGrade + 1) * rowHeight})`}>
          <g transform={`translate(${size / 2}, -${size / 2}) scale(1, -1)`}>
            <rect
              width={rectSize + 2}
              height={rectSize + 2}
              transform={`translate(${x},${y}) rotate(45) translate(-${h},-${h})`}
              style={{stroke: 'black', fill: 'darkgreen'}}
            />
          </g>
          <g transform={`translate(${size / 2}, -${size / 2}) scale(1, -1)`}>
            <rect
              width={toolBoxSizeWidth}
              height={toolBoxSizeHeight}
              transform={`translate(${x},${y}) translate(${h * 3},0)`}
              style={{stroke: 'white', fill: 'lightgrey'}}
            />
            <g
              transform={`translate(${x + (toolBoxSizeWidth + h * 3) / 2}, ${
                y + columnLabelHeight
              }) scale(1, -1)`}
            >
              {factors}
              <text
                x={0}
                y={
                  -columnLabelHeight *
                  (Object.keys(record.FactorLevels).length + 2)
                }
                textAnchor="middle"
                className="axis-title"
                key={Object.keys(record.FactorLevels).length}
              >
                Record #{record.RecordNo} {record.RecordName}
              </text>
              <text
                x={0}
                y={0}
                textAnchor="middle"
                className="axis-title"
                key={Object.keys(record.FactorLevels).length + 1}
              >
                Performance: {record.Performance}
              </text>
            </g>
          </g>
        </g>
      </g>
    </g>
  );
}

function XAxis(props) {
  var colWidth = columnWidth;
  if (props.singleColumn) {
    colWidth = columnWidth * 2;
  }
  return (
    <g className="axis">
      <line
        x1="0"
        y1="0"
        x2={colWidth * props.labels.length}
        y2="0"
        stroke="black"
        strokeWidth="1"
      />
      {props.labels.map((l, i) => (
        <text
          key={i}
          x={(i + 0.5) * colWidth}
          y={columnLabelHeight}
          textAnchor="middle"
          className="axis-label"
        >
          {l}
        </text>
      ))}
      <text
        x={(colWidth * props.labels.length) / 2}
        y={2 * columnLabelHeight}
        textAnchor="middle"
        className="axis-title"
      >
        {props.title}
      </text>
    </g>
  );
}

function YAxis(props) {
  // TODO: Why?
  const x = -columnLabelHeight / 2;
  return (
    <g className="axis">
      <line
        x1="0"
        y1="0"
        x2="0"
        y2={rowHeight * props.labels.length}
        stroke="black"
        strokeWidth="1"
      />
      {props.labels.map((l, i) => (
        <g
          transform={`translate(${x},${(i + 0.5) * rowHeight}) rotate(-90)`}
          key={i}
        >
          <text textAnchor="middle" className="axis-label">
            {l.label}
          </text>
        </g>
      ))}
      <g
        transform={`translate(${3 * x},${
          (rowHeight * props.labels.length) / 2
        }) rotate(-90)`}
      >
        <text textAnchor="middle" className="axis-title">
          {props.title}
        </text>
      </g>
    </g>
  );
}

export class Graph extends React.Component {
  state = {
    mode: 0,
    showToolbox: false,
    toolboxCol: -1,
    toolboxGrade: -1,
    toolboxIndex: -1,
    record: null,
  };

  toggleToolbox(col, grade, index, show) {
    if (
      col == this.state.toolboxCol &&
      grade == this.state.toolboxGrade &&
      index == this.state.toolboxIndex
    ) {
      this.state.showToolbox = false;
    } else {
      this.state.showToolbox = show;
    }
    if (!this.state.showToolbox) {
      this.state.toolboxCol = -1;
      this.state.toolboxGrade = -1;
      this.state.toolboxIndex = -1;
      this.state.record = null;
      this.setState(this.state);
    } else {
      this.state.toolboxCol = col;
      this.state.toolboxGrade = grade;
      this.state.toolboxIndex = index;

      var allRecords = this.props.user.getAllPerformanceRecords();
      var record = allRecords[grade].Records[this.props.colFilters[col]][index];

      this.state.record = record;
      var self = this;
      var onComplete = function () {
        self.setState(self.state);
        self.props.app.refreshDialog();
      };
      this.submitRecordSelect(onComplete);
    }
  }

  submitRecordSelect(onComplete) {
    var user = this.props.user;
    var prompt = user.getPrompt();
    var promptId = prompt.PromptId;
    var phaseId = user.getCurrentPhaseId();

    var response = {};

    response.RecordNo = this.state.record.RecordNo;

    var jsonResponse = JSON.stringify(response);
    user.submitResponse(promptId, phaseId, jsonResponse, onComplete);
  }

  render() {
    var props = this.props;
    //    var singleColumn = props.data.length > 1 ? false : true;
    var singleColumn = props.singleColumn; // letting the parent element have more control
    var drawingAreaH, drawingAreaW;
    var colWidth = columnWidth;
    var showToolbox = this.state.showToolbox;
    if (singleColumn) {
      colWidth = columnWidth * 2;
    }
    var allowToolboxToggle = props.allowToolboxToggle && !showToolbox;
    const labels = props.data.map(v => v.label);
    const columns = props.data.map((v, i) => (
      <g transform={`translate(${i * colWidth}, 0)`} key={i}>
        <Column
          singleColumn={singleColumn}
          onDiamondClick={this.toggleToolbox}
          allowToolboxToggle={allowToolboxToggle}
          rcount={v.rcount}
          col={i}
        />
      </g>
    ));

    const rowBackground = props.data[0].rcount.map((_, i) => {
      return (
        <rect
          width={props.data.length * colWidth}
          height={rowHeight}
          y={i * rowHeight}
          fill={rowColors[i % rowColors.length]}
          key={i}
        />
      );
    });

    var records;
    if (showToolbox) {
      // This is when tool box was toggled on, which triggered a change of state
      var key =
        'k' +
        this.state.toolboxCol +
        ':' +
        this.state.toolboxGrade +
        ':' +
        this.state.toolboxIndex;
      var toolbox = (
        <Toolbox
          user={props.user}
          colFilters={props.colFilters}
          singleColumn={singleColumn}
          toolboxCol={this.state.toolboxCol}
          toolboxGrade={this.state.toolboxGrade}
          toolboxIndex={this.state.toolboxIndex}
          data={props.data}
          record={this.state.record}
          key={key}
        />
      );
      records = [toolbox];
    } else if (props.recordsToShow && props.recordsToShow.length > 0) {
      showToolbox = true;
      var allRecords = this.props.user.getAllPerformanceRecords();
      // This is when the properties of the Graph says to draw showing two records explicitly
      records = props.recordsToShow.map(function (r, i) {
        for (var j = 0; j < allRecords[r.grade].Records[r.filter].length; j++) {
          var record = allRecords[r.grade].Records[r.filter][j];
          if (r.no == record.RecordNo) {
            var col;
            for (var jj = 0; jj < props.colFilters.length; jj++) {
              if (r.filter == props.colFilters[jj]) {
                col = jj;
                break;
              }
            }
            var key = 'k' + r.filter + ':' + r.grade + ':' + j;
            return (
              <Toolbox
                user={props.user}
                colFilters={props.colFilters}
                singleColumn={singleColumn}
                toolboxCol={col}
                toolboxGrade={r.grade}
                toolboxIndex={j}
                data={props.data}
                record={record}
                key={key}
              />
            );
          }
        }
      });
    }

    if (showToolbox) {
      drawingAreaW = '100%';
      drawingAreaH =
        paddingBottom + props.yLabels.length * rowHeight + paddingTop;
    } else {
      drawingAreaW = paddingLeft + props.data.length * colWidth + paddingRight;
      drawingAreaH =
        paddingBottom + props.yLabels.length * rowHeight + paddingTop;
    }

    return (
      <svg
        className="graph"
        style={{
          width: drawingAreaW,
          height: drawingAreaH,
        }}
      >
        <g transform={`translate(${paddingLeft}, ${paddingTop})`}>
          {rowBackground}
          {columns}
          <YAxis labels={props.yLabels} title={props.yTitle} />
          <g
            transform={`translate(0,${
              props.data[0].rcount.length * rowHeight
            })`}
          >
            <XAxis
              labels={labels}
              title={props.xTitle}
              singleColumn={singleColumn}
            />
          </g>
          {records}
        </g>
      </svg>
    );
  }
}

export class ChartSelectTargetFactor extends React.Component {
  state = {enabled: false};

  isEnabled() {
    return this.state.enabled;
  }

  handleChange = event => {
    this.setState({enabled: true});
  };

  handleSubmit(event) {
    event.preventDefault(); // default might be to follow a link, instead, takes control over the event

    var user = this.props.user;
    var onComplete = this.props.onComplete;
    var e = document.getElementById('promptId');
    var promptId = e ? e.value : '';
    var e = document.getElementById('phaseId');
    var phaseId = e ? e.value : '';
    var f = document.getElementById('chartactionForm');
    e = f.elements['chartactioninput'];
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
  }

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
      return <ChartFactorPromptOption option={option} key={i} />;
    });

    return (
      <form
        id="chartactionForm"
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
            key={'ChartSelectTargetFactor'}
          >
            Enter
          </button>
        </p>
      </form>
    );
  }
}

export function ChartFactorPromptOption(props) {
  var option = props.option;
  return (
    <tr>
      <td>
        <label>
          <input
            type="radio"
            name="chartactioninput"
            value={option.ResponseId}
          />
          <br />
          {option.Text}
        </label>
      </td>
    </tr>
  );
}

export function ChartMemoForm(props) {
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
        <Chart
          user={props.user}
          showTargetFactorRecords
          app={props.app}
          key={'CHART_MEMO_FORM'}
        />
      </div>
    </div>
  );
}

export function FactorsSummaryForm(props) {
  var question = 'Select "Yes" for factor that you found makes a difference.';
  var formName = 'chartactionForm';
  return (
    <div>
      <MultiFactorsCausality
        formName={formName}
        question={question}
        user={props.user}
        onComplete={props.onComplete}
        app={props.app}
      />
      <ChartButtons user={props.user} app={props.app} />
    </div>
  );
}

export function FactorsLevelsSummaryForm(props) {
  var question =
    'Choose the level of the factor that you found would be best for performance.';
  var formName = 'chartactionForm';
  return (
    <div>
      <MultiFactorsCausalityLevels
        formName={formName}
        question={question}
        user={props.user}
        onComplete={props.onComplete}
        app={props.app}
      />
      <ChartButtons user={props.user} app={props.app} />
    </div>
  );
}

export class ChartButtons extends React.Component {
  // propTypes: {
  //   fitness: React.PropTypes.object,
  //   parentshealth: React.PropTypes.object,
  //   familysize: React.PropTypes.object,
  //   education: React.PropTypes.object,
  //   homeclimate: React.PropTypes.object,
  // },

  static defaultProps = {
    fitness: {
      filterFactorName: 'Fitness',
      filterLevelsLabels: ['Excellent', 'Average'],
      filterRecords: ['fitness:excellent', 'fitness:average'],
    },
    parentshealth: {
      filterFactorName: "Parents' Health",
      filterLevelsLabels: ['Excellent', 'Fair'],
      filterRecords: ['parentshealth:excellent', 'parentshealth:fair'],
    },
    familysize: {
      filterFactorName: 'Family Size',
      filterLevelsLabels: ['Large', 'Small'],
      filterRecords: ['familysize:large', 'familysize:small'],
    },
    education: {
      filterFactorName: 'Education',
      filterLevelsLabels: ['College', 'Some College', 'No College'],
      filterRecords: [
        'education:college',
        'education:somecollege',
        'education:nocollege',
      ],
    },
    homeclimate: {
      filterFactorName: 'Home Climate',
      filterLevelsLabels: ['Hot', 'Cold'],
      filterRecords: ['homeclimate:hot', 'homeclimate:cold'],
    },
  };

  state = {showChart: null};

  changeState(chart) {
    if (this.state.showChart == chart) {
      // Toggle off if the same chart is selected
      this.state.showChart = null;
    } else {
      // Toggle on the selected chart
      this.state.showChart = chart;
    }
    this.setState(this.state); // This call triggers re-rendering
  }

  render() {
    var self = this;
    var state = this.state;
    var user = this.props.user;
    var app = this.props.app;
    var fitnessOnClick = function () {
      self.changeState(self.props.fitness);
    };
    var parentshealthOnClick = function () {
      self.changeState(self.props.parentshealth);
    };
    var familysizeOnClick = function () {
      self.changeState(self.props.familysize);
    };
    var educationOnClick = function () {
      self.changeState(self.props.education);
    };
    var homeclimateOnClick = function () {
      self.changeState(self.props.homeclimate);
    };
    var hideChartOnClick = function () {
      self.changeState(null);
    };

    if (state.showChart) {
      return (
        <div>
          <div className="no-border-frame">
            <button onClick={fitnessOnClick}>
              {this.props.fitness.filterFactorName}
            </button>
            <button onClick={parentshealthOnClick}>
              {this.props.parentshealth.filterFactorName}
            </button>
            <button onClick={familysizeOnClick}>
              {this.props.familysize.filterFactorName}
            </button>
            <button onClick={educationOnClick}>
              {this.props.education.filterFactorName}
            </button>
            <button onClick={homeclimateOnClick}>
              {this.props.homeclimate.filterFactorName}
            </button>
          </div>
          <label>
            Click on the button for a factor if you want to see the chart for
            the factor.
          </label>
          <div className="hbox">
            <Chart
              user={user}
              filterFactorName={state.showChart.filterFactorName}
              filterLevelsLabels={state.showChart.filterLevelsLabels}
              filterRecords={state.showChart.filterRecords}
              app={app}
              key={state.showChart.filterFactorName}
            />
          </div>
          <button autoFocus onClick={hideChartOnClick}>
            Hide Chart
          </button>
        </div>
      );
    } else {
      return (
        <div>
          <div className="no-border-frame">
            <button autoFocus onClick={fitnessOnClick}>
              {this.props.fitness.filterFactorName}
            </button>
            <button autoFocus onClick={parentshealthOnClick}>
              {this.props.parentshealth.filterFactorName}
            </button>
            <button autoFocus onClick={familysizeOnClick}>
              {this.props.familysize.filterFactorName}
            </button>
            <button autoFocus onClick={educationOnClick}>
              {this.props.education.filterFactorName}
            </button>
            <button autoFocus onClick={homeclimateOnClick}>
              {this.props.homeclimate.filterFactorName}
            </button>
          </div>
          <label>
            Click on the button for a factor if you want to see the chart for
            the factor.
          </label>
        </div>
      );
    }
    return null;
  }
}
