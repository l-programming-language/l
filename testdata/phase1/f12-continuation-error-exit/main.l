Use ProbeIO's Print from `../companions/gate.l` as "Print".
"firstText" is `A`.
"secondText" is `B`.

    Create "First" as a New Public Basic App.
The App Prints firstText.
Show First's Output.
If firstText equals firstText, exit with an error of `continuation stopped`.

    Create "Second" as a New Public Basic App.
The App Prints secondText.
Show Second's Output.

Start First.
Start Second.
Show `launched|`.
Await First.
Await Second.
Show `unreachable`.
