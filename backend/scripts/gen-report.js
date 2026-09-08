// Turns every domain suite's cucumber JSON output (reports/json/*.json,
// one per tests/<domain> package) into a single merged HTML report at
// reports/cucumber-report.html. Run via `task test-bdd-report`, which runs
// the suite first so the JSON exists.
import reporter from "cucumber-html-reporter";

reporter.generate({
  theme: "bootstrap",
  jsonDir: "reports/json",
  output: "reports/cucumber-report.html",
  reportSuiteAsScenarios: true,
  launchReport: true,
});
