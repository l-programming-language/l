Use ProbeIO's Print from `../companions/stop_gate.l` as "Print".
"WorkerText" is `worker`.
Create "Worker" as a New Basic App.
The App Prints WorkerText.
"OtherText" is `other`.
Create "Other" as a New Basic App.
The App Prints OtherText.
"GateText" is `gate`.
Create "Gate" as a New Basic App.
The App Prints GateText.
Start Worker.
Start Other.
Execute Gate.
Stop Worker.
Await Worker.
Show Worker's Output.
If Worker has an Error, Show `stopped|`.
Show `worker-stopped|`.
Await Other.
Show Other's Output.
