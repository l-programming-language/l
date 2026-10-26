Use ParagraphObserver's Print from `paragraph.l` as "Print".
"firstText" is `First`.
"secondText" is `Second`.
"result" is `done`.
Create "First" as a New Basic App.
The App Prints firstText.
Create "Second" as a New Basic App.
The App Prints secondText.

    Use Second as "Selected".
Run Selected.
Await Selected.
Define "result" as Selected's Output.
Show result.
Run First.
Await First.
Show First's Output.

Show result.
