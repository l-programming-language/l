Use ProbeIO's Print from `../companions/stop_gate.l` as "Print".
"workerText" is `worker`.
"gateText" is `gate`.

    Create "Worker" as a New Public Basic App.
The App Prints workerText.
Show `unawaited-continuation-must-not-run`.

Create "Gate" as a New Basic App.
The App Prints gateText.
Start Worker.
Execute Gate.
Show `done`.
