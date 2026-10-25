Use ParagraphObserver's Print from `paragraph.l` as "Print".
"message" is `Inner`.
"otherMessage" is `Other`.

    Create "Worker" as a New Public Basic App.
The App Prints message.
Run Other.
Await Other.
Define "result" as Worker's Output.
Show result.
Show `finished`.

    Create "Other" as a New Public Basic App.
The App Prints otherMessage.
Show Other's Output.

Run Worker.
Await Worker.
Show `done`.
